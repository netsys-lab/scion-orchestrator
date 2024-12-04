//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/scionproto/scion/pkg/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

// Windows-specific configuration keys
const (
	cfgLogFile = "logfile"
)

// Event IDs for the system event log used by the launcher
const (
	eventIdStarted = 1
	eventIdStopped = 2
	eventIdFailed  = 3
)

var isService bool

func init() {
	RunFunc = runWindows
}

func runWindows() error {
	var err error
	isService, err = svc.IsWindowsService()

	executable := filepath.Base(os.Args[0])

	if isService {
		elog, err := eventlog.Open(executable)
		if err != nil {
			return err
		}
		wh.elog = elog
		err = svc.Run(executable, wh)
	} else {
		dlog := debug.New(executable)
		wh.dlog = dlog
		err = debug.Run(executable, wh)
	}
	if err != nil {
		// ec := err.(syscall.Errno)
		return err
	}

	return nil

}

var wh = &WinHandler{}

type WinHandler struct {
	elog *eventlog.Log
	dlog *debug.ConsoleLog
}

func (a *WinHandler) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (appSpecificEc bool, ec uint32) {
	appSpecificEc = true
	ec = 0

	// Accept no controls until initialization has finished
	changes <- svc.Status{State: svc.Running, Accepts: 0}

	// Windows does not support POSIX signals, use a cancellable context to
	// initiate clean shutdown instead.
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() {
		defer log.HandlePanic()
		done <- realMain(ctx)
	}()

	// Service is ready to accept controls
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.Running, Accepts: accepted}
	a.elog.Info(eventIdStarted, "Service started")
loop:
	for {
		select {
		case a.svcErr = <-done:
			// Main exited on its own
			if a.svcErr != nil {
				ec = 1
				a.elog.Error(eventIdFailed, fmt.Sprintf("%v", a.svcErr))
			}
			cancel()
			return
		case c := <-r:
			// Service control signal
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				break loop
			default:
				msg := fmt.Sprintf("Unexpected service control request %d", c.Cmd)
				a.elog.Error(eventIdFailed, msg)
				panic(msg)
			}
		}
	}

	// If the main goroutine shuts down everything in time, this won't get
	// a chance to run.
	time.AfterFunc(5*time.Second, func() {
		defer log.HandlePanic()
		msg := "Main goroutine did not shut down in time (waited 5s). " +
			"It's probably stuck. Forcing shutdown."
		a.elog.Error(eventIdFailed, msg)
		panic(msg)
	})

	if a.svcErr = <-done; a.svcErr != nil {
		ec = 1
		a.elog.Error(eventIdFailed, fmt.Sprintf("%v", a.svcErr))
	}
	changes <- svc.Status{State: svc.Stopped}
	a.elog.Info(eventIdStopped, "Service stopped")
	return
}
