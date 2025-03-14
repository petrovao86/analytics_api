package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"example.com/analytics_api/internal/api"
	"example.com/analytics_api/internal/events"
	"example.com/analytics_api/pkg/config"
	"example.com/analytics_api/pkg/service"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultConfigPath = "../bin/app.yaml"
	DefaultAddr       = ":8080"
)

var (
	configPath        string
	logLevel          string
	svcs              servicesSlice
	supportedServices = []string{"api", "generator"}
)

type servicesSlice []string

func (s *servicesSlice) String() string {
	return strings.Join(*s, ",")
}
func (s *servicesSlice) Set(val string) error {
	if slices.Index(supportedServices, val) < 0 {
		return fmt.Errorf("unknown service %v", val)
	}
	*s = append(*s, val)
	return nil
}

func init() {
	flag.StringVar(
		&configPath,
		"config",
		DefaultConfigPath,
		"path to configuration file; supported formats are JSON, YAML, INI, ENV and TOML",
	)
	flag.StringVar(
		&logLevel,
		"log-level",
		log.InfoLevel.String(),
		"logging level",
	)
	flag.Var(
		&svcs,
		"service",
		"list of allowed steps",
	)
}

func main() {
	flag.Parse()

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors: true,
		},
	)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cr := config.NewKoanfReader(configPath)

	if svcs == nil {
		svcs = supportedServices
	}

	run := make([]service.IService, 0)
	for _, svc := range svcs {
		switch svc {
		case "api":
			s, err := api.NewService(
				service.WithLogger(log.StandardLogger()),
				api.WithHandlerFactory("events", events.NewHandler),
				config.WithReader[service.IService](cr),
			)
			if err != nil {
				log.Fatal(err)
			}
			run = append(run, s)
		case "generator":
			go generator(ctx, cr)
		}
	}

	for _, svc := range run {
		go svc.Run()
	}
	<-ctx.Done()
	for _, svc := range run {
		ctx, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		svc.Stop(ctx)
	}
}
