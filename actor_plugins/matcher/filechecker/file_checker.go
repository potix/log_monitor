package filechecker

import (
    "os"
    "log"
    "regexp"
    "bufio"
    "path/filepath"
    "encoding/gob"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
)

type fileInfo struct {
    fileID string
    trackLinkFile string
    pos int64
}

// FileChecker is FileChecker
type FileChecker struct {
    callers string
    config *configurator.Config
    fileInfo *fileInfo
}

func (f * FileChecker)loadFileInfo(fileID string) (error) {
    infoFilePath := filepath.Join(f.callers, fileID)
    _, err := os.Stat(infoFilePath)
    if err != nil {
        return nil
    }
    file, err := os.Open(infoFilePath)
    if err != nil {
        errors.Wrapf(err, "can not read file info (%v)", infoFilePath)
    }
    defer file.Close()
    enc := gob.NewDecoder(file)
    newFileInfo := new(fileInfo)
    err = enc.Decode(newFileInfo)
    if err != nil {
        errors.Wrapf(err, "can not decode file info (%v)", infoFilePath)
    }
    f.fileInfo = newFileInfo
    return nil
}

func (f *FileChecker)saveFileInfo(fileID string) (error) {
    infoFilePath := filepath.Join(f.callers, fileID)
    file, err := os.Create(infoFilePath)
    if err != nil {
        errors.Wrapf(err, "can not create file info (%v)", infoFilePath)
    }
    defer file.Close()
    enc := gob.NewEncoder(file)
    err = enc.Encode(f.fileInfo)
    if err != nil {
        errors.Wrapf(err, "can not encode file info (%v)", infoFilePath)
    }
    return nil
}

// Check is check
func (f *FileChecker)Check(fileID string, trackLinkFile string, pathMatcher *configurator.PathMatcher) (error) {
    if f.fileInfo == nil {
        err := f.loadFileInfo(fileID)
        if err != nil {
            return errors.Wrapf(err, "can not load file info (%v)", fileID)
        }
        f.fileInfo = &fileInfo{
            fileID: fileID,
            trackLinkFile: trackLinkFile,
            pos: 0,
        }
	err = f.saveFileInfo(fileID)
	if err != nil {
	    log.Printf("can not save file info: %v", err)
	}
    }
    fi, err := os.Stat(trackLinkFile)
    if err != nil {
        return errors.Wrapf(err, "not found trackLinkFile (%v)", trackLinkFile)
    }
    if fi.Size() <= f.fileInfo.pos {
        return nil
    }
    oldPos := f.fileInfo.pos
    file, err := os.Open(trackLinkFile)
    if err != nil {
        return errors.Wrapf(err, "can not open trackLinkFile (%v)", trackLinkFile)
    }
    defer file.Close()
    _, err = file.Seek(f.fileInfo.pos, 0)
    if err != nil {
        return errors.Wrapf(err, "can not seek trackLinkFile (%v)", trackLinkFile)
    }
    reader := bufio.NewReader(file)
    for {
        data, err := reader.ReadBytes('\n')
        if err != nil {
            log.Printf("can not read bytes (%v:%v)", trackLinkFile, f.fileInfo.pos)
            break
        }
        for _, matcher := range pathMatcher.MsgMatchers {
            matched, err := regexp.Match(matcher.Pattern, data)
            if err != nil {
                log.Printf("can not macth message (%v, %v): %v", matcher.Pattern, string(data), err)
            }
            if matched {
                // XXX TODO notify
                log.Printf("Notify!! %v", string(data))
            }
        }
        f.fileInfo.pos += int64(len(data))
    }
    if oldPos == f.fileInfo.pos {
        return nil
    }
    err = f.saveFileInfo(f.fileInfo.fileID)
    if err != nil {
        log.Printf("can not save file info: %v", err)
    }
    return nil
}

// NewFileChecker is create new file reader
func NewFileChecker(callers string, config *configurator.Config) (*FileChecker) {
    return &FileChecker {
        callers: callers,
        config: config,
        fileInfo: nil,
    }
}
