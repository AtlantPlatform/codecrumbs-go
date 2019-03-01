package greeter

import (
	"net/http"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/astranet/galaxy/logging"
	"github.com/astranet/galaxy/metrics"
)

type Handler interface {
	Greet(c *gin.Context)
	GreetCount(c *gin.Context)
}

var HandlerSpec Handler = &handler{}

type HandlerOptions struct {
}

func checkHandlerOptions(opt *HandlerOptions) *HandlerOptions {
	if opt == nil {
		opt = &HandlerOptions{}
	}
	return opt
}

func NewHandler(
	svc Service,
	opt *HandlerOptions,
) Handler {
	return &handler{
		opt: checkHandlerOptions(opt),
		tags: metrics.Tags{
			"layer":   "handler",
			"service": "greeter",
		},
		fields: log.Fields{
			"layer":   "handler",
			"service": "greeter",
		},

		svc: svc,
	}
}

type handler struct {
	svc    Service
	tags   metrics.Tags
	fields log.Fields
	opt    *HandlerOptions
}

func (h *handler) Greet(c *gin.Context) {
	metrics.ReportFuncCall(h.tags)
	statsFn := metrics.ReportFuncTiming(h.tags)
	defer statsFn()

	msg, err := h.svc.Greet(c.GetString("name"))
	if err != nil {
		bugsnag.Notify(err)
		c.String(http.StatusInternalServerError, "error: %v", err)
		return
	}

	c.String(200, "%s", msg)
}

func (h *handler) GreetCount(c *gin.Context) {
	metrics.ReportFuncCall(h.tags)
	statsFn := metrics.ReportFuncTiming(h.tags)
	fnLog := log.WithFields(logging.WithFn(h.fields))
	defer statsFn()

	count, err := h.svc.GreetCount(c.GetString("name"))
	if err != nil {
		fnLog.Warningln(err)
		bugsnag.Notify(err)
		c.String(http.StatusInternalServerError, "error: %v", err)
		return
	}

	c.String(200, "%d", count)
}
