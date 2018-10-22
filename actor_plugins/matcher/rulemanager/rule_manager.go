package rulemanager

import (
    "time"
    "regexp"
    "log"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
)

// RuleManager is RuleManager
type RuleManager struct {
    config *configurator.Config
    configurator *configurator.Configurator
    finish chan bool
}

func (r *RuleManager) reloadLoop() {
    for {
        select {
        case <-time.After(time.Duration(r.config.AutoReload) *time.Second):
	    config, err := r.configurator.Load()
	    if err != nil {
	        log.Printf("can not load config: %v", err)
		break
	    }
	    r.config = config
        case <-r.finish:
            return
        }
    }
}

// GetRule is get rule
func (r *RuleManager) GetRule(filename string) (*configurator.PathMatcher) {
    for _, pathMatcher := range r.config.PathMatchers {
	matched, err := regexp.MatchString(pathMatcher.Pattern, filename)
	if err != nil {
	    log.Printf("can not match string (%v, %v): %v", pathMatcher.Pattern, filename, err)
            continue
	}
	if !matched {
	    continue
	}
	return pathMatcher
    }
    return nil
}

// Start is tart
func (r *RuleManager) Start() (error) {
    r.finish = make(chan bool)
    go r.reloadLoop()
    return nil
}

// Stop is stop
func (r *RuleManager) Stop() {
    close(r.finish)
}

// NewRuleManager is create new rule manager
func NewRuleManager(config *configurator.Config, configurator  *configurator.Configurator) (*RuleManager, error) {
    return &RuleManager {
	config: config,
        configurator: configurator,
    }, nil
}
