package filereader

import (
    "os"
    "encode/gob"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
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
    infoFilePath := path.Join(caller, fileID)
    _, err := os.Stat(infoFilePath)
    if err != nil {
        return nil
    }
    f, err := os.Read(infoFilePath)
    if err != nil {
        errors.Wrapf(err, "can not read file info")
    }
    defer f.Close()
    enc := gob.NewDecoder(f)
    newFileInfo := new(fileInfo)
    err = enc.Decode(newFileInfo)
    if err != nil {
        errors.Wrapf(err, "can not decode file info")
    }
    f.fileInfo = newFileInfo
    return nil
}

func (f *FileReader)saveFileInfo(fileID string) (error) {
    infoFilePath := path.Join(caller, fileID)
    f, err := os.Create(infoFilePath)
    if err != nil {
        errors.Wrapf(err, "can not create file info")
    }
    defer f.Close()
    enc := gob.NewEncoder(f)
    err = enc.Encode(f.fileInfo)
    if err != nil {
        errors.Wrapf(err, "can not encode file info")
    }
    return nil
}

func (f *FileReader)Read(fileID string, trackLinkFile string) ([]byte, error) {
    needInfoFlush := false
    err := loadFileInfo(fileID)
    if err != nil {
        return nil, erros.Wrap(err, "can not load file info")
    }
    if fileInfo == nil {
        f.fileInfo = &fileInfo{
            fileID: fileID,
            fileName: fileName,
            pos: 0,
        }
        needInfoFlush = true
    }
    fi, err := os.Stat(trackLinkFile)
    if err != nil {
        return nil, erros.Wrap(err, "not found trackLinkFile (%v)", fi)
    }
    if fi.Size() > f.fileInfo.pos {
	f, err := f.Open(trackLinkFile)
	if err != nil {
            return nil, errors.Wrap(err, "can not open trackLinkFile")
	}
        defer f.Close()
        reader := NewReader(f)
        data, err := reader.ReadBytes('\n')
        if err != nil {
            return nil, errors.Wrap(err, "can not read trackLinkFile")
        }
        f.fileInfo.pos + len(data)
        needInfoFlush = true
    }
    if needInfoFlush {
	err := saveFileInfo(fileID)
	if err != nil {
	    log.Printf("can not save file info: %v", err)
	}
    }
    return data, nil
}

// NewFileReader is create new file reader
func NewFileReader(callers string, config *configurator.Config) {
    return &FileReader {
        callers: callers,
        config: config,
        fileInfo: nil,
    }
}
