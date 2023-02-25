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
	"github.com/shaardie/is-connected/pkg/tcp"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Doer interface {
	Do(context.Context)
}

func main() {
	cfgFilename := flag.String("config", "./config.yaml", "Configuration File")
	flag.Parse()

	cfg, err := config.New(*cfgFilename)
	if err != nil {
		log.Fatalf("Failed to read config, %v", err)
	}

	// Log
	var zapLog *zap.Logger
	if cfg.Debug {
		zapLog, err = zap.NewDevelopment()
	} else {
		zapLog, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to create zap logger, %v", err)
	}
	log := zapr.NewLogger(zapLog)

	// Context
	ctx := context.Background()
	ctx = logr.NewContext(ctx, log)

	doers := []Doer{}

	// TCP
	for _, tcpConfig := range cfg.TCP {
		doers = append(doers, tcp.New(ctx, tcpConfig))
	}

	// DNS
	for _, dnsConfig := range cfg.DNS {
		doers = append(doers, dns.New(ctx, dnsConfig))
	}

	// HTTP
	for _, httpConfig := range cfg.HTTP {
		doers = append(doers, iscHTTP.New(ctx, httpConfig))
	}

	if cfg.AsServer {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(cfg.Address, nil)
		}()

		for {
			for _, doer := range doers {
				go doer.Do(ctx)
			}
			time.Sleep(cfg.Interval)
		}
	} else {
		for _, doer := range doers {
			doer.Do(ctx)
		}
	}
}
