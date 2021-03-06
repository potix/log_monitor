package configurator

// Config is config
type Config struct {
    SavePrefix string `json:"save_prefix" yaml:"save_prefix" toml:"save_prefix"`
    AddrPort string `json:"addr_port" yaml:"addr_port" toml:"addr_port"`
    Label string `json:"label" yaml:"label" toml:"label"`
    FlushInterval uint32 `json:"flush_interval" yaml:"flush_interval" toml:"flush_interval"`
}
