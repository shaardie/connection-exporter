package config

import (
	"fmt"
	"time"

	"github.com/shaardie/is-connected/pkg/dns"
	"github.com/shaardie/is-connected/pkg/http"
	"github.com/shaardie/is-connected/pkg/tcp"
	"github.com/spf13/viper"
)

type Config struct {
	Logging logging
	Server  server
	Tests   tests
}

type logging struct {
	Level      string
	Structured bool
}

type server struct {
	Address  string
	Interval time.Duration
}

type tests struct {
	TCP struct {
		Enabled bool
		Config  []tcp.Config
	}

	DNS struct {
		Enabled bool
		Config  []dns.Config
	}

	HTTP struct {
		Enabled bool
		Config  []http.Config
	}
}

func New(filename string) (*Config, error) {
	viper.SetDefault("server", server{
		Address:  "127.0.0.1:8144",
		Interval: 15 * time.Second,
	})

	viper.SetDefault("logging", logging{
		Level:      "info",
		Structured: true,
	})

	viper.SetDefault("tests", tests{
		TCP: struct {
			Enabled bool
			Config  []tcp.Config
		}{
			Enabled: true,
			Config: []tcp.Config{
				{
					Host:    "example.com",
					Port:    80,
					Network: "tcp4",
				},
				{
					Host:    "example.com",
					Port:    80,
					Network: "tcp6",
				},
			},
		},
	},
	)
	if filename != "" {
		viper.SetConfigFile(filename)
		viper.SetConfigType("yaml")

		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config, %w", err)
		}

	}

	cfg := &Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config, %w", err)
	}

	return cfg, nil
}
