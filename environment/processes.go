package environment

import (
	"log"
	"os/exec"
	"syscall"
)

var processes []*exec.Cmd
var killInitiated bool

func KillAllChilds() {
	if killInitiated {
		return
	}
	killInitiated = true
	log.Println("[Signal] Killing all child processes")
	for _, p := range processes {
		// TODO: Not implemented on Windows...
		p.Process.Signal(syscall.SIGTERM)
	}
}

func init() {
	processes = make([]*exec.Cmd, 0)
}
