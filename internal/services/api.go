package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"example.com/analytics_api/pkg/config"
	"example.com/analytics_api/pkg/handler"
	"example.com/analytics_api/pkg/registry"
	"example.com/analytics_api/pkg/service"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
	log "github.com/sirupsen/logrus"
)

func WithHandlerFactory(name string, h handler.Constructor) service.ServiceOpt {
	return func(s service.IService) error {
		if i, ok := s.(handler.IFactory); ok {
			return i.Register(name, h)
		}
		return nil
	}
}

type ApiConfig struct {
	Addr string
}

func NewApiConfig() *ApiConfig {
	return &ApiConfig{
		Addr: ":8080",
	}
}

func (c *ApiConfig) Read(cr config.IReader) error {
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
	c         *ApiConfig
	logger    *log.Logger
	app       *fiber.App
	handlersF registry.IRegistry[handler.Constructor]
}

func NewApi(opts ...service.ServiceOpt) (service.IService, error) {
	s := &apiService{
		c:         NewApiConfig(),
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

type logger struct {
	l *log.Logger
}

var _ fiberlog.AllLogger = (*logger)(nil)

func (l *logger) newEntryW(args ...any) *log.Entry {
	e := log.NewEntry(l.l)
	if len(args) > 0 {
		if (len(args) & 1) == 1 {
			args = append(args, "__KEYVALS_UNPAIRED__")
		}

		for i := 0; i < len(args); i += 2 {
			e = e.WithField(fmt.Sprintf("%s", args[i]), args[i+1])
		}
	}
	return e
}

func (l *logger) Debug(args ...any) {
	l.l.Debug(args...)
}

func (l *logger) Debugf(format string, args ...any) {
	l.l.Debugf(format, args...)
}

func (l *logger) Debugw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Debug(format)
}

func (l *logger) Error(args ...any) {
	l.l.Error(args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.l.Errorf(format, args...)
}

func (l *logger) Errorw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Error(format)
}

func (l *logger) Fatal(args ...any) {
	l.l.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.l.Fatalf(format, args...)
}

func (l *logger) Fatalw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Fatal(format)
}

func (l *logger) Info(args ...any) {
	l.l.Info(args...)
}

func (l *logger) Infof(format string, args ...any) {
	l.l.Infof(format, args...)
}

func (l *logger) Infow(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Info(format)
}

func (l *logger) Panic(args ...any) {
	l.l.Panic(args...)
}

func (l *logger) Panicf(format string, args ...any) {
	l.l.Panicf(format, args...)
}

func (l *logger) Panicw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Panic(format)
}

func (l *logger) Trace(args ...any) {
	l.l.Trace(args...)
}

func (l *logger) Tracef(format string, args ...any) {
	l.l.Tracef(format, args...)
}

func (l *logger) Tracew(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Trace(format)
}

func (l *logger) Warn(args ...any) {
	l.l.Warn(args...)
}

func (l *logger) Warnf(format string, args ...any) {
	l.l.Warnf(format, args...)
}

func (l *logger) Warnw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Warn(format)
}

func (l *logger) SetLevel(lvl fiberlog.Level) {
	levels := slices.Clone(log.AllLevels)
	slices.Reverse(levels)
	l.l.SetLevel(levels[lvl])
}

func (l *logger) SetOutput(out io.Writer) {
	l.l.SetOutput(out)
}

func (l *logger) WithContext(ctx context.Context) fiberlog.CommonLogger {
	return l
}
