package metrics

import "encoding/json"

const (
	SERVICE_STATUS_RUNNING = "running"
	SERVICE_STATUS_STOPPED = "stopped"
	SERVICE_STATUS_ERROR   = "error"
	SERVICE_STATUS_INIT    = "init"
	SERVICE_STATUS_UNUSED  = "unused"
)

type ServiceStatus struct {
	Status  string
	Message string
}

type HostStatus struct {
	Mode               string
	BootstrapServer    ServiceStatus
	Dispatcher         ServiceStatus
	Daemon             ServiceStatus
	CertificateRenewal ServiceStatus
	ControlServices    map[string]ServiceStatus
	BorderRouters      map[string]ServiceStatus
	Status             string
	LastUpdated        string
}

var Status *HostStatus

func (hostStatus *HostStatus) Json() ([]byte, error) {
	return json.MarshalIndent(hostStatus, "", "  ")
}

func Init() {
	Status = &HostStatus{
		BootstrapServer: ServiceStatus{
			Status: SERVICE_STATUS_INIT,
		},
		Dispatcher: ServiceStatus{
			Status: SERVICE_STATUS_INIT,
		},
		Daemon: ServiceStatus{
			Status: SERVICE_STATUS_INIT,
		},
		CertificateRenewal: ServiceStatus{
			Status: SERVICE_STATUS_INIT,
		},
		ControlServices: map[string]ServiceStatus{},
		BorderRouters:   map[string]ServiceStatus{},
	}
}
