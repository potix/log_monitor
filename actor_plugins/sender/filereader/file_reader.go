package filereader

import (
    "os"
    "log"
    "bytes"
    "io"
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

func (f *FileReader)Read(fileID string, trackLinkFile string) ([]byte, bool, error) {
    if f.fileInfo == nil {
        err := f.loadFileInfo(fileID)
        if err != nil {
            return nil, false, errors.Wrapf(err, "can not load file info (%v)", fileID)
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
        return nil, false, errors.Wrapf(err, "not found trackLinkFile (%v)", trackLinkFile)
    }
    if fi.Size() <= f.fileInfo.pos {
        return nil, false, nil
    }

    file, err := os.Open(trackLinkFile)
    if err != nil {
        return nil, false, errors.Wrapf(err, "can not open trackLinkFile (%v)", trackLinkFile)
    }
    defer file.Close()
    _, err = file.Seek(f.fileInfo.pos, 0)
    if err != nil {
        return nil, false, errors.Wrapf(err, "can not seek trackLinkFile (%v)", trackLinkFile)
    }
    data := make([]byte, 0, 1048576)
    dataBuffer := bytes.NewBuffer(data)
    reader := bufio.NewReader(file)
    eof := false
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            if err == io.EOF {
                eof = true
            } else {
                log.Printf("can not read trackLinkFile (%v): %v", trackLinkFile, err)
            }
            break
        }
        _, err = dataBuffer.Write(line)
        if err != nil {
            log.Printf("can not read trackLinkFile (%v): %v", trackLinkFile, err)
            return nil, eof, errors.Wrap(err, "can not write to buffer")
        }
        if dataBuffer.Len() > 1048576 {
            break
        }
    }
    return dataBuffer.Bytes(), eof, nil
    
}

// UpdatePosition is update file position
func (f *FileReader)UpdatePosition(readLen int) {
    if f.fileInfo == nil {
        return
    }
    f.fileInfo.pos += int64(readLen)
    err := f.saveFileInfo(f.fileInfo.fileID)
    if err != nil {
        log.Printf("can not save file info: %v", err)
    }
}


// NewFileReader is create new file reader
func NewFileReader(callers string, config *configurator.Config) (*FileReader) {
    return &FileReader {
        callers: callers,
        config: config,
        fileInfo: nil,
    }
}
