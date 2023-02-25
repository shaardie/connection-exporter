package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shaardie/connection-exporter/pkg/logging"
)

const (
	NetworkLabel = "network"
	URLLabel     = "url"
)

var (
	metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connection_exporter_http_success",
		Help: "Successful http request",
	}, []string{NetworkLabel, URLLabel})
)

type HTTP struct {
	cfg    Config
	client http.Client
	metric prometheus.Gauge
}

type Config struct {
	URL     string
	Network string
}

func New(cfg Config) *HTTP {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &HTTP{
		cfg: cfg,
		client: http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
					return dialer.DialContext(ctx, cfg.Network, addr)
				},
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		metric: metric.With(
			prometheus.Labels{
				NetworkLabel: cfg.Network,
				URLLabel:     cfg.URL,
			}),
	}
}

func (http *HTTP) Do(ctx context.Context) {
	logger := logging.FromContextOrDiscard(ctx)

	resp, err := http.client.Get(http.cfg.URL)
	if err != nil {
		http.metric.Set(0)
		logger.Infow("HTTP request failed", "config", http.cfg, "error", err)
		return
	}
	resp.Body.Close()
	logger.Debugw("HTTP request succeeded", "config", http.cfg)
	http.metric.Set(1)
}
