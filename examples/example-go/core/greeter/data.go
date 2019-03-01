package greeter

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/astranet/galaxy/data"
	"github.com/astranet/galaxy/metrics"
)

type DataRepo interface {
	data.Repo

	GetGreetNum(name string) (uint64, error)
	IncGreetNum(name string, v int) (uint64, error)
}

type DataRepoOptions struct {
}

func checkDataRepoOptions(opt *DataRepoOptions) *DataRepoOptions {
	if opt == nil {
		opt = &DataRepoOptions{}
	}
	return opt
}

func NewDataRepo(
	globalRepo data.Repo,
	opt *DataRepoOptions,
) DataRepo {
	return &repo{
		Repo: globalRepo,

		opt: checkDataRepoOptions(opt),
		tags: metrics.Tags{
			"layer":   "data",
			"service": "greeter",
		},
		fields: log.Fields{
			"layer":   "data",
			"service": "greeter",
		},

		counter:    make(map[string]uint64),
		counterMux: new(sync.RWMutex),
	}
}

type repo struct {
	data.Repo

	tags   metrics.Tags
	fields log.Fields
	opt    *DataRepoOptions

	counter    map[string]uint64
	counterMux *sync.RWMutex
}

func (r *repo) GetGreetNum(name string) (uint64, error) {
	metrics.ReportFuncCall(r.tags)
	statsFn := metrics.ReportFuncTiming(r.tags)
	defer statsFn()

	r.counterMux.RLock()
	v := r.counter[name]
	r.counterMux.RUnlock()

	return v, nil
}

func (r *repo) IncGreetNum(name string, v int) (uint64, error) {
	metrics.ReportFuncCall(r.tags)
	statsFn := metrics.ReportFuncTiming(r.tags)
	defer statsFn()

	r.counterMux.Lock()
	r.counter[name] += uint64(v)
	result := r.counter[name]
	r.counterMux.Unlock()

	return result, nil
}
