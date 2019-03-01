package greeter

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/astranet/galaxy/metrics"
)

//go:generate galaxy -D ../.. expose -y -P greeter

type Service interface {
	Greet(name string) (msg string, err error)
	GreetCount(name string) (count uint64, err error)
}

type ServiceOptions struct {
}

func checkServiceOptions(opt *ServiceOptions) *ServiceOptions {
	if opt == nil {
		opt = &ServiceOptions{}
	}
	return opt
}

func NewService(
	repo DataRepo,
	opt *ServiceOptions,
) Service {
	return &service{
		opt: checkServiceOptions(opt),
		tags: metrics.Tags{
			"layer":   "service",
			"service": "greeter",
		},
		fields: log.Fields{
			"layer":   "service",
			"service": "greeter",
		},

		repo: repo,
	}
}

type service struct {
	repo   DataRepo
	tags   metrics.Tags
	fields log.Fields
	opt    *ServiceOptions
}

func (s *service) Greet(name string) (string, error) {
	metrics.ReportFuncCall(s.tags)
	statsFn := metrics.ReportFuncTiming(s.tags)
	defer statsFn()

	count, err := s.repo.IncGreetNum(name, 1)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Hello %s! I've seen you %d times.", name, count)

	return msg, nil
}

func (s *service) GreetCount(name string) (uint64, error) {
	metrics.ReportFuncCall(s.tags)
	statsFn := metrics.ReportFuncTiming(s.tags)
	defer statsFn()

	count, err := s.repo.GetGreetNum(name)
	if err != nil {
		return 0, err
	}

	return count, nil
}
