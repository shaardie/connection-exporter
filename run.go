package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/shaardie/is-connected/pkg/config"
	"github.com/shaardie/is-connected/pkg/dns"
	iscHTTP "github.com/shaardie/is-connected/pkg/http"
	"github.com/shaardie/is-connected/pkg/tcp"
	"github.com/spf13/pflag"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Doer interface {
	Do(context.Context)
}

func run() {

	cfgFilename := pflag.CommandLine.String("config", "", "Configration file")
	pflag.Parse()

	cfg, err := config.New(*cfgFilename)
	if err != nil {
		log.Fatalf("Failed to configure, %v", err)
	}

	// Log
	var zapLog *zap.Logger
	if cfg.Logging.Level == "debug" {
		zapLog, err = zap.NewDevelopment()
	} else {
		zapLog, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to create zap logger, %v", err)
	}
	log := zapr.NewLogger(zapLog)

	log.Info("Config", "Config", cfg)

	// Context
	ctx := context.Background()
	ctx = logr.NewContext(ctx, log)

	doers := []Doer{}

	// TCP
	for _, tcpConfig := range cfg.Tests.TCP.Config {
		doers = append(doers, tcp.New(ctx, tcpConfig))
	}

	// DNS
	for _, dnsConfig := range cfg.Tests.DNS.Config {
		doers = append(doers, dns.New(ctx, dnsConfig))
	}

	// HTTP
	for _, httpConfig := range cfg.Tests.HTTP.Config {
		doers = append(doers, iscHTTP.New(ctx, httpConfig))
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.Server.Address, nil)
	}()

	for {
		for _, doer := range doers {
			go doer.Do(ctx)
		}
		time.Sleep(cfg.Server.Interval)
	}
}
