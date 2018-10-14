package eventmanager

import (
    "os"
    "syscall"
    "fmt"
    "log"
    "sync"
    "os/user"
    "time"
    "regexp"
    "io/ioutil"
    "path/filepath"
    "github.com/pkg/errors"
    "github.com/fsnotify/fsnotify"
    "github.com/potix/log_monitor/configurator"
    "github.com/potix/log_monitor/actorplugger"
)

type pathInfo struct {
    expire int64
    actorName string
    actorConfig string
}

type fileStatus struct {
    name string
    dirty bool
    lastUpdate int64
    actorPlugin actorplugger.ActorPlugin
    mutex *sync.Mutex
}

// EventManager is event manager
type EventManager struct{
    loopEnd  chan bool
    watcher *fsnotify.Watcher
    paths map[string]*pathInfo
    pathsMutex *sync.Mutex
    files map[string]*fileStatus
    filesMutex *sync.Mutex
}

func (e *EventManager) newActorPlugin(path string) (actorplugger.ActorPlugin, error) {
    parentDir := filepath.Dir(path)
    parentPathInfo, ok := e.getPathInfo(parentDir)
    if !ok {
        return nil, errors.Errorf("not found parent path info (%v, %v)", path, parentDir)
    }
    actorPluginFilePath, actorPluginNewFunc, ok := actorplugger.GetActorPlugin(parentPathInfo.actorName)
    actorPluginDir := filepath.Dir(actorPluginFilePath)
    actorPluginConfigPath := filepath.Join(actorPluginDir, parentPathInfo.actorConfig)
    return  actorPluginNewFunc(actorPluginConfigPath)
}

func (e *EventManager) foundFile(fileID string, name string) {
    actorPlugin, err := e.newActorPlugin(name)
    if err != nil {
	log.Printf("can not create actor plugin (%v, %v): %v", fileID, name, err)
        return
    }
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    _, ok := e.files[fileID]
    if ok {
        log.Printf("already exists file (%v, %v)", fileID, name)
        return
    }
    e.files[fileID] = &fileStatus {
        name: name,
        dirty: true,
	actorPlugin: actorPlugin,
	lastUpdate: time.Now().Unix(),
	mutex : new(sync.Mutex),
    }
    actorPlugin.FoundFile(fileID, name)
}

func (e *EventManager) createdFile(fileID string, event fsnotify.Event) {
    actorPlugin, err := e.newActorPlugin(event.Name)
    if err != nil {
	log.Printf("can not create actor plugin (%v, %v): %v", fileID, event.Name, err)
        return
    }
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    oldStatus, ok := e.files[fileID]
    if ok {
        if oldStatus.name != event.Name {
            log.Printf("change name (%v, %v -> %v)", fileID, oldStatus.name, event.Name)
	    oldStatus.mutex.Lock()
	    defer oldStatus.mutex.Unlock()
            oldStatus.name = event.Name
	    oldStatus.lastUpdate = time.Now().Unix()
        } else {
            log.Printf("already exists file (%v, %v)", fileID, event.Name)
        }
        return
    }
    e.files[fileID] = &fileStatus {
        name: event.Name,
        dirty: true,
	actorPlugin: actorPlugin,
	lastUpdate: time.Now().Unix(),
	mutex : new(sync.Mutex),
    }
}

func (e *EventManager) removedFile(fileID string, event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    _, ok := e.files[fileID]
    if !ok {
        log.Printf("not exists file (%v, %v)", fileID, event.Name)
        return
    }
    delete(e.files, fileID)
}

func (e *EventManager) setDirtyFile(fileID string, event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[fileID]
    if !ok {
        log.Printf("not exists file (%v, %v)", fileID, event.Name)
        return
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    status.dirty = true
    status.lastUpdate = time.Now().Unix()
}

func (e *EventManager) modifiedFile(fileID string) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[fileID]
    if !ok {
        log.Printf("not exists file (%v)", fileID)
        return
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    if !status.dirty {
        log.Printf("not dirty  (%v, %v)", fileID, status.name)
        return
    }


    // XXX TODO modify



    status.dirty = false
}

func (e *EventManager) eventLoop() {
    for {
        select {
        case <- e.loopEnd:
            return
        case event, ok := <-e.watcher.Events:
            if !ok {
                 // end loop
                 return
            }
            log.Printf("%v", event)
            if event.Op&fsnotify.Chmod == fsnotify.Chmod {
                // nop
                break
            }
            fileID, info, err := e.getFileInfo(event.Name)
            if err != nil {
                log.Printf("can not get file id (%v)", event.Name)
                break
            }
            if event.Op&fsnotify.Create == fsnotify.Create {
               if info.IsDir() {
                   parent := filepath.Dir(event.Name)
                   info, ok := e.paths[parent]
                   if !ok {
                       log.Printf("not found parent %v", parent) 
                   } else {
                       e.addPath(event.Name, info.expire, info.actorName, info.actorConfig)
                   }
               } else {
                   e.createdFile(fileID, event)
               }
            }
            if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
               if info.IsDir() {
                   e.deletePath(event.Name)
               } else {
                   e.removedFile(fileID, event)
               }
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
               if !info.IsDir() {
                   e.setDirtyFile(fileID, event)
               }
            }
            if !info.IsDir() {
                e.modifiedFile(fileID)
            }
        case err, ok := <-e.watcher.Errors:
            if !ok {
                 // end loop
                 return
            }
            log.Println("error: ", err)
        }
    }
}

