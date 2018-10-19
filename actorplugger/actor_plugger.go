package actorplugger

import (
    "plugin"
    "os/user"
    "log"
    "regexp"
    "path/filepath"
    "io/ioutil"
    "github.com/pkg/errors"
)

// ActorPlugin is actor plugin
type ActorPlugin interface {
    FoundFile(fileName string, fileID string, trackLinkFilePath string)
    CreatedFile(fileName string, fileID string, trackLinkFilePath string)
    RemovedFile(fileName string, fileID string, trackLinkFilePath string)
    RenamedFile(oldFileName string, newFileName string, fileID string)
    ModifiedFile(fileName string, fileID string)
}

const (
   // GetActorPluginInfo is GetActorPluginInfo symbple
   GetActorPluginInfo string = "GetActorPluginInfo"
)

// ActorPluginNewFunc is ActorPluginNewFunc
type ActorPluginNewFunc func(callers string, configFile string) (ActorPlugin, error)

// GetActorPluginInfoFunc is GetActorPluginInfoFunc
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
        newPath := filepath.Join(pluginPath, file.Name())
        if file.IsDir() {
            err := loadActorPluginFiles(newPath)
            if err != nil {
                log.Printf("can not load plugin files (%v): %v", newPath, err)
            }
            continue
        }
	ext := filepath.Ext(file.Name())
	if ext != ".so" && ext != ".dylib" {
	    continue
	}
	err := loadActorPlugin(newPath)
	if err != nil {
	    log.Printf("can not load plugin file (%v): %v", newPath, err)
	    continue
	}
    }
    return nil
}

// LoadActorPlugins is load actor Plugins
func LoadActorPlugins(pluginPath string) (error) {
	return loadActorPluginFiles(pluginPath)
}

// GetActorPlugin is get actor plugin
func GetActorPlugin(name string) (string, ActorPluginNewFunc, bool) {
        info, ok := registeredActorPlugins[name]
        return info.actorPluginFilePath, info.actorPluginNewFunc, ok
}

