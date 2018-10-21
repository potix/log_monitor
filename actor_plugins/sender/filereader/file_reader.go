package filereader

import (
    "os"
    "log"
    "bufio"
    "path/filepath"
    "encoding/gob"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actor_plugins/sender/configurator"
)

type fileInfo struct {
    fileID string
    trackLinkFile string
    pos int64
}

// FileReader is FileReader
type FileReader struct {
    callers string
    config *configurator.Config
    fileInfo *fileInfo
}

func (f * FileReader)loadFileInfo(fileID string) (error) {
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

func (f *FileReader)saveFileInfo(fileID string) (error) {
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

func (f *FileReader)Read(fileID string, trackLinkFile string) ([]byte, error) {
    if f.fileInfo == nil {
        err := f.loadFileInfo(fileID)
        if err != nil {
            return nil, errors.Wrapf(err, "can not load file info (%v)", fileID)
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
        return nil, errors.Wrapf(err, "not found trackLinkFile (%v)", trackLinkFile)
    }
    if fi.Size() > f.fileInfo.pos {
	file, err := os.Open(trackLinkFile)
	if err != nil {
            return nil, errors.Wrapf(err, "can not open trackLinkFile (%v)", trackLinkFile)
	}
        defer file.Close()
        reader := bufio.NewReader(file)
        data, err := reader.ReadBytes('\n')
        if err != nil {
            return nil, errors.Wrapf(err, "can not read trackLinkFile (%v)", trackLinkFile)
        }
        f.fileInfo.pos += int64(len(data))
	err = f.saveFileInfo(fileID)
	if err != nil {
	    log.Printf("can not save file info: %v", err)
	}
        return data, nil
    }
    return nil, nil
}

// NewFileReader is create new file reader
func NewFileReader(callers string, config *configurator.Config) (*FileReader) {
    return &FileReader {
        callers: callers,
        config: config,
        fileInfo: nil,
    }
}