func (e *EventManager) addPath(path string, expire int64, actorName string, actorConfig string) (error) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
        _, ok := e.paths[path]
        if ok {
            log.Printf("already exists path (%v)", path)
            return nil;
        }
        err := e.watcher.Add(path)
        if err != nil {
            errors.Wrap(err, "can not add path to watcher")
	} else {
            e.paths[path] = &pathInfo{
                expire: expire,
                actorName: actorName,
                actorConfig: actorConfig,
            }
        }
        return nil
}

func (e *EventManager) deletePath(path string) (error) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
        _, ok := e.paths[path]
        if !ok {
            log.Printf("not exists path (%v)", path)
            return nil;
        }
        err := e.watcher.Remove(path)
        if err != nil {
            errors.Wrap(err, "can not delete path from watcher")
	} else {
            delete(e.paths, path)
        }
        return nil
}

func (e *EventManager) getPathInfo(path string) (*pathInfo, bool) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
	pathInfo, ok := e.paths[path]
	return pathInfo, ok
}

// Start is start
func (e *EventManager) Start() (error) {
     e.loopEnd = make(chan bool)
     go e.eventLoop()
     return nil
}

// Stop is stop
func (e *EventManager) Stop() {
     close(e.loopEnd)
}

// Clean is clean
func (e *EventManager) Clean() {
     e.watcher.Close()
}

func  (e *EventManager) getFileInfo(filePath string) (string, os.FileInfo, error){
        info, err := os.Stat(filePath)
        if err != nil {
            return "", nil, errors.Wrapf(err, "can not get file info (%v)", filePath)
        }
        stat, ok := info.Sys().(*syscall.Stat_t)
        if !ok {
            return "", nil, errors.Wrapf(err, "can not get file stat (%v)", filePath)
        }
        fileID := fmt.Sprintf("%v:%v", stat.Dev, stat.Ino)
        log.Println("fileID:", fileID)
	return fileID, info, nil
}

func (e *EventManager) fixupPath(targetPath string) (string) {
    u, err := user.Current()
    if err != nil {
        return targetPath
    }
    re := regexp.MustCompile("^~/")
    return re.ReplaceAllString(targetPath, u.HomeDir+"/")
}

func (e *EventManager) addTargets(targetPath string, expire int64, actorName string, actorConfig string) {
    targetPath = e.fixupPath(targetPath)
    fileList, err := ioutil.ReadDir(targetPath)
    if err != nil {
        log.Printf("can not read dir (%v): %v", targetPath, err)
        return
    }
    err = e.addPath(targetPath, expire, actorName, actorConfig)
    if err != nil {
        log.Printf("can not add path (%v): %v", targetPath, err)
        return
    }
    for _, file := range fileList {
        newPath := filepath.Join(targetPath, file.Name())
        if file.IsDir() {
            e.addTargets(newPath, expire, actorName, actorConfig)
	    continue
        }
	fileID, _, err := e.getFileInfo(newPath)
        if err != nil {
            log.Printf("can not get file id (%v)", newPath)
            continue
        }
	e.foundFile(fileID, newPath)
    }
}

// NewEventManager is create new event manager
func NewEventManager(configurator *configurator.Configurator) (*EventManager, error) {
    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrap(err, "can not load config")
    }
    watcher, err :=  fsnotify.NewWatcher()
    if err != nil {
        return nil, errors.Wrapf(err, "can not create event manager")
    }
    eventManager := &EventManager {
        loopEnd: make(chan bool),
        watcher : watcher,
        paths : make(map[string]*pathInfo),
        pathsMutex : new(sync.Mutex),
        files : make(map[string]*fileStatus),
        filesMutex : new(sync.Mutex),
    }
    for _, targetInfo := range config.Targets {
         eventManager.addTargets(targetInfo.Path, targetInfo.Expire, targetInfo.ActorName, targetInfo.ActorConfig)
    }
    return eventManager, nil
}

