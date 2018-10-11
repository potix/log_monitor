package actor_plugger

import (
    "plugin"
    "os/user"
    "log"
    "regexp"
    "path/filepath"
    "io/ioutil"
    "github.com/pkg/errors"
)

type ActorPlugin interface {
    Initialize (err)
    AddFile(fileId string, fileName string)
    RemoveFile(fileId string, filename string)
    RenameFile(fileId string, filename string)
    ModifyFile(fileId string, filename string)
    Finalize()
}

type ActorPluginNewFunc func(configFile string)

type GetActorPluginInfoFunc func() (string, ActorNewFunc)

var registeredActorPlugins = make(map[string]ActorPluginNewFunc)

func registerActorPlugin(name string, actorPluginNewFunc ActorPluginNewFunc) {
        registeredActorPlugins[name] = actorPluginNewFunc
}

func registerActorPlugin(getActorPluginInfoFunc GetActorInfoFunc) {
    name, actorNewFunc := getActorPluginInfoFunc()	
    registerActorPlugin(name, actorNewFunc)
}

func getActorPluginSymbole(openedPlugin *plugin.Plugin) (GetActorPluginInfoFunc, error) {
    s, err := openedPlugin.Lookup("GetPluginInfo")
    if err != nil {
        return nil, errors.Wrap(err, "not found GetPluginInfoFunc symbole")
    }
    return s.(GetActorPluginInfoFunc), nil
}

func loadActorPlugin(pluginFilePath string) (error) {
    openedPlugin, err := plugin.Open(pluginFilePath)
    if err != nil {
	return errors.Wrapf(err, "can not open plugin file (file = %v)", pluginFilePath)
    }
    f, err := getActorPluginSymbole(openedPlugin)
    if err != nil {
	return errors.Wrapf(err, "not plugin file (file = %v)", pluginFilePath)
    }
    a.registerActorPlugin(f)
}

func fixupActorPluginPath(pluginPath string) (string) {
    u, err := user.Current()
    if err != nil {
        return pluginPath
    }
    re := regexp.MustCompile("^~/")
    return re.ReplaceAllString(pluginPath, u.HomeDir+"/")
}

func loadActorPluginFiles(pluginPath string) (error) {
    if pluginPath == "" {
        return errors.New("invalid plugin path")
    }
    pluginPath := fixupPluginPath(pluginPath)
    fileList, err := ioutil.ReadDir(pluginPath)
    if err != nil {
        return errors.Wrapf(err, "can not read directory (path = %v)", pluginPath)
    }
    for _, file := range fileList {
        if file.IsDir() {
            continue
        }
	ext := filepath.Ext(file.Name())
	if ext != ".so" && ext != ".dylib" {
	    continue
	}
	pluginFilePath := filepath.Join(pluginPath, file.Name())
	err := loadActorPlugin(pluginType, pluginFilePath)
	if err != nil {
	    log.Printf("can not load plugin (file = %v)", pluginFilePath)
	    continue
	}
    }
    return nil
}

func LoadActorPlugins(pluginPath string) (error) {
	return p.loadActorPluginFiles(pluginPath)
}

func GetActorPlugin(name string) (ActorPluginNewFunc, bool) {
        actorPluginNewFunc, ok := registeredActorPlugins[name]
        return actorPluginNewFunc, ok
}

