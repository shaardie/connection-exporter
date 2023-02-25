package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/shaardie/is-connected/pkg/dns"
	"github.com/shaardie/is-connected/pkg/http"
	"github.com/shaardie/is-connected/pkg/tcp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Address  string        `yaml:"address"`
	Debug    bool          `yaml:"debug"`
	Interval time.Duration `yaml:"interval"`
	AsServer bool          `yaml:"as_server"`

	TCP  []tcp.Config  `yaml:"tcp"`
	DNS  []dns.Config  `yaml:"dns"`
	HTTP []http.Config `yaml:"http"`
}

func New(filename string) (*Config, error) {
	cfg := &Config{}

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file %s, %w", filename, err)
	}

	err = yaml.Unmarshal(fileContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file %s, %w", filename, err)
	}

	return cfg, nil
}
