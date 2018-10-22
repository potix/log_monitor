package configurator

// Notifier is Notifier
type Notifier struct {
    Name string
    Config string
}

// MsgMatcher is MsgMatcher
type MsgMatcher struct {
    Pattern string
}

// PathMatcher is Matcher
type PathMatcher struct {
    Pattern string
    MsgMatchers []*MsgMatcher
    Expire int64
    Notifier string
    Mode string
}

// Config is Config
type Config struct {
    AutoReload int64
    Notifiers map[string]*Notifier
    PathMatchers []*PathMatcher
}
