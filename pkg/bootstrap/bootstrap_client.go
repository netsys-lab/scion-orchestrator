package bootstrap

import (
	"net"

	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
)

// TODO: Replace insecure mode, verify everything is working properly
func BootstrapFromAddress(env *environment.HostEnvironment, config *conf.Config) error {

	bootstrapperConfig := &Config{
		SecurityMode:    "insecure",
		SciondConfigDir: env.ConfigPath,
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Bootstrap.Server)
	if err != nil {
		return err
	}

	return FetchConfiguration(bootstrapperConfig, tcpAddr)
}
