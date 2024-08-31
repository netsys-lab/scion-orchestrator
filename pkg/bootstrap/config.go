// Package config contains the configuration of bootstrapper.
package bootstrap

import (
	"path/filepath"
)

const (
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
)

type SecurityMode string

const (
	Strict     SecurityMode = "strict"     // only store a TRC if it validates against an existing TRC update chain
	Permissive SecurityMode = "permissive" // only store a TRC if it does not conflict with an existing TRC update chain
	Insecure   SecurityMode = "insecure"   // store any TRC, mark it as insecure, don't validate the topology signature
)

var (
	version       bool
	versionString string
	helpConfig    bool
	configPath    string
	IfaceName     string
)

type Config struct {
	InterfaceName   string       `toml:"iface,omitempty"`
	SciondConfigDir string       `toml:"sciond_config_dir"`
	SecurityMode    SecurityMode `toml:"security_mode,omitempty"`
	Logging         LogConfig    `toml:"log,omitempty"`
	CryptoEngine    string       `toml:"crypto_engine,omitempty"`
}

func (cfg *Config) WorkingDir() string {
	return filepath.Join(cfg.SciondConfigDir, "bootstrapper")
}

// LogConfig is the configuration for the logger.
type LogConfig struct {
	Console ConsoleConfig `toml:"console,omitempty"`
}

// ConsoleConfig is the config for the console logger.
type ConsoleConfig struct {
	// Level of console logging (defaults to DefaultLogLevel).
	Level string `toml:"level,omitempty"`
}
