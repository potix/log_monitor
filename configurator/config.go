package configurator

// Target is target
type Target struct {
    Path string `json:"path" yaml:"path" toml:"path"`
    ActorName string `json:"actor_name" yaml:"actor_name" toml:"actor_name"`  
    ActorConfig string `json:"actor_config" yaml:"actor_config" toml:"actor_config"`  
}

// LogMonitorConfig is config of log monitor
type LogMonitorConfig struct {
    WorkDir string `json:"work_dir" yaml:"work_dr" toml:"work_dir"`  
    ActorPluginPath string `json:"actor_plugin_path" yaml:"actor_plugin_path" toml:"actor_plugin_path"`  
    Targets map[string]Target `json:"targets" yaml:"targets" toml:"targets"`  
}

// LogRecieverConfig is config of log reciever
type LogRecieverConfig struct {
    AddrPort string `json:"addr_port" yaml:"addr_port" toml:"addr_port"`
    StoreRoot string `json:"store_root" yaml:"store_root" toml:"store_root"` 
    PathFormat string `json:"path_format" yaml:"path_format" toml:"path_format"` // default "${LABEL}/${HOST}_${ADDR}/${PATH}"
}
