package api

import (
	"context"
	"errors"
	"fmt"

	"example.com/analytics_api/pkg/config"
	"example.com/analytics_api/pkg/handler"
	"example.com/analytics_api/pkg/registry"
	"example.com/analytics_api/pkg/service"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
	log "github.com/sirupsen/logrus"
)

func WithHandlerFactory(name string, h handler.Constructor) service.Opt {
	return func(s service.IService) error {
		if i, ok := s.(handler.IFactory); ok {
			return i.Register(name, h)
		}
		return nil
	}
}

type Config struct {
	Addr string
}

func NewConfig() *Config {
	return &Config{
		Addr: ":8080",
	}
}

func (c *Config) Read(cr config.IReader) error {
	var err error
	newC := *c

	addr, addrErr := config.Get[string](cr, "addr")
	if addrErr != nil && !errors.Is(addrErr, config.ErrNotFound) {
		err = errors.Join(err, addrErr)
	} else if addrErr == nil {
		newC.Addr = addr
	}

	if err != nil {
		return err
	}

	*c = newC
	return nil

}

type apiService struct {
	c         *Config
	logger    *log.Logger
	app       *fiber.App
	handlersF registry.IRegistry[handler.Constructor]
}

func NewService(opts ...service.Opt) (service.IService, error) {
	s := &apiService{
		c:         NewConfig(),
		logger:    log.StandardLogger(),
		app:       fiber.New(),
		handlersF: registry.NewRegistry[handler.Constructor](),
	}
	prometheus := fiberprometheus.New("analytics")
	prometheus.RegisterAt(s.app, "/metrics")
	s.app.Use(prometheus.Middleware)
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	fiberlog.SetLogger(&logger{l: s.logger})
	return s, nil
}

var _ service.ILogWriter = (*apiService)(nil)

func (c *apiService) SetLogger(l *log.Logger) error {
	c.logger = l
	return nil
}

var _ handler.IFactory = (*apiService)(nil)

func (s *apiService) Register(name string, c handler.Constructor) error {
	_, err := s.handlersF.Register(name, c)
	return err
}

func (s *apiService) Build(name string, opts ...handler.Opt) (handler.IHandler, error) {
	c, err := s.handlersF.Get(name)
	if err != nil {
		return nil, err
	}
	return c(opts...)
}

var _ config.IConfigurable = (*apiService)(nil)

func (s *apiService) Configure(cr config.IReader) error {
	apiCr, found := cr.Sub("api")
	var err error
	if !found {
		return fmt.Errorf("api config not found")
	}
	err = errors.Join(err, s.c.Read(apiCr))

	handlers := make([]handler.IHandler, 0)
	handlersCr, handlersErr := config.Sub(apiCr, "handlers")
	err = errors.Join(err, handlersErr)
	if handlersErr == nil {
		for name := range handlersCr.Map() {
			handlerCr, hadlerErr := config.Sub(handlersCr, name)
			err = errors.Join(err, hadlerErr)
			if err == nil {
				handler, buildErr := s.Build(name, config.WithReader[handler.IHandler](handlerCr))
				err = errors.Join(err, buildErr)
				if buildErr == nil {
					handlers = append(handlers, handler)
				}
			}

		}
	}
	if err != nil {
		return err
	}

	for _, handler := range handlers {
		gr := s.app.Group(handler.Path())
		handler.AddRoutes(gr)
	}
	return nil
}

func (s *apiService) Run() error {
	return s.app.Listen(s.c.Addr)
}

func (s *apiService) Stop(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
