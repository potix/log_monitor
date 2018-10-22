package configurator

// Actor is actor
type Actor struct {
    Name string `json:"name" yaml:"name" toml:"name"`  
    Config string `json:"config" yaml:"config" toml:"config"`  
}

// Target is target
type Target struct {
    Path string `json:"path" yaml:"path" toml:"path"`
    Actors []*Actor `json:"actors" yaml:"actors" toml:"actors"`
}

// LogMonitorConfig is config of log monitor
type LogMonitorConfig struct {
    WorkDir string `json:"work_dir" yaml:"work_dr" toml:"work_dir"`  
    ActorPluginPath string `json:"actor_plugin_path" yaml:"actor_plugin_path" toml:"actor_plugin_path"`  
    Targets map[string]*Target `json:"targets" yaml:"targets" toml:"targets"`  
}

// LogRecieverConfig is config of log reciever
type LogRecieverConfig struct {
    AddrPort string `json:"addr_port" yaml:"addr_port" toml:"addr_port"`
    Path string `json:"path" yaml:"path" toml:"path"` 
    PathFormat string `json:"path_format" yaml:"path_format" toml:"path_format"` // default "${LABEL}/${HOST}_${ADDR}/${FILE_PATH}"
}
