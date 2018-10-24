package configurator

// Config is Config
type Config struct {
        HostPort      string `json:"hostPort"       yaml:"hostPort"       toml:"hostPort"`
        Username      string `json:"username"       yaml:"username"       toml:"username"`
        Password      string `json:"password"       yaml:"password"       toml:"password"`
        AuthType      string `json:"authType"       yaml:"authType"       toml:"authType"`
        UseTLS        bool   `json:"useTls"         yaml:"useTls"         toml:"useTls"`
        UseStartTLS   bool   `json:"useStartTls"    yaml:"useStartTls"    toml:"useStartTls"`
        From          string `json:"from"           yaml:"from"           toml:"from"`
        To            string `json:"to"             yaml:"to"             toml:"to"`
        SubjectFormat string `json:"subject_format" yaml:"subject_format" toml:"subject_format"`
}
