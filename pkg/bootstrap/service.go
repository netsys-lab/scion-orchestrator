package bootstrap

import (
	"github.com/netsys-lab/scion-as/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

func RunBootstrapService(configDir string, url string) error {

	var eg errgroup.Group

	eg.Go(func() error {
		return RunTrcFileWatcher(configDir)
	})

	eg.Go(func() error {
		return RunBootstrapServer(configDir, url)
	})

	metrics.ASStatus.BootstrapServer.Status = metrics.SERVICE_STATUS_RUNNING

	return eg.Wait()
}
