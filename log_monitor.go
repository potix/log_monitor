package main

import (
    "os"
    "os/signal"
    "syscall"
    "log"
    //"github.com/potix/log_monitor/configurator"
    "github.com/potix/log_monitor/event_manager"
    //"github.com/potix/log_monitor/rule_manager"
    //"github.com/potix/log_monitor/file_monitor"
)

func signalWait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGQUIT:
			fallthrough
		case syscall.SIGTERM:
			return
		default:
			log.Printf("unexpected signal (sig = %v)", sig)
		}
	}
}

func main() {
    eventManager, err := event_manager.NewEventManager()
    if err != nil {
      log.Fatal(err)
    }
    eventManager.Start()


    eventManager.AddPath(".")
    signalWait()

    eventManager.Stop()
    eventManager.Clean()
}




