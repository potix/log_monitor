package file_checker

import (
    "github.com/potix/log_monitor/actor_plugins/matcher/rule_manager"
)

type FileChecker struct {
    ruleManager *rule_manager.RuleManager
}



func NewFileChecker(ruleManager *rule_manager.RuleManager) (*FileChecker, error) {
	return &FileChecker {
		ruleManager : ruleManager,
        }, nil
}

