package rulemanager

import (
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
)

// RuleManager is RuleManager
type RuleManager struct {
    configurator *configurator.Configurator
}


// Start is tart
func (r *RuleManager) Start() (error) {
     return nil
}

// Stop is stop
func (r *RuleManager) Stop() {
}

// Clean is clean
func (r *RuleManager) Clean() {
}

// NewRuleManager is create new rule manager
func NewRuleManager(configurator *configurator.Configurator) (*RuleManager, error) {
    return  &RuleManager {
        configurator: configurator,
    }, nil
}
