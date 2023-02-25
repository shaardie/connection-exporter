package dns

import (
	"context"
	"net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shaardie/connection-exporter/pkg/logging"
)

const (
	NetworkLabel  = "network"
	HostLabel     = "host"
	ResolverLabel = "resolver"
)

var (
	metrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connection_exporter_dns_success",
		Help: "Successful dns requests",
	}, []string{NetworkLabel, HostLabel, ResolverLabel})
)

type DNS struct {
	cfg      Config
	metric   prometheus.Gauge
	resolver net.Resolver
}

type Config struct {
	CustomResolver bool   `yaml:"custom_resolver"`
	Host           string `yaml:"host"`
	Network        string `yaml:"network"`
}

func New(cfg Config) *DNS {
	resolverLabel := "system"
	if cfg.CustomResolver {
		resolverLabel = "custom"
	}

	dns := &DNS{
		cfg: cfg,
		metric: metrics.With(prometheus.Labels{
			NetworkLabel:  cfg.Network,
			HostLabel:     cfg.Host,
			ResolverLabel: resolverLabel,
		}),
	}

	if cfg.CustomResolver {
		dns.resolver = net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, _, address string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, cfg.Network, address)
			},
		}
	}
	return dns

}
func (dns *DNS) Do(ctx context.Context) {
	logger := logging.FromContextOrDiscard(ctx)

	_, err := dns.resolver.LookupIP(ctx, "ip", dns.cfg.Host)
	if err != nil {
		dns.metric.Set(0)
		logger.Infow("Lookup failed", "config", dns.cfg, "error", err)
		return
	}

	logger.Debugw("Lookup succeeded", "config", dns.cfg)
	dns.metric.Set(1)
}
