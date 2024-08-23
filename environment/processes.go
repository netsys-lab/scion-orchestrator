package environment

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
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
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-cancelChan
		log.Printf("[Signal] Caught signal %v", sig)
		KillAllChilds()
	}()
}
