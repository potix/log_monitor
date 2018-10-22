package configurator

type Notifier struct {
    Name string
    Config string
}

type Matcher struct {
    PathPattern string
    MsgPattern string
    Expire int64
    Notifier string
    Mode string
}

type Config struct {
    AutoReload bool
    Notifiers map[string]*Notifier
    Matchers map[string]*Matcher
}
