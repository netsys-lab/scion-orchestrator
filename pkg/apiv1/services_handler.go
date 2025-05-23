package apiv1

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
	"github.com/netsys-lab/scion-orchestrator/pkg/osutils"
)

type ApiServiceState struct {
	Status string `json:"status"`
	Logs   string `json:"logs"`
}

func GetServiceDetailsHandler(eng *gin.RouterGroup) {
	eng.GET("/services/:serviceId", func(c *gin.Context) {
		serviceId := c.Param("serviceId")
		var logs string
		var err error

		if serviceId == "scion-daemon" {
			if metrics.Status.ServiceMode == "service" {
				logs, err = osutils.GetJournalLogs("scion-daemon.service", 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				logs, err = osutils.GetFileLogs(filepath.Join(environment.HostEnv.LogPath, "sciond.log"), 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		if serviceId == "scion-orchestrator" {
			if metrics.Status.ServiceMode == "service" {
				logs, err = osutils.GetJournalLogs("scion-orchestrator.service", 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				logs, err = osutils.GetFileLogs(filepath.Join(environment.HostEnv.LogPath, "sciond.log"), 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		if serviceId == "scion-dispatcher" {
			if metrics.Status.ServiceMode == "service" {
				logs, err = osutils.GetJournalLogs("scion-dispatcher.service", 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				logs, err = osutils.GetFileLogs(filepath.Join(environment.HostEnv.LogPath, "dispatcher.log"), 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		if strings.Contains(serviceId, "scion-control-service-") {
			if metrics.Status.ServiceMode == "service" {
				logs, err = osutils.GetJournalLogs(serviceId+".service", 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				csLogFile := serviceId + ".log"
				logs, err = osutils.GetFileLogs(filepath.Join(environment.HostEnv.LogPath, csLogFile), 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		if strings.Contains(serviceId, "scion-border-router-") {
			if metrics.Status.ServiceMode == "service" {
				logs, err = osutils.GetJournalLogs(serviceId+".service", 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				brLogFile := serviceId + ".log"
				logs, err = osutils.GetFileLogs(filepath.Join(environment.HostEnv.LogPath, brLogFile), 100)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		c.JSON(http.StatusOK, ApiServiceState{Logs: logs, Status: metrics.Status.Daemon.Status})

	})
}
