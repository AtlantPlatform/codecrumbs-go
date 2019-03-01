package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/astranet/galaxy/cluster"
	"github.com/astranet/galaxy/logging"
	"github.com/astranet/galaxy/metrics"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
	"github.com/xlab/closer"

	"github.com/astranet/example_api/core/greeter"
)

var app = cli.App("example_api", "Example API main webserver app.")

func main() {
	app.Before = prepareApp
	app.Action = runApp
	app.Run(os.Args)
}

func prepareApp() {
	log.SetLevel(logging.Level(*appLogLevel))
	if *envName == "local" || *appLogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func runApp() {
	defer closer.Close()

	if toBool(*statsdDisabled) {
		// initializes statsd client with a mock one with no-op enabled
		metrics.Disable()
	} else {
		go func() {
			for {
				err := metrics.Init(*statsdAddr, checkStatsdPrefix(*statsdPrefix), &metrics.StatterConfig{
					EnvName:              *envName,
					StuckFunctionTimeout: duration(*statsdStuckDur, 5*time.Minute),
					MockingEnabled:       toBool(*statsdMocking) || *envName == "local",
				})
				if err != nil {
					bugsnag.Notify(err)
					time.Sleep(time.Minute)
					continue
				}
				break
			}
			closer.Bind(func() {
				metrics.Close()
			})
			if *envName != "local" {
				metrics.RunMemstatsd(*envName, 30*time.Second)
			}
		}()
	}
	c := cluster.NewAstraCluster(*appName, &cluster.AstraOptions{
		Tags:  []string{*clusterName, *envName},
		Nodes: *clusterNodes,
		Debug: *appLogLevel == "debug",
	})
	go func() {
		if err := c.ListenAndServe(*netAddr); err != nil {
			closer.Fatalln(err)
		}
	}()
	setRoutes(c)
	closer.Hold()
}

func setRoutes(c cluster.Cluster) {
	wg := new(sync.WaitGroup)
	r := gin.Default()
	start := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		wait(c, greeter.HandlerSpec)
		greeterCli := c.NewClient(greeter.HandlerSpec)
		r.POST("/api/v1/greeter/greet/:name", gin.WrapH(greeterCli.Use("Greet")))
		r.GET("/api/v1/greeter/greetCount/:name", gin.WrapH(greeterCli.Use("GreetCount")))
	}()

	wg.Wait()
	log.Println("example_api: service backend discovery ended in", time.Since(start))
	go func() {
		log.Println("example_api: starting HTTP server")
		if err := r.Run(*httpListenHost + ":" + *httpListenPort); err != nil {
			closer.Fatalln(err)
		}
	}()
}

func wait(c cluster.Cluster, specs ...cluster.HandlerSpec) {
	ctx, cancelFn := context.WithTimeout(context.Background(), duration(*svcWaitTimeout, time.Minute))
	closer.Bind(cancelFn)
	if err := c.Wait(ctx, specs...); err != nil {
		log.Warnln(err) // non fatal
	}
}
