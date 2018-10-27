package filechecker

import (
    "os"
    "log"
    "io"
    "regexp"
    "bufio"
    "path"
    "path/filepath"
    "encoding/gob"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/actor_plugins/matcher/configurator"
    "github.com/potix/log_monitor/actor_plugins/matcher/notifierplugger"
)

type fileInfo struct {
    FileID string
    TrackLinkFile string
    Pos int64
}

// FileChecker is FileChecker
type FileChecker struct {
    callers string
    config *configurator.Config
    fileInfo *fileInfo
}

func (f * FileChecker)loadFileInfo(fileID string) (error) {
    infoFilePath := filepath.Join(f.config.SavePrefix, f.callers, fileID)
    _, err := os.Stat(infoFilePath)
    if err != nil {
        return nil
    }
    file, err := os.Open(infoFilePath)
    if err != nil {
        return errors.Wrapf(err, "can not read file info (%v)", infoFilePath)
    }
    defer file.Close()
    enc := gob.NewDecoder(file)
    newFileInfo := new(fileInfo)
    err = enc.Decode(newFileInfo)
    if err != nil {
        //os.Remove(infoFilePath)
        return errors.Wrapf(err, "can not decode file info (%v)", infoFilePath)
    }
    f.fileInfo = newFileInfo
    return nil
}

func (f *FileChecker)saveFileInfo(fileID string) (error) {
    infoFileDir := filepath.Join(f.config.SavePrefix, f.callers)
    _, err := os.Stat(infoFileDir)
    if err != nil {
       err := os.MkdirAll(infoFileDir, 0755)
       if err != nil {
           return errors.Wrapf(err, "can not create directory (%v)", infoFileDir)
       }
    }
    infoFilePath := filepath.Join(infoFileDir, fileID)
    file, err := os.Create(infoFilePath)
    if err != nil {
        return errors.Wrapf(err, "can not create file info (%v)", infoFilePath)
    }
    defer file.Close()
    enc := gob.NewEncoder(file)
    err = enc.Encode(f.fileInfo)
    if err != nil {
        return errors.Wrapf(err, "can not encode file info (%v)", infoFilePath)
    }
    return nil
}

func (f *FileChecker)callNotify(data []byte, fileID string, fileName string, pathMatcher *configurator.PathMatcher) (error) {
    notifierPlugins := make([]notifierplugger.NotifierPlugin, 0, len(pathMatcher.Notifiers))
    for _, notifier := range pathMatcher.Notifiers {
        pluginFilePath, pluginNewFunc, ok := notifierplugger.GetNotifierPlugin(notifier.Name)
        if !ok {
            return errors.Errorf("not found notifier plugin (%v)", notifier.Name)
        }
        configPath := path.Join(filepath.Dir(pluginFilePath), notifier.Config)
        plugin, err := pluginNewFunc(f.callers, configPath)
        if err != nil {
            return errors.Wrapf(err, "can not create plugin (%v, %v)",  notifier.Name, configPath )
        }
        notifierPlugins = append(notifierPlugins, plugin)
    }
    for _, notifierPlugin := range notifierPlugins {
        notifierPlugin.Notify(data, fileID, fileName, pathMatcher.Label)
    }
    return nil
}

// Check is check
func (f *FileChecker)Check(fileID string, trackLinkFile string, fileName string, pathMatcher *configurator.PathMatcher) (error) {
    if f.fileInfo == nil {
        err := f.loadFileInfo(fileID)
        if err != nil {
            return errors.Wrapf(err, "can not load file info (%v)", fileID)
        }
        f.fileInfo = &fileInfo{
            FileID: fileID,
            TrackLinkFile: trackLinkFile,
            Pos: 0,
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
    if fi.Size() <= f.fileInfo.Pos {
        return nil
    }
    oldPos := f.fileInfo.Pos
    file, err := os.Open(trackLinkFile)
    if err != nil {
        return errors.Wrapf(err, "can not open trackLinkFile (%v)", trackLinkFile)
    }
    defer file.Close()
    _, err = file.Seek(f.fileInfo.Pos, 0)
    if err != nil {
        return errors.Wrapf(err, "can not seek trackLinkFile (%v)", trackLinkFile)
    }
    reader := bufio.NewReader(file)
    for {
        data, err := reader.ReadBytes('\n')
        if err != nil {
            if err != io.EOF {
                log.Printf("can not read bytes (%v:%v): %v", trackLinkFile, f.fileInfo.Pos, err)
            }
            break
        }
        trimData := data[:len(data) -1]
        for _, matcher := range pathMatcher.MsgMatchers {
            matched, err := regexp.Match(matcher.Pattern, trimData)
            if err != nil {
                log.Printf("can not macth message (%v, %v): %v", matcher.Pattern, string(data), err)
            }
            if !matched {
                continue
            }
            if f.config.SkipNotify || pathMatcher.SkipNotify {
                continue
            }
            err = f.callNotify(data, fileID, fileName, pathMatcher)
            if err != nil {
                log.Printf("can not notify (%v, %v): %v", matcher.Pattern, string(data), err)
            } else {
                log.Printf("notified (%v)", matcher.Pattern)
            }
        }
        f.fileInfo.Pos += int64(len(data))
    }
    if oldPos == f.fileInfo.Pos {
        return nil
    }
    err = f.saveFileInfo(f.fileInfo.FileID)
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
