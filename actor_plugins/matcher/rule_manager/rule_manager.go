package rule_manager

import (
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
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

func NewRuleManager(configurator *configurator.Configurator) (*RuleManager, error) {
    return  &RuleManager {
        configurator: configurator,
    }, nil
}
