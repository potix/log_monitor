package main

import (
    "fmt"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actorplugger"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
    "github.com/potix/log_monitor/actor_plugins/matcher/filechecker"
    "github.com/potix/log_monitor/actor_plugins/matcher/rulemanager"
)

// Matcher is matcher
type Matcher struct {
    configurator *configurator.Configurator
    fileChecker *filechecker.FileChecker
    ruleManager *rulemanager.RuleManager
}

// Initialize is initialize
func (m *Matcher) Initialize() (error) {
    fmt.Println("initialize")
    return nil
}

// FoundFile is add file
func (m *Matcher) FoundFile(fileName string, fileID string) {
    fmt.Println("FoundFile", fileID, fileName)
}

// CreatedFile is add file
func (m *Matcher) CreatedFile(fileName string, fileID string) {
    fmt.Println("CreatedFile", fileID, fileName)
}

// RemovedFile is remove file
func (m *Matcher) RemovedFile(fileName string, fileID string) {
    fmt.Println("RemovedFile", fileID, fileName)
}

// RenamedFile is rename file
func (m *Matcher) RenamedFile(oldFileName string, newFileName string, fileID string) {
    fmt.Println("RenamedFile", fileID, oldFileName, newFileName)
}

// ModifiedFile is modify file
func (m *Matcher) ModifiedFile(fileName string, fileID string) {
    fmt.Println("modifiedFile", fileID, fileName)
}

// ExpiredFile is expire file
func (m *Matcher) ExpiredFile(fileName string, fileID string) {
    fmt.Println("expiredFile", fileID, fileName)
}

// Finalize is finalize
func (m *Matcher) Finalize() {
    fmt.Println("finalize")
}

// NewMatcher is create new matcher
func NewMatcher(configFile string) (actorplugger.ActorPlugin, error) {
    fmt.Println("configFile = ", configFile)

    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }

    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }

    fmt.Println("config = ", config)

    ruleManager, err := rulemanager.NewRuleManager(configurator)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create rule manager")
    }

    fileChecker, err := filechecker.NewFileChecker(ruleManager)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create file checker")
    }

    return &Matcher {
        configurator: configurator,
        fileChecker: fileChecker,
        ruleManager: ruleManager,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "matcher", NewMatcher
}
