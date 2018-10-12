package main

import (
    "fmt"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actor_plugger"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
    "github.com/potix/log_monitor/actor_plugins/matcher/file_checker"
    "github.com/potix/log_monitor/actor_plugins/matcher/rule_manager"
)

type Matcher struct {
    configurator *configurator.Configurator
    fileChecker *file_checker.FileChecker
    ruleManager *rule_manager.RuleManager
}

func (m *Matcher) Initialize() (error) {
    fmt.Println("initialize")
    return nil
}

func (m *Matcher) AddFile(fileId string, fileName string) {
    fmt.Println("AddFile", fileId, fileName)
}

func (m *Matcher) RemoveFile(fileId string, fileName string) {
    fmt.Println("RemoveFile", fileId, fileName)
}

func (m *Matcher) RenameFile(fileId string, oldFileName string, newFileName string) {
    fmt.Println("RenameFile", fileId, oldFileName, newFileName)
}

func (m *Matcher) ModifyFile(fileId string, fileName string) {
    fmt.Println("modifyFile", fileId, fileName)
}

func (m *Matcher) ExpireFile(fileId string, fileName string) {
    fmt.Println("expireFile", fileId, fileName)
}

func (m *Matcher) Finalize() {
    fmt.Println("finalize")
}

func NewMatcher(configFile string) (actor_plugger.ActorPlugin, error) {
    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }

    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }

    fmt.Println(config)

    ruleManager, err := rule_manager.NewRuleManager(configurator)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create rule manager")
    }

    fileChecker, err := file_checker.NewFileChecker(ruleManager)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create file checker")
    }

    return &Matcher {
        configurator: configurator,
        fileChecker: fileChecker,
        ruleManager: ruleManager,
    }, nil
}

func GetActorPluginInfo() (string, actor_plugger.ActorPluginNewFunc) {
    return "matcher", NewMatcher
}
