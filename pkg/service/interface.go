package service

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type IService interface {
	Run() error
	Stop(context.Context) error
}

type ServiceOpt func(IService) error

type ILogWriter interface {
	SetLogger(*log.Logger) error
}

func WithLogger(l *log.Logger) ServiceOpt {
	return func(s IService) error {
		if i, ok := s.(ILogWriter); ok {
			return i.SetLogger(l)
		}
		return nil
	}
}
