package event_manager

import (
    "os"
    "syscall"
    "fmt"
    "log"
    "sync"
    "sync/atomic"
    "github.com/pkg/errors"
    "github.com/fsnotify/fsnotify"
)

type fileStatus struct {
    name string
    dirty uint64
    pos uint64
}

type eventManager struct{
    loopEnd  chan bool
    watcher *fsnotify.Watcher
    paths map[string]bool
    pathsMutex *sync.Mutex
    files map[string]*fileStatus
    filesMutex *sync.Mutex
}

func (e *eventManager) addFile(fileId string, event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    oldStatus, ok := e.files[fileId]
    if ok {
        if oldStatus.name != event.Name {
            log.Printf("change name (%v, %v -> %v)", fileId, oldStatus.name, event.Name)
            oldStatus.name = event.Name
        } else {
            log.Printf("already exists file (%v, %v)", fileId, event.Name)
        }
        return
    }
    e.files[fileId] = &fileStatus {
        name: event.Name,
        dirty: 1,
        pos: 0,
    }
}

func (e *eventManager) removeFile(fileId string, event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    _, ok := e.files[fileId]
    if !ok {
        log.Printf("not exists file (%v, %v)", fileId, event.Name)
        return
    }
    delete(e.files, fileId)
}

func (e *eventManager) setDirtyFile(fileId string, event fsnotify.Event) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[fileId]
    if !ok {
        log.Printf("not exists file (%v, %v)", fileId, event.Name)
        return
    }
    atomic.StoreUint64(&status.dirty, 1)
}

func (e *eventManager) checkFileContent(fileId string) {
    e.filesMutex.Lock()
    defer e.filesMutex.Unlock()
    status, ok := e.files[fileId]
    if !ok {
        log.Printf("not exists file (%v)", fileId)
        return
    }
    dirty := atomic.LoadUint64(&status.dirty)
    if dirty == 0 {
        log.Printf("not dirty  (%v, %v)", fileId, status.name)
        return
    }


    // XXX TODO check


    atomic.StoreUint64(&status.dirty, 0)
}

func (e *eventManager) eventLoop() {
    for {
        select {
        case <- e.loopEnd:
            return
        case event, ok := <- e.watcher.Events:
            if !ok {
                 // end loop
                 return
            }
            log.Printf("%v", event)
            if event.Op&fsnotify.Chmod == fsnotify.Chmod {
                // nop
                break
            }
            info, err := os.Stat(event.Name)
            if err != nil {
                log.Printf("can not get file info (%v)", event.Name)
                break
            }
            stat, ok := info.Sys().(*syscall.Stat_t)
            if !ok {
                log.Printf("can not get file stat (%v)", event.Name)
                break
            }
            fileId := fmt.Sprintf("%v:%v", stat.Dev, stat.Ino)
            log.Println("fileId:", fileId)
            if event.Op&fsnotify.Create == fsnotify.Create {
               if info.IsDir() {
                   e.AddPath(event.Name)
               } else {
                   e.addFile(fileId, event)
               }
            }
            if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
               if info.IsDir() {
                   e.RemovePath(event.Name)
               } else {
                   e.removeFile(fileId, event)
               }
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
               if !info.IsDir() {
                   e.setDirtyFile(fileId, event)
               }
            }
            if !info.IsDir() {
                e.checkFileContent(fileId)
            }
        case err, ok := <- e.watcher.Errors:
            if !ok {
                 // end loop
                 return
            }
            log.Println("error: ", err)       
        }
    }
}

func (e *eventManager) AddPath(path string) (error) {
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
            e.paths[path] = true
        }
        return nil
}
func (e *eventManager) RemovePath(path string) (error) {
	e.pathsMutex.Lock()
        defer e.pathsMutex.Unlock()
        _, ok := e.paths[path]
        if !ok {
            log.Printf("not exists path (%v)", path)
            return nil;
        }
        err := e.watcher.Remove(path)
        if err != nil {
            errors.Wrap(err, "can not remove path from watcher")
	} else {
            delete(e.paths, path)
        }
        return nil
}

func (e *eventManager) Start() {
     e.ruleManager.Start()
     e.loopEnd = make(chan bool)
     go e.eventLoop()
}

func (e *eventManager) Stop() {
     close(e.loopEnd)
     e.ruleManager.Stop()
}

func (e *eventManager) Clean() {
     e.watcher.Close()
     e.ruleManager.Clean()
}

func NewEventManager(configurator *configurator.Configurator) (*eventManager, error) {
     ruleManager, err := ruleManager.NewRuleManager(configurator)
     if (err != nil) {
         return nil, errors.Wrapf("can not create rule manager")
     }

     fileChecker, err := fileChecker.NewFileChecker()
     if (err != nil) {
         return nil, errors.Wrapf(err, "can not create file checker")
     }

     watcher, err :=  fsnotify.NewWatcher()
     if err != nil {
         return nil, errors.Wrapf(err, "can not create event manager")
     }

     return &eventManager {
          fileChecker: fileChecker,
          ruleManager: ruleManager,
          loopEnd: make(chan bool),
	  watcher : watcher,
	  paths : make(map[string]bool),
          pathsMutex : new(sync.Mutex),  
	  files : make(map[string]*fileStatus),
	  filesMutex : new(sync.Mutex),
     }, nil
}

