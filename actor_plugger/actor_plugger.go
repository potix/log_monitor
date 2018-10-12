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
    Initialize() (error)
    AddFile(fileId string, fileName string)
    RemoveFile(fileId string, fileName string)
    RenameFile(fileId string, oldFileName string, newFileName string)
    ModifyFile(fileId string, fileName string)
    ExpireFile(fileId string, fileName string)
    Finalize()
}

const (
   GetActorPluginInfo string = "GetActorPluginInfo"
)

type ActorPluginNewFunc func(configFile string) (ActorPlugin, error)

type GetActorPluginInfoFunc func() (string, ActorPluginNewFunc)

type actorPluginInfo struct {
     actorPluginFilePath string
     actorPluginNewFunc ActorPluginNewFunc
}

var registeredActorPlugins = make(map[string]*actorPluginInfo)

func registerActorPlugin(pluginFilePath string,  getActorPluginInfoFunc GetActorPluginInfoFunc) {
    name, actorPluginNewFunc := getActorPluginInfoFunc()	
    registeredActorPlugins[name] = &actorPluginInfo {
        actorPluginFilePath: pluginFilePath,
        actorPluginNewFunc: actorPluginNewFunc,
    }
}

func getActorPluginSymbole(openedPlugin *plugin.Plugin) (GetActorPluginInfoFunc, error) {
    s, err := openedPlugin.Lookup(GetActorPluginInfo)
    if err != nil {
        return nil, errors.Wrap(err, "not found GetPluginInfoFunc symbole")
    }
    //return s.(GetActorPluginInfoFunc), nil
    return s.(func() (string, ActorPluginNewFunc)), nil
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
    registerActorPlugin(pluginFilePath, f)
    return nil
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
    pluginPath = fixupActorPluginPath(pluginPath)
    fileList, err := ioutil.ReadDir(pluginPath)
    if err != nil {
        return errors.Wrapf(err, "can not read directory (path = %v)", pluginPath)
    }
    for _, file := range fileList {
        if file.IsDir() {
            newPluginPath := filepath.Join(pluginPath, file.Name())
            err := loadActorPluginFiles(newPluginPath)
            if err != nil {
                log.Printf("can not load plugin files (%v): %v", newPluginPath, err)
            }
            continue
        }
	ext := filepath.Ext(file.Name())
	if ext != ".so" && ext != ".dylib" {
	    continue
	}
	pluginFilePath := filepath.Join(pluginPath, file.Name())
	err := loadActorPlugin(pluginFilePath)
	if err != nil {
	    log.Printf("can not load plugin file (%v): %v", pluginFilePath, err)
	    continue
	}
    }
    return nil
}

func LoadActorPlugins(pluginPath string) (error) {
	return loadActorPluginFiles(pluginPath)
}

func GetActorPlugin(name string) (string, ActorPluginNewFunc, bool) {
        info, ok := registeredActorPlugins[name]
        return info.actorPluginFilePath, info.actorPluginNewFunc, ok
}

