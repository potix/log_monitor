package main

import (
    "os"
    "os/signal"
    "syscall"
    "flag"
    "log"
    "github.com/potix/log_monitor/configurator"
    "github.com/potix/log_monitor/actor_plugger"
//    "github.com/potix/log_monitor/rule_manager"
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
    var configFile string
    flag.StringVar(&configFile, "config", "log_monitor.toml", "config file")
    flag.Parse()

    configurator, err := configurator.NewConfigurator(configFile)
    if (err != nil) {
	log.Fatalf("can not create configurator: %v", err)
    }

    config, err := configurator.Load()
    if err != nil {
	log.Fatalf("can not load config: %v", err)
    }
 
    err = os.Chdir(config.WorkDir)
    if err != nil {
	log.Fatalf("can not change dir (%v): %v", config.WorkDir, err)
    }

    err = actor_plugger.LoadActorPlugins(config.ActorPluginPath)
    if err != nil {
	log.Fatalf("can not load actor plugins (%v): ", config.ActorPluginPath, err)
    }

    eventManager, err := event_manager.NewEventManager(configurator)
    if err != nil {
      log.Fatalf("can not create event manager: %v ", err)
    }
    err = eventManager.Start()
    if err != nil {
      log.Fatalf("can not start event manager: %v ", err)
    }

    signalWait()

    eventManager.Stop()
    eventManager.Clean()
}




