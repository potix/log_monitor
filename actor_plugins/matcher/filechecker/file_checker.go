package filechecker

import (
    "os"
    "log"
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

// Check is checl
func (f *FileChecker)Check(fileID string, trackLinkFile string) (error) {
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
    if fi.Size() > f.fileInfo.pos {
	file, err := os.Open(trackLinkFile)
	if err != nil {
            return errors.Wrapf(err, "can not open trackLinkFile (%v)", trackLinkFile)
	}
        defer file.Close()
        reader := bufio.NewReader(file)
        data, err := reader.ReadBytes('\n')
        if err != nil {
            return errors.Wrapf(err, "can not read trackLinkFile (%v)", trackLinkFile)
        }
        // XXXX check
        f.fileInfo.pos += int64(len(data))
        err := f.saveFileInfo(f.fileInfo.fileID)
        if err != nil {
            log.Printf("can not save file info: %v", err)
        }
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
