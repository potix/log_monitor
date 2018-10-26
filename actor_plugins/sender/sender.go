package main

import (
    "os"
    "log"
    "time"
    "sync"
    "sync/atomic"
    "context"
    "google.golang.org/grpc"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actorplugger"
    "github.com/potix/log_monitor/actor_plugins/sender/filereader"
    "github.com/potix/log_monitor/actor_plugins/sender/configurator"
    logpb "github.com/potix/log_monitor/logpb"
)

type targetInfo struct {
    fileNameMutex *sync.Mutex
    fileName string
    fileID string
    trackLinkFile string
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

func (t *targetInfo) getTrackLinkFile() (string) {
    return t.trackLinkFile
}

type fileCheckInfo struct {
    eventCh chan bool
    needCheck uint32
}

func (f *fileCheckInfo) getNeedCheck() bool {
    return atomic.CompareAndSwapUint32(&f.needCheck, 1, 0)
}

func (f *fileCheckInfo) setNeedCheck() {
    atomic.StoreUint32(&f.needCheck, 1)
}

// Sender is sender
type Sender struct {
   callers string
   fileReader *filereader.FileReader
   config *configurator.Config
   targetInfo *targetInfo
   fileCheckInfo *fileCheckInfo
   hostname string
}



func (s *Sender) fileCheckLoop() {
    for {
        select {
        case _, ok := <-s.fileCheckInfo.eventCh:
            if !ok {
               return
            }
        case <-time.After(time.Duration(s.config.FlushInterval) * time.Second):
        }
        fileID := s.targetInfo.getFileID()
        fileName := s.targetInfo.getFileName()
	trackLinkFile := s.targetInfo.getTrackLinkFile()
	conn, err := grpc.Dial(s.config.AddrPort,  grpc.WithInsecure())
	if err != nil {
            log.Printf("can not dial: %v", err)
            continue
	}
	client := logpb.NewLogClient(conn)
        defer conn.Close()
again:
        data, eof, err := s.fileReader.Read(fileID, trackLinkFile)
        if err != nil {
            log.Printf("can not check file (%v, %v): %v", fileID, trackLinkFile, err)
            continue
        }
	transferRequest := &logpb.TransferRequest {
		Label: s.config.Label,
		Host: s.hostname,
		Path: fileName,
		LogData: data,
	}
	transferReply, err := client.Transfer(context.Background(), transferRequest)
	if err != nil {
            log.Printf("can not recieve reply : %v", err)
            continue
	}
        if !transferReply.Success  {
            log.Printf("can not transfer : %v", transferReply.Msg)
            continue
        }
        s.fileReader.UpdatePosition(len(data))
        if !eof {
            goto again
        }

    }
}

func (s *Sender) initialize(fileName string, fileID string, trackLinkFile string) {
    s.targetInfo = &targetInfo{
        fileNameMutex: new(sync.Mutex),
        fileName: fileName,
        fileID: fileID,
        trackLinkFile: trackLinkFile,
    }
    s.fileCheckInfo = &fileCheckInfo{
        eventCh: make(chan bool),
        needCheck: 0,
    }
    go s.fileCheckLoop()
}

func (s *Sender) finalize(fileName string, fileID string, trackLinkFile string) {
    if s.targetInfo == nil {
        return
    }
    close(s.fileCheckInfo.eventCh)
}

// FoundFile is add file
func (s *Sender) FoundFile(fileName string, fileID string, trackLinkFile string) {
    s.initialize(fileName, fileID, trackLinkFile)
}

// CreatedFile is add file
func (s *Sender) CreatedFile(fileName string, fileID string, trackLinkFile string) {
    s.initialize(fileName, fileID, trackLinkFile)
}

// RemovedFile is remove file
func (s *Sender) RemovedFile(fileName string, fileID string, trackLinkFile string) {
    s.finalize(fileName, fileID, trackLinkFile)
}

// RenamedFile is rename file
func (s *Sender) RenamedFile(oldFileName string, newFileName string, fileID string) {
    s.targetInfo.setFileName(newFileName)
}

// ModifiedFile is modify file
func (s *Sender) ModifiedFile(fileName string, fileID string) {
    s.fileCheckInfo.setNeedCheck()
    if s.config.FlushInterval == 0 {
        s.fileCheckInfo.eventCh <- true
    }
}

// NewSender is create new matcher
func NewSender(callers string, configFile string) (actorplugger.ActorPlugin, error) {
    log.Printf("configFile = %v", configFile)
    hostname, err := os.Hostname()
    if err != nil {
	return nil, errors.Wrap(err, "can not get hostname")
    }
    configurator, err := configurator.NewConfigurator(configFile)
    if err != nil {
        return nil, errors.Wrapf(err, "can not create configurator (%v)", configFile)
    }

    config, err := configurator.Load()
    if err != nil {
        return nil, errors.Wrapf(err, "can not load config (%v)", configFile)
    }
    log.Printf("config = %v", config)
    newCallers := callers + ".sender"
    fileReader := filereader.NewFileReader(newCallers, config)
    return &Sender {
        callers: newCallers,
        fileReader: fileReader,
        config: config,
        targetInfo: nil,
        fileCheckInfo: nil,
        hostname: hostname,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "sender", NewSender
}
