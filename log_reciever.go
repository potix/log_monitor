package main

import (
    "os/signal"
    "syscall"
    "flag"
    "log"
    "github.com/potix/log_monitor/reciever"
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
    flag.StringVar(&configFile, "config", "log_reciever.toml", "config file")
    flag.Parse()

    configurator, err := configurator.NewConfigurator(configFile)
    if (err != nil) {
        log.Fatalf("can not create configurator: %v", err)
    }

    config, err := configurator.LoadLogRecieverConfig()
    if err != nil {
        log.Fatalf("can not load config: %v", err)
    }

    signalWait()
}
