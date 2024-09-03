package environment

import (
	"log"
	"os/exec"
	"syscall"
)

var processes []*exec.Cmd
var killInitiated bool

func KillAllChilds() bool {
	if killInitiated {
		return true
	}
	killInitiated = true
	log.Println("[Signal] Killing all child processes")
	processesKilled := false
	for _, p := range processes {
		// TODO: Not implemented on Windows...
		p.Process.Signal(syscall.SIGTERM)
		processesKilled = true
	}

	return processesKilled
}

func init() {
	processes = make([]*exec.Cmd, 0)
}
