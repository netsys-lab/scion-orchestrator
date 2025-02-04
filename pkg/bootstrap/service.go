package bootstrap

import (
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

func RunBootstrapService(configDir string, url string, config *conf.Config) error {

	var eg errgroup.Group

	eg.Go(func() error {
		return RunTrcFileWatcher(configDir, config)
	})

	eg.Go(func() error {
		return RunBootstrapServer(configDir, url, config)
	})

	metrics.Status.BootstrapServer.Status = metrics.SERVICE_STATUS_RUNNING

	return eg.Wait()
}
