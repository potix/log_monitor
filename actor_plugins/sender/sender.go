package main

import (
    "os"
    "log"
    "time"
    "sync"
    "sync/atomic"
    "context"
    "golang.org/x/sync/semaphore"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actorplugger"
    "github.com/potix/log_monitor/actor_plugins/sender/filereader"
    "github.com/potix/log_monitor/actor_plugins/sender/configurator"
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
   callers string
   fileReader *filereader.FileReader
   config *configurator.Config
   targetInfo *targetInfo
   fileCheckInfo *fileCheckInfo
   flushInfo *flushInfo
   hostname string
}

func (s *Sender) fileCheckLoop() {
    for {
        s.fileCheckInfo.kickEvent.Acquire(context.Background(), 1)
        if s.fileCheckInfo.getFinish() {
            return
        }
        if !s.fileCheckInfo.getNeedCheck() {
            return
        }
        fileID := s.targetInfo.getFileID()
        fileName := s.targetInfo.getFileName()
        data, err := s.fileReader.Read(fileID, s.targetInfo.getTrackLinkFile())
        if err != nil {
            log.Printf("can not check file (%v:%v)", )
        }
        // XXXXXXXXXXXX serd grpc
	transferRequest := logpb.TransferRequest {
		Label: s.config.Label,
		Host: s.hostname,
		Path: fileName,
		LogData: data, 
	}

	conn, err := grpc.Dial("127.0.0.1:19003", nil)
	client := logpb.NewLogClient(conn) LogClient {
	client.Transfer(context.BackGround(), transferRequest, nil) (*TransferReply, error) {
			        out := new(TransferReply)
				        err := c.cc.Invoke(ctx, "/Log/Transfer", in, out, opts...)
					        if err != nil {
							                return nil, err
									        }
										        return out, nil
										}

										o    conn, err := grpc.Dial("127.0.0.1:19003", grpc.WithInsecure())
										    if err != nil {
											            log.Fatal("client connection error:", err)
												        }
													    defer conn.Close()
													        client := pb.NewCatClient(conn)
														    message := &pb.GetMyCatMessage{"tama"}
														        res, err := client.GetMyCat(context.TODO(), message)


    }
}

func (s *Sender) flushLoop() {
    for {
        select {
        case <-time.After(time.Duration(s.config.FlushInterval) * time.Second):
            s.fileCheckInfo.kickEvent.Release(1)
        case <-s.flushInfo.finish:
            return
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
        kickEvent: semaphore.NewWeighted(0),
        needCheck: 0,
        finish: 0,
    }
    s.flushInfo = &flushInfo{
        finish: make(chan bool),
    }
    go s.fileCheckLoop()
    if s.config.FlushInterval != 0 {
         go s.flushLoop()
    }
}

func (s *Sender) finalize(fileName string, fileID string, trackLinkFile string) {
    if s.targetInfo == nil {
        return
    }
    close(s.flushInfo.finish)
    s.fileCheckInfo.setFinish()
    s.fileCheckInfo.kickEvent.Release(1)
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
        s.fileCheckInfo.kickEvent.Release(1)
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
        flushInfo: nil,
        hostname: hostname,
    }, nil
}

// GetActorPluginInfo is GetActorPluginInfo
func GetActorPluginInfo() (string, actorplugger.ActorPluginNewFunc) {
    return "sender", NewSender
}
