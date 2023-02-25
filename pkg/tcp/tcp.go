package tcp

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	NetworkLabel = "network"
	HostLabel    = "host"
	PortLabel    = "port"
)

var (
	metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tcp_success",
		Help: "Successful http request",
	}, []string{NetworkLabel, HostLabel, PortLabel})
)

type TCP struct {
	cfg    Config
	dialer net.Dialer
	metric prometheus.Gauge
}

type Config struct {
	Host    string `yaml:"host"`
	Port    uint   `yaml:"port"`
	Network string `yaml:"network"`
}

func New(ctx context.Context, cfg Config) *TCP {
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info("Initialize TCP Test", "config", cfg)
	return &TCP{
		cfg:    cfg,
		dialer: net.Dialer{},
		metric: metric.With(
			prometheus.Labels{
				NetworkLabel: cfg.Network,
				HostLabel:    cfg.Host,
				PortLabel:    fmt.Sprintf("%v", cfg.Port),
			}),
	}
}

func (tcp *TCP) Do(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)

	conn, err := tcp.dialer.DialContext(ctx, tcp.cfg.Network, fmt.Sprintf("%v:%v", tcp.cfg.Host, tcp.cfg.Port))
	if err != nil {
		tcp.metric.Set(0)
		logger.V(1).Info("Dialing failed", "config", tcp.cfg, "error", err)
		return
	}
	conn.Close()
	logger.V(1).Info("Dialing succeeded", "config", tcp.cfg)
	tcp.metric.Set(1)
}
