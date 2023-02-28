package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shaardie/connection-exporter/pkg/logging"
)

const (
	networkLabel  = "network"
	urlLabel      = "url"
	redirectLabel = "redirect"
)

var (
	metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connection_exporter_http_success",
		Help: "Successful http request",
	}, []string{networkLabel, urlLabel, redirectLabel})
)

type HTTP struct {
	cfg    Config
	client http.Client
	metric prometheus.Gauge
}

type Config struct {
	URL      string
	Redirect bool
	Network  string
}

func New(cfg Config) *HTTP {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	checkRedirect := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	if cfg.Redirect {
		checkRedirect = nil
	}

	return &HTTP{
		cfg: cfg,
		client: http.Client{
			CheckRedirect: checkRedirect,
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
				networkLabel:  cfg.Network,
				urlLabel:      cfg.URL,
				redirectLabel: fmt.Sprintf("%v", cfg.Redirect),
			}),
	}
}

func (http *HTTP) Do(ctx context.Context) {
	logger := logging.FromContextOrDiscard(ctx)

	resp, err := http.client.Get(http.cfg.URL)
	if err != nil {
		http.metric.Set(0)
		logger.Infow("Request failed", "config", http.cfg, "error", err)
		return
	}
	resp.Body.Close()
	logger.Infow("Request succeeded", "config", http.cfg)
	http.metric.Set(1)
}
