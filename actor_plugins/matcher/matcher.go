package main

import (
    "log"
    "time"
    "sync"
    "sync/atomic"
    "context"
    "golang.org/x/sync/semaphore"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actorplugger"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
    "github.com/potix/log_monitor/actor_plugins/matcher/filechecker"
    "github.com/potix/log_monitor/actor_plugins/matcher/rulemanager"
)

type targetInfo struct {
    fileNameMutex *sync.Mutex
    fileName string
    fileID string
}

func (t *targetInfo) getFileName() (string) {
    t.fileNameMutex.Lock()
    defer t.fileNameMutex.Unlock()
    return t.fileName
}
func (t *targetInfo) setFileName(fileName string) {
    t.fileNameMutex.Lock()
    defer t.fileNameMutex.Unlock()
    t.fileName = fileName
}

func (t *targetInfo) getFileID() (string) {
    return t.fileID
}

type fileCheckInfo struct {
    kickEvent *semaphore.Weighted
    needCheck uint32
    finish  uint32
}

func (f *fileCheckInfo) getNeedCheck() bool {
    return atomic.CompareAndSwapUint32(&f.needCheck, 1, 0)
}

func (f *fileCheckInfo) setNeedCheck() {
    atomic.StoreUint32(&f.needCheck, 1)
}

func (f *fileCheckInfo) getFinish() bool {
    return atomic.LoadUint32(&f.finish) == 1
}

func (f *fileCheckInfo) setFinish() {
    atomic.StoreUint32(&f.finish, 1)
}

type expireCheckInfo struct {
    expire int64
    finish chan bool
}

func (e *expireCheckInfo) getExpire() (int64) {
    return atomic.LoadInt64(&e.expire)
}

func (e *expireCheckInfo) updateExpire(expire int64) {
    atomic.StoreInt64(&e.expire, expire)
}

// Matcher is matcher
type Matcher struct {
    configurator *configurator.Configurator
    ruleManager *rulemanager.RuleManager
    fileChecker *filechecker.FileChecker
    targetInfo *targetInfo
    fileCheckInfo *fileCheckInfo
    expireCheckInfo *expireCheckInfo
}

func (m *Matcher) fileCheckLoop() {
    for {
        m.fileCheckInfo.kickEvent.Acquire(context.Background(), 1)
        if m.fileCheckInfo.getFinish() {
            return
        }
        if !m.fileCheckInfo.getNeedCheck() {
            continue
        }
        fileID := m.targetInfo.getFileID()
        fileName := m.targetInfo.getFileName()
        rule := m.ruleManager.getRule(fileName)
        if rule == nil {
            log.Printf("not found rule for target (%v)", fileName)
            continue
        }
        err := m.fileChecker.Check(fileID, fileName, rule)
        if err != nil {
            log.Printf("can not check file (%v:%v)", )
        }
    }
}

func (m *Matcher) expireCheckLoop() {
    for {
        select {
        case <-time.After(1*time.Second):
            t := time.Now().Unix()
            if t >= m.expireCheckInfo.expire {
                fileID := m.targetInfo.getFileID()
                m.fileChecker.Flush(fileID)
            }
        case <-m.expireCheckInfo.finish:
            return
        }
    }
}

func (m *Matcher) initialize(fileName string, fileID string) {
    rule := m.ruleManager.getRule(fileName)
    if rule == nil {
        log.Printf("not found rule for target (%v)", fileName)
        return
    } 
    m.targetInfo = &targetInfo{
        fileNameMutex: new(sync.Mutex),
        fileName: fileName,
        fileID: fileID,
    }   
    m.fileCheckInfo = &fileCheckInfo{
        kickEvent: semaphore.NewWeighted(0),
        needCheck: 0,
        finish: 0,
    }   
    m.expireCheckInfo = &expireCheckInfo{
        expire: rule.expire + time.Now().Unix(),
        finish: make(chan bool),
    }
    go m.fileCheckLoop()   
    go m.expireCheckLoop()   
}

func (m *Matcher) finalize(fileName string, fileID string) {
    if m.targetInfo == nil {
        return
    }
    close(m.expireCheckInfo.finish)
    m.fileCheckInfo.setFinish()
    m.fileCheckInfo.kickEvent.Release(1)
}

// FoundFile is add file
func (m *Matcher) FoundFile(fileName string, fileID string) {
    m.initialize(fileName, fileID)
}

// CreatedFile is add file
func (m *Matcher) CreatedFile(fileName string, fileID string) {
    m.initialize(fileName, fileID)
}

// RemovedFile is remove file
func (m *Matcher) RemovedFile(fileName string, fileID string) {
    m.finalize(fileName, fileID)
}

// RenamedFile is rename file
func (m *Matcher) RenamedFile(oldFileName string, newFileName string, fileID string) {
   m.targetInfo.setFileName(newFileName)
}

// ModifiedFile is modify file
func (m *Matcher) ModifiedFile(fileName string, fileID string) {
    m.fileCheckInfo.kickEvent.Release(1)
}

// NewMatcher is create new matcher
func NewMatcher(configFile string) (actorplugger.ActorPlugin, error) {
    log.Printf("configFile = %v", configFile)

    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }

    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }

    log.Printf("config = %v", config)

    ruleManager, err := rulemanager.NewRuleManager(configurator)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create rule manager")
    }

    fileChecker, err := filechecker.NewFileChecker(config)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create file checker")
    }

    return &Matcher {
        configurator: configurator,
        fileChecker: fileChecker,
        ruleManager: ruleManager,
        targetInfo: nil,
        fileCheckInfo: nil,
        expireCheckInfo: nil,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "matcher", NewMatcher
}
