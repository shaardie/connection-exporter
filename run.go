package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/shaardie/is-connected/pkg/config"
	"github.com/shaardie/is-connected/pkg/dns"
	iscHTTP "github.com/shaardie/is-connected/pkg/http"
	"github.com/shaardie/is-connected/pkg/logging"
	"github.com/shaardie/is-connected/pkg/tcp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Tests interface {
	Do(context.Context)
}

func run() {
	cfgFilename := flag.String("config", "", "Configration file")
	flag.Parse()

	cfg, err := config.New(*cfgFilename)
	if err != nil {
		log.Fatalf("Failed to configure, %v", err)
	}

	logger, ctx, err := logging.New(context.Background(), cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to create logger, %v", err)
	}

	tests := []Tests{}

	// TCP
	if cfg.Tests.TCP.Enabled {
		logger.V(1).Info("TCP Tests enabled")
		for _, config := range cfg.Tests.TCP.Config {
			logger.V(1).Info("Initialize TCP Test", "config", config)
			tests = append(tests, tcp.New(config))
		}
	}

	// DNS
	if cfg.Tests.DNS.Enabled {
		logger.V(1).Info("DNS Tests enabled")
		for _, config := range cfg.Tests.DNS.Config {
			logger.V(1).Info("Initialize DNS Test", "config", config)
			tests = append(tests, dns.New(config))
		}
	}

	// HTTP
	if cfg.Tests.HTTP.Enabled {
		logger.V(1).Info("HTTP Tests enabled")
		for _, config := range cfg.Tests.HTTP.Config {
			logger.V(1).Info("Initialize HTTP Test", "config", config)
			tests = append(tests, iscHTTP.New(config))
		}
	}

	go func() {
		logger.V(1).Info("start metrics server", "server", cfg.Server)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.Server.Address, nil)
	}()

	logger.Info("Initialized")

	for {
		logger.Info("Loop through tests")
		for _, test := range tests {
			go test.Do(ctx)
		}
		time.Sleep(cfg.Server.Interval)
	}
}
