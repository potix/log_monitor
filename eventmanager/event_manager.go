package eventmanager

import (
    "os"
    "syscall"
    "fmt"
    "log"
    "sync"
    "os/user"
    "regexp"
    "io/ioutil"
    "path/filepath"
    "github.com/pkg/errors"
    "github.com/fsnotify/fsnotify"
    "github.com/potix/log_monitor/configurator"
    "github.com/potix/log_monitor/actorplugger"
)

type pathInfo struct {
    actorName string
    actorConfig string
}

type fileStatus struct {
    fileID string
    dirty bool
    actorPlugin actorplugger.ActorPlugin
    mutex *sync.Mutex
}

type renameInfo struct {
    name string
    fileStatus *fileStatus
}

// EventManager is event manager
type EventManager struct{
    loopEnd  chan bool
    watcher *fsnotify.Watcher
    paths map[string]*pathInfo
    pathsMutex *sync.Mutex
    files map[string]*fileStatus
    filesMutex *sync.Mutex
    renameFiles map[string]*renameInfo
    renameFilesMutex *sync.Mutex
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

func (e *EventManager) foundFile(name string, fileID string) {
    actorPlugin, err := e.newActorPlugin(name)
    if err != nil {
	log.Printf("[foundFile] can not create actor plugin (%v, %v): %v", fileID, name, err)
        return
    }
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    _, ok := e.files[name]
    if ok {
        log.Printf("[found File] already exists file (%v, %v)", name, fileID)
        return
    }
    e.files[name] = &fileStatus {
        fileID: fileID,
        dirty: true,
	actorPlugin: actorPlugin,
	mutex : new(sync.Mutex),
    }
    actorPlugin.FoundFile(name, fileID)
}

func (e *EventManager) createdFile(event fsnotify.Event, fileID string) {
    actorPlugin, err := e.newActorPlugin(event.Name)
    if err != nil {
	log.Printf("[createdFile] can not create actor plugin (%v, %v): %v", fileID, event.Name, err)
        return
    }
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[event.Name]
    if ok {
        if status.fileID == fileID {
             log.Printf("[createdFile] already exists file (%v, %v)", fileID, event.Name)
             return
        }
        // rename to exists file name
        status.mutex.Lock()
        defer status.mutex.Unlock()
        status.actorPlugin.RemovedFile(event.Name, status.fileID)
        delete(e.files, event.Name)
    }

    // renamed file
    e.renameFilesMutex.Lock()
    defer e.renameFilesMutex.Unlock()
    renameInfo, ok := e.renameFiles[fileID]
    if ok {
          renameInfo.fileStatus.mutex.Lock()
          defer renameInfo.fileStatus.mutex.Unlock()
          renameInfo.fileStatus.actorPlugin = actorPlugin
          renameInfo.fileStatus.dirty = true
          e.files[event.Name] = renameInfo.fileStatus
          actorPlugin.RenamedFile(renameInfo.name, event.Name, fileID)
          delete(e.renameFiles, fileID)
          return
    }

    // created file
    e.files[event.Name] = &fileStatus {
        fileID: fileID,
        dirty: true,
	actorPlugin: actorPlugin,
	mutex : new(sync.Mutex),
    }
    actorPlugin.CreatedFile(event.Name, fileID)
}

func (e *EventManager) removedFile(event fsnotify.Event) (bool) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[event.Name]
    if !ok {
        log.Printf("[removedFile] not exists file (%v)", event.Name)
        return false
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    status.actorPlugin.RemovedFile(event.Name, status.fileID)
    delete(e.files, event.Name)
    return true
}

func (e *EventManager) renamedFile(event fsnotify.Event) (bool) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[event.Name]
    if !ok {
        log.Printf("[renamedFile] not exists file (%v)", event.Name)
        return false
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    e.renameFilesMutex.Lock()
    defer e.renameFilesMutex.Unlock()
    e.renameFiles[status.fileID] = &renameInfo {
        name: event.Name,
        fileStatus: status,
    }
    delete(e.files, event.Name)
    return true
}

func (e *EventManager) setDirtyFile(event fsnotify.Event, fileID string) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[event.Name]
    if !ok {
        log.Printf("[setDirtyFile] not exists file (%v, %v)", event.Name, fileID)
        return
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    status.dirty = true
}

func (e *EventManager) modifiedFile(event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status , ok := e.files[event.Name]
    if !ok {
        log.Printf("[modifiedFile] not exists file (%v)", event.Name)
        return
    }
    status.mutex.Lock()
    defer status.mutex.Unlock()
    if !status.dirty {
        log.Printf("[modifiedFile] not dirty  (%v, %v)", event.Name, status.fileID)
        return
    }
    status.actorPlugin.ModifiedFile(event.Name, status.fileID)
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
            log.Printf("[eventLoop] event = %v", event)
            if event.Name == "" || event.Op&fsnotify.Chmod == fsnotify.Chmod{
                // nop
                break
            }
            fileID, info, getFileInfoErr := e.getFileInfo(event.Name)
            if event.Op&fsnotify.Create == fsnotify.Create {
               if getFileInfoErr != nil {
                   log.Printf("[event Loop] can not get file info (%v)", event.Name)
                   break
               }
               if info.IsDir() {
                   parent := filepath.Dir(event.Name)
                   info, ok := e.paths[parent]
                   if !ok {
                       log.Printf("[event Loop] not found parent %v", parent) 
                   } else {
                       e.addPath(event.Name, info.actorName, info.actorConfig)
                   }
               } else {
                   e.createdFile(event, fileID)
               }
            }
            if event.Op&fsnotify.Remove == fsnotify.Remove {
                   ok := e.removedFile(event)
                   if !ok {
                       e.deletePath(event.Name)
                   }
            }
            if event.Op&fsnotify.Rename == fsnotify.Rename {
                   ok := e.renamedFile(event)
                   if !ok {
                       e.deletePath(event.Name)
                   }
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
               if getFileInfoErr != nil {
                   log.Printf("[event.Loop] can not get file info (%v)", event.Name)
                   break
               }
               if !info.IsDir() {
                   e.setDirtyFile(event, fileID)
               }
            }
            if event.Op&fsnotify.Remove != fsnotify.Remove && event.Op&fsnotify.Rename != fsnotify.Rename {
               if getFileInfoErr != nil {
                   log.Printf("[event.Loop] can not get file info (%v)", event.Name)
                   break
               }
               if !info.IsDir() {
                   e.modifiedFile(event)
               }
            }
        case err, ok := <-e.watcher.Errors:
            if !ok {
                 // end loop
                 return
            }
            log.Println("[eventLoop] error: ", err)
        }
    }
}

