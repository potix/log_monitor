package main

import (
    "os"
    "os/signal"
    "syscall"
    "log"
    "github.com/potix/log_monitor/configurator"
    "github.com/potix/log_monitor/rule_manager"
    "github.com/potix/log_monitor/event_manager"
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
    flag.Var(&configFile, "config", "config file")
    flag.Parse()

    configurator, err := configurator.NewConfigurator(configFile)
    if (err != nil) {
	log.Fatal("can not create configurator: (%v)", err)
    }

    eventManager, err := event_manager.NewEventManager(configurator)
    if err != nil {
      log.Fatal("can not create event manager", err)
    }
    eventManager.Start()

    signalWait()

    eventManager.Stop()
    eventManager.Clean()
}




