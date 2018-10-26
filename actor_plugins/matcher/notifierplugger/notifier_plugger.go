package notifierplugger

import (
    "plugin"
    "os/user"
    "log"
    "regexp"
    "path/filepath"
    "io/ioutil"
    "github.com/pkg/errors"
)

// NotifierPlugin is actor plugin
type NotifierPlugin interface {
    Notify(msg []byte, fileID string, fileName string, label string)
}

const (
   // GetNotifierPluginInfo is GetNotifierPluginInfo symbple
   GetNotifierPluginInfo string = "GetNotifierPluginInfo"
)

// NotifierPluginNewFunc is NotifierPluginNewFunc
type NotifierPluginNewFunc func(callers string, configFile string) (NotifierPlugin, error)

// GetNotifierPluginInfoFunc is GetNotifierPluginInfoFunc
type GetNotifierPluginInfoFunc func() (string, NotifierPluginNewFunc)

type notifierPluginInfo struct {
     notifierPluginFilePath string
     notifierPluginNewFunc NotifierPluginNewFunc
}

var registeredNotifierPlugins = make(map[string]*notifierPluginInfo)

func registerNotifierPlugin(pluginFilePath string,  getNotifierPluginInfoFunc GetNotifierPluginInfoFunc) {
    name, notifierPluginNewFunc := getNotifierPluginInfoFunc()	
    registeredNotifierPlugins[name] = &notifierPluginInfo {
        notifierPluginFilePath: pluginFilePath,
        notifierPluginNewFunc: notifierPluginNewFunc,
    }
}

func getNotifierPluginSymbole(openedPlugin *plugin.Plugin) (GetNotifierPluginInfoFunc, error) {
    s, err := openedPlugin.Lookup(GetNotifierPluginInfo)
    if err != nil {
        return nil, errors.Wrapf(err, "not found %v symbole", GetNotifierPluginInfo)
    }
    //return s.(GetNotifierPluginInfoFunc), nil
    return s.(func() (string, NotifierPluginNewFunc)), nil
}

func loadNotifierPlugin(pluginFilePath string) (error) {
    openedPlugin, err := plugin.Open(pluginFilePath)
    if err != nil {
	return errors.Wrapf(err, "can not open plugin file (file = %v)", pluginFilePath)
    }
    f, err := getNotifierPluginSymbole(openedPlugin)
    if err != nil {
	return errors.Wrapf(err, "not plugin file (file = %v)", pluginFilePath)
    }
    registerNotifierPlugin(pluginFilePath, f)
    return nil
}

func fixupNotifierPluginPath(pluginPath string) (string) {
    u, err := user.Current()
    if err != nil {
        return pluginPath
    }
    re := regexp.MustCompile("^~/")
    return re.ReplaceAllString(pluginPath, u.HomeDir+"/")
}

func loadNotifierPluginFiles(pluginPath string) (error) {
    if pluginPath == "" {
        return errors.New("invalid plugin path")
    }
    pluginPath = fixupNotifierPluginPath(pluginPath)
    fileList, err := ioutil.ReadDir(pluginPath)
    if err != nil {
        return errors.Wrapf(err, "can not read directory (path = %v)", pluginPath)
    }
    for _, file := range fileList {
        newPath := filepath.Join(pluginPath, file.Name())
        if file.IsDir() {
            err := loadNotifierPluginFiles(newPath)
            if err != nil {
                log.Printf("can not load plugin files (%v): %v", newPath, err)
            }
            continue
        }
	ext := filepath.Ext(file.Name())
	if ext != ".so" && ext != ".dylib" {
	    continue
	}
	err := loadNotifierPlugin(newPath)
	if err != nil {
	    log.Printf("can not load plugin file (%v): %v", newPath, err)
	    continue
	}
    }
    return nil
}

// LoadNotifierPlugins is load actor Plugins
func LoadNotifierPlugins(pluginPath string) (error) {
	return loadNotifierPluginFiles(pluginPath)
}

// GetNotifierPlugin is get actor plugin
func GetNotifierPlugin(name string) (string, NotifierPluginNewFunc, bool) {
        info, ok := registeredNotifierPlugins[name]
        return info.notifierPluginFilePath, info.notifierPluginNewFunc, ok
}

