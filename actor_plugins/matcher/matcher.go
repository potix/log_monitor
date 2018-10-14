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

// CreateFile is add file
func (m *Matcher) CreateFile(fileID string, fileName string) {
    fmt.Println("CreateFile", fileID, fileName)
}

// RemoveFile is remove file
func (m *Matcher) RemoveFile(fileID string, fileName string) {
    fmt.Println("RemoveFile", fileID, fileName)
}

// RenameFile is rename file
func (m *Matcher) RenameFile(fileID string, oldFileName string, newFileName string) {
    fmt.Println("RenameFile", fileID, oldFileName, newFileName)
}

// ModifyFile is modify file
func (m *Matcher) ModifyFile(fileID string, fileName string) {
    fmt.Println("modifyFile", fileID, fileName)
}

// ExpireFile is expire file
func (m *Matcher) ExpireFile(fileID string, fileName string) {
    fmt.Println("expireFile", fileID, fileName)
}

// Finalize is finalize
func (m *Matcher) Finalize() {
    fmt.Println("finalize")
}

// NewMatcher is create new matcher
func NewMatcher(configFile string) (actorplugger.ActorPlugin, error) {
    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }

    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }

    fmt.Println(config)

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
