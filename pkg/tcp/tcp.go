package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shaardie/connection-exporter/pkg/logging"
	"golang.org/x/sys/unix"
)

const (
	NetworkLabel = "network"
	HostLabel    = "host"
	PortLabel    = "port"
)

var (
	successMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connection_exporter_tcp_success",
		Help: "Successful tcp request",
	}, []string{NetworkLabel, HostLabel, PortLabel})
	rttMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connection_exporter_tcp_rtt",
		Help: "TCP Round Trip Time in seconds",
	}, []string{NetworkLabel, HostLabel, PortLabel})
)

type TCP struct {
	cfg           Config
	dialer        net.Dialer
	successMetric prometheus.Gauge
	rttMetric     prometheus.Gauge
}

type Config struct {
	Host    string `yaml:"host"`
	Port    uint   `yaml:"port"`
	Network string `yaml:"network"`
}

func New(cfg Config) *TCP {
	return &TCP{
		cfg:    cfg,
		dialer: net.Dialer{},
		successMetric: successMetric.With(
			prometheus.Labels{
				NetworkLabel: cfg.Network,
				HostLabel:    cfg.Host,
				PortLabel:    fmt.Sprintf("%v", cfg.Port),
			}),
		rttMetric: rttMetric.With(
			prometheus.Labels{
				NetworkLabel: cfg.Network,
				HostLabel:    cfg.Host,
				PortLabel:    fmt.Sprintf("%v", cfg.Port),
			}),
	}
}

func (tcp *TCP) Do(ctx context.Context) {
	logger := logging.FromContextOrDiscard(ctx)

	conn, err := tcp.dialer.DialContext(ctx, tcp.cfg.Network, fmt.Sprintf("%v:%v", tcp.cfg.Host, tcp.cfg.Port))
	if err != nil {
		tcp.successMetric.Set(0)
		logger.Infow("Dialing failed", "config", tcp.cfg, "error", err)
		return
	}
	defer conn.Close()

	logger.Debugw("Dialing succeeded", "config", tcp.cfg)
	tcp.successMetric.Set(1)

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		logger.Errorw("connection is not tcp", "config", tcp.cfg, "error", err)
		return
	}

	raw, err := tcpConn.SyscallConn()
	if err != nil {
		logger.Errorw("Failed to get the raw tcp connection", "config", tcp.cfg, "error", err)
		return
	}

	info := new(unix.TCPInfo)
	crtlErr := raw.Control(
		func(fd uintptr) {
			info, err = unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
		},
	)
	if crtlErr != nil {
		logger.Errorw("Failed to invoce function on raw tcp connection", "config", tcp.cfg, "error", err)
		return
	}

	if err != nil {
		logger.Errorw("Failed to get call getsockopt on raw tcp connection", "config", tcp.cfg, "error", err)
		return
	}

	logger.Debugw("TCP Info successfully received", "tpc_info", info, "config", tcp.cfg)
	tcp.rttMetric.Set((time.Duration(info.Rtt) * time.Microsecond).Seconds())
}
