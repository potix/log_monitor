package main

import (
    "log"
    "sync"
    "sync/atomic"
    "context"
    "path"
    "path/filepath"
    "golang.org/x/sync/semaphore"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actorplugger"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
    "github.com/potix/log_monitor/actor_plugins/matcher/filechecker"
    "github.com/potix/log_monitor/actor_plugins/matcher/rulemanager"
    "github.com/potix/log_monitor/actor_plugins/matcher/notifierplugger"
)

type targetInfo struct {
    fileNameMutex *sync.Mutex
    fileName string
    fileID string
    trackLinkFilePath string
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

func (t *targetInfo) getTrackLinkFilePath() (string) {
    return t.trackLinkFilePath
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

// Matcher is matcher
type Matcher struct {
    callers string
    configurator *configurator.Configurator
    ruleManager *rulemanager.RuleManager
    fileChecker *filechecker.FileChecker
    targetInfo *targetInfo
    fileCheckInfo *fileCheckInfo
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
        trackLinkFilePath := m.targetInfo.getTrackLinkFilePath()
	pathMatchers := m.ruleManager.GetRule(fileName)
        if pathMatchers == nil {
            log.Printf("not found rule for target (%v)", fileName)
            continue
        }
        err := m.fileChecker.Check(fileID, trackLinkFilePath, fileName, pathMatchers)
        if err != nil {
	    log.Printf("can not check file (%v, %v, %v): %v", fileID, trackLinkFilePath, fileName, err)
        }
    }
}

func (m *Matcher) initialize(fileName string, fileID string, trackLinkFilePath string) {
    pathMatchers := m.ruleManager.GetRule(fileName)
    if pathMatchers == nil {
        log.Printf("not found rule for target (%v)", fileName)
        return
    }
    m.targetInfo = &targetInfo{
        fileNameMutex: new(sync.Mutex),
        fileName: fileName,
        fileID: fileID,
        trackLinkFilePath: trackLinkFilePath,
    }
    m.fileCheckInfo = &fileCheckInfo{
        kickEvent: semaphore.NewWeighted(0),
        needCheck: 0,
        finish: 0,
    }
    m.ruleManager.Start()
    go m.fileCheckLoop()
}

func (m *Matcher) finalize(fileName string, fileID string, trackLinkFilePath string) {
    if m.targetInfo == nil {
        return
    }
    m.fileCheckInfo.setFinish()
    m.fileCheckInfo.kickEvent.Release(1)
    m.ruleManager.Stop()
}

// FoundFile is add file
func (m *Matcher) FoundFile(fileName string, fileID string, trackLinkFilePath string) {
    m.initialize(fileName, fileID, trackLinkFilePath)
}

// CreatedFile is add file
func (m *Matcher) CreatedFile(fileName string, fileID string, trackLinkFilePath string) {
    m.initialize(fileName, fileID, trackLinkFilePath)
}

// RemovedFile is remove file
func (m *Matcher) RemovedFile(fileName string, fileID string, trackLinkFilePath string) {
    m.finalize(fileName, fileID, trackLinkFilePath)
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
func NewMatcher(callers string, configFile string) (actorplugger.ActorPlugin, error) {
    log.Printf("configFile = %v", configFile)
    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }
    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }
    pluginPath := path.Join(filepath.Dir(configFile), config.NotifierPluginPath)
    err = notifierplugger.LoadNotifierPlugins(pluginPath)
    if err != nil {
        log.Fatalf("can not load notifier plugins (%v): %v", pluginPath, err)
    }
    log.Printf("config = %v", config)
    newCallers := callers + ".matcher"
    ruleManager, err := rulemanager.NewRuleManager(config, configurator)
    if (err != nil) {
        return nil, errors.Wrapf(err, "can not create rule manager")
    }
    return &Matcher {
        callers: newCallers,
        configurator: configurator,
        fileChecker: filechecker.NewFileChecker(newCallers, config),
        ruleManager: ruleManager,
        targetInfo: nil,
        fileCheckInfo: nil,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "matcher", NewMatcher
}
