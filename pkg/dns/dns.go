package dns

import (
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	NetworkLabel  = "network"
	HostLabel     = "host"
	ResolverLabel = "resolver"
)

var (
	metrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dns_success",
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

func New(ctx context.Context, cfg Config) *DNS {
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info("Initialize DNS Test", "config", cfg)

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
	logger := logr.FromContextOrDiscard(ctx)

	_, err := dns.resolver.LookupIP(ctx, "ip", dns.cfg.Host)
	if err != nil {
		dns.metric.Set(0)
		logger.V(1).Info("Lookup failed", "config", dns.cfg, "error", err)
		return
	}

	logger.V(1).Info("Lookup succeeded", "config", dns.cfg)
	dns.metric.Set(1)
}