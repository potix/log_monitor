package configurator

// Notifier is Notifier
type Notifier struct {
    Name string `json:"name" yaml:"name" toml:"name"`
    Config string `json:"config" yaml:"config" toml:"config"`
}

// MsgMatcher is MsgMatcher
type MsgMatcher struct {
    Pattern string `json:"pattern" yaml:"pattern" toml:"pattern"`
}

// PathMatcher is Matcher
type PathMatcher struct {
    Pattern string `json:"pattern" yaml:"pattern" toml:"pattern"`
    SkipNotify bool `json:"skip_notify" yaml:"skip_notify" toml:"skip_notify"`
    Label string `json:"label" yaml:"label" toml:"label"`
    MsgMatchers []*MsgMatcher `json:"msg_matchers" yaml:"msg_matchers" toml:"msg_matchers"`
    Notifiers []*Notifier `json:"notifiers" yaml:"notifiers" toml:"notifiers"`
}

// Config is Config
type Config struct {
    AutoReload int64 `json:"auto_reload" yaml:"auto_reload" toml:"auto_reload"`
    NotifierPluginPath string  `json:"notifier_plugin_path" yaml:"notifier_plugin_path" toml:"notifier_plugin_path"`
    SkipNotify bool `json:"skip_notify" yaml:"skip_notify" toml:"skip_notify"`
    PathMatchers []*PathMatcher `json:"path_matchers" yaml:"path_matchers" toml:"path_matchers"`
}
