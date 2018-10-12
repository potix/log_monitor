package rule_manager

import (
    "github.com/potix/log_monitor/configurator"
)

type RuleManager struct {
    configurator *configurator.Configurator
}


func (r *RuleManager) Start() (error) {
     return nil
}

func (r *RuleManager) Stop() {
}

func (r *RuleManager) Clean() {
}

func NewRuleManager(configurator *configurator.Configurator) (*RuleManager) {
    return  &RuleManager {
        configurator: configurator,
    }
}
