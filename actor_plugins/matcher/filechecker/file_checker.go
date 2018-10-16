package filechecker

import (
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
)

// FileChecker is FileChecker
type FileChecker struct {
     config *configurator.Config
}

// NewFileChecker is create new file checker
func NewFileChecker(config *configurator.Config) (*FileChecker, error) {
	return &FileChecker {
		config: config,
        }, nil
}

