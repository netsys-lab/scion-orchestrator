package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceStatusGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_status",
			Help: "Current status of various services (1=running, 0=stopped, 2=error, 3=init, 4=unused)",
		},
		[]string{"service", "type"},
	)
	hostModeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "host_mode",
			Help: "Current mode of host (1=endhost, 2=infrastructure_host)",
		},
		[]string{"service", "type"},
	)
	serviceStatusString = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_status_info",
			Help: "Service status as a string label",
		},
		[]string{"service", "status"},
	)
)

func mapStatusToValue(status string) float64 {
	switch status {
	case SERVICE_STATUS_RUNNING:
		return 1
	case SERVICE_STATUS_STOPPED:
		return 0
	case SERVICE_STATUS_ERROR:
		return 2
	case SERVICE_STATUS_INIT:
		return 3
	case SERVICE_STATUS_UNUSED:
		return 4
	default:
		return -1
	}
}

func (asStatus *HostStatus) UpdateMetrics() {

	if asStatus.Mode == "endhost" {
		hostModeGauge.WithLabelValues("Mode", "core").Set(1)
	} else {
		hostModeGauge.WithLabelValues("Mode", "core").Set(2)
		serviceStatusGauge.WithLabelValues("BootstrapServer", "core").Set(mapStatusToValue(asStatus.BootstrapServer.Status))
		serviceStatusGauge.WithLabelValues("CertificateRenewal", "core").Set(mapStatusToValue(asStatus.CertificateRenewal.Status))
		for name, status := range asStatus.ControlServices {
			serviceStatusGauge.WithLabelValues(name, "control").Set(mapStatusToValue(status.Status))
		}

		for name, status := range asStatus.BorderRouters {
			serviceStatusGauge.WithLabelValues(name, "router").Set(mapStatusToValue(status.Status))
		}
	}

	serviceStatusGauge.WithLabelValues("Dispatcher", "core").Set(mapStatusToValue(asStatus.Dispatcher.Status))
	serviceStatusGauge.WithLabelValues("Daemon", "core").Set(mapStatusToValue(asStatus.Daemon.Status))

}

func RunPrometheusHTTPServer(url string) error {
	// Start the prometheus server
	h := http.NewServeMux()

	// Create a custom registry
	customRegistry := prometheus.NewRegistry()

	// Register only your custom metrics
	customRegistry.MustRegister(serviceStatusGauge)
	customRegistry.MustRegister(hostModeGauge)

	// Create a custom HTTP handler
	metricsHandler := promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{})

	h.Handle("/metrics", metricsHandler)
	// Define a basic health check endpoint
	h.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	go func() {
		for {
			Status.UpdateMetrics()
			time.Sleep(1 * time.Second)
		}
	}()
	log.Println("[Metrics] Starting HTTP prometheus server on", url)
	err := http.ListenAndServe(url, h)
	return err
}
