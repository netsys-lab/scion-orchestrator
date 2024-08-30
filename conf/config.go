package conf

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Command   string
	Bootstrap Bootstrap
	Metrics   Metrics
	IsdAs     string `toml:"isd_as,omitempty"`
	Ca        CA
}

type CA struct {
	CertValidityHours int
	Clients           []string
	Server            string
}

type Bootstrap struct {
	Server string
}

type Metrics struct {
	Server string
}

func NewConfig() *Config {
	return &Config{
		Command: "",
		Bootstrap: Bootstrap{
			Server: "",
		},
		Metrics: Metrics{
			Server: "",
		},
		Ca: CA{
			CertValidityHours: 72,
			Server:            ":3000",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	c := NewConfig()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	log.Println(c)

	return c, nil
}
