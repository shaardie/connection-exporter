package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/shaardie/connection-exporter/pkg/config"
	"github.com/shaardie/connection-exporter/pkg/dns"
	ceHTTP "github.com/shaardie/connection-exporter/pkg/http"
	"github.com/shaardie/connection-exporter/pkg/logging"
	"github.com/shaardie/connection-exporter/pkg/tcp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Tests interface {
	Do(context.Context)
}

func main() {
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

	logger.Debugw("Initializing", "config", cfg)

	tests := []Tests{}

	// TCP
	if cfg.Tests.TCP.Enabled {
		logger.Infow("TCP Tests enabled")
		for _, config := range cfg.Tests.TCP.Config {
			logger.Infow("Initialize TCP Test", "config", config)
			tests = append(tests, tcp.New(config))
		}
	}

	// DNS
	if cfg.Tests.DNS.Enabled {
		logger.Infow("DNS Tests enabled")
		for _, config := range cfg.Tests.DNS.Config {
			logger.Infow("Initialize DNS Test", "config", config)
			tests = append(tests, dns.New(config))
		}
	}

	// HTTP
	if cfg.Tests.HTTP.Enabled {
		logger.Infow("HTTP Tests enabled")
		for _, config := range cfg.Tests.HTTP.Config {
			logger.Infow("Initialize HTTP Test", "config", config)
			tests = append(tests, ceHTTP.New(config))
		}
	}

	if len(tests) == 0 {
		logger.Info("No tests specified")
	}

	logger.Infow("Start http server for metrics and health", "server", cfg.Server)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Health check")
		w.Write([]byte("OK"))
	})
	go http.ListenAndServe(cfg.Server.Address, nil)

	for {
		logger.Info("Loop through tests")
		for _, test := range tests {
			go test.Do(ctx)
		}
		time.Sleep(cfg.Server.Interval)
	}
}
