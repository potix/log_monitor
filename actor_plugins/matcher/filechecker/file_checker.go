package filechecker

import (
    "github.com/potix/log_monitor/actor_plugins/matcher/rulemanager"
)

// FileChecker is FileChecker
type FileChecker struct {
    ruleManager *rulemanager.RuleManager
}



// NewFileChecker is create new file checker
func NewFileChecker(ruleManager *rulemanager.RuleManager) (*FileChecker, error) {
	return &FileChecker {
		ruleManager : ruleManager,
        }, nil
}

