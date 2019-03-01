package main

import (
	"os"
	"time"

	"github.com/astranet/galaxy/cluster"
	"github.com/astranet/galaxy/logging"
	"github.com/astranet/galaxy/metrics"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	"github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
	"github.com/xlab/closer"

	"github.com/astranet/example_api/core/greeter"
)

var app = cli.App("core_greeter", "Greeter Service server from example_api.")

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
	if err := publishService(c); err != nil {
		closer.Fatalln(err)
	}
	if err := c.ListenAndServe(*netAddr); err != nil {
		closer.Fatalln(err)
	}
	go func() {
		if *appLogLevel == "debug" {
			if err := c.ListenAndServeHTTP(*httpAddr); err != nil {
				closer.Fatalln(err)
			}
		}
	}()
	closer.Hold()
}

func publishService(c cluster.Cluster) error {
	greeterRepo := greeter.NewDataRepo(nil, nil)
	greeterService := greeter.NewService(greeterRepo, nil)
	greeterHandler := greeter.NewHandler(greeterService, nil)
	return c.Publish(greeterHandler)
}