func (e *EventManager) addPath(path string, actorName string, actorConfig string) (error) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
        _, ok := e.paths[path]
        if ok {
            log.Printf("[addPath] already exists path (%v)", path)
            return nil
        }
        err := e.watcher.Add(path)
        if err != nil {
            return errors.Wrap(err, "can not add path to watcher")
	} else {
            e.paths[path] = &pathInfo{
                actorName: actorName,
                actorConfig: actorConfig,
            }
        }
        log.Printf("[addPath] add path (%v)", path)
        return nil
}

func (e *EventManager) deletePath(path string) (error) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
        _, ok := e.paths[path]
        if !ok {
            log.Printf("[deletePath] not exists path (%v)", path)
            return nil
        }
        err := e.watcher.Remove(path)
        if err != nil {
            return errors.Wrap(err, "can not delete path from watcher")
	} else {
            delete(e.paths, path)
        }
        log.Printf("[deletePath] delete path (%v)", path)
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
            return "", nil, errors.Wrapf(err, "[getFileInfo] can not get file info (%v)", filePath)
        }
        stat, ok := info.Sys().(*syscall.Stat_t)
        if !ok {
            return "", nil, errors.Wrapf(err, "[getFileInfo] can not get file stat (%v)", filePath)
        }
        fileID := fmt.Sprintf("%v:%v", stat.Dev, stat.Ino)
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

func (e *EventManager) addTargets(targetPath string, actorName string, actorConfig string) {
    targetPath = e.fixupPath(targetPath)
    fileList, err := ioutil.ReadDir(targetPath)
    if err != nil {
        log.Printf("[addTargets] can not read dir (%v): %v", targetPath, err)
        return
    }
    err = e.addPath(targetPath, actorName, actorConfig)
    if err != nil {
        log.Printf("[addTargets] can not add path (%v): %v", targetPath, err)
        return
    }
    for _, file := range fileList {
        newPath := filepath.Join(targetPath, file.Name())
        if (targetPath == ".") {
            newPath = targetPath + "/" + newPath
        }
        if file.IsDir() {
            e.addTargets(newPath, actorName, actorConfig)
	    continue
        }
	fileID, _, err := e.getFileInfo(newPath)
        if err != nil {
            log.Printf("[addTargets] can not get file info (%v)", newPath)
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
        renameFiles : make(map[string]*renameInfo),
        renameFilesMutex : new(sync.Mutex),
    }
    for _, targetInfo := range config.Targets {
         eventManager.addTargets(targetInfo.Path, targetInfo.ActorName, targetInfo.ActorConfig)
    }
    return eventManager, nil
}
