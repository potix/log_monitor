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
)

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

type flushInfo struct {
    finish chan bool
}

// Sender is matcher
type Sender struct {
   fileReader *fileReader
   fileCheckInfo *fileCheckInfo
   flushInfo *flushInfo
   config *configurator.Config
}

func (m *Matcher) fileCheckLoop() {
    for {
        m.fileCheckInfo.kickEvent.Acquire(context.Background(), 1)
        if m.fileCheckInfo.getFinish() {
            return
        }
        if !m.fileCheckInfo.needCheck() {
            return
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

func (m *Matcher) flushLoop() {
    for {
        select {
        case <-time.After(m.config.FlushInterval*time.Second):
            m.fileCheckInfo.kickEvent.Release(1)
        case <-m.flushInfo.finish:
            return
        }
    }
}

func (m *Matcher) initialize(fileName string, fileID string) {
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
    m.flushInfo = &flushInfo{
        finish: make(chan bool),
    }
    go m.fileCheckLoop()
    if m.config.FlushInterval != 0 {
         go m.flushLoop()
    }
}

func (m *Matcher) finalize(fileName string, fileID string) {
    if m.targetInfo == nil {
        return
    }
    close(m.flushInfo.finish)
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
    if m.config.FlushInterval == 0 {
        m.fileCheckInfo.kickEvent.Release(1)
    }
}

// NewSender is create new matcher
func NewSender(configFile string) (actorplugger.ActorPlugin, error) {
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

    return &Matcher {
        fileReader: fileReader,
        config: config,
        targetInfo: nil,
        fileCheckInfo: nil,
        flushInfo: nil,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "sender", NewSender
}
