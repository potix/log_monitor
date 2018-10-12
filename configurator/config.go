package configurator

type Target struct {
    Path string `json:"path" yaml:"path" toml:"path"`
    Expire int64 `json:"expire" yaml:"expire" toml:"expire"`
}

type Config struct {
    WorkDir string `json:"work_dir" yaml:"work_dr" toml:"work_dir"`  
    ActorPluginPath string `json:"actor_plugin_path" yaml:"actor_plugin_path" toml:"actor_plugin_path"`  
    ActorName string `json:"actor_name" yaml:"actor_name" toml:"actor_name"`  
    ActorConfig string `json:"actor_config" yaml:"actor_config" toml:"actor_config"`  
    Targets map[string]Target `json:"targets" yaml:"targets" toml:"targets"`  
}
