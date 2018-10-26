package logstore

import (
    "os"
    "path"
    "path/filepath"
    "strings"
    "context"
    "github.com/pkg/errors"
    "github.com/potix/log_monitor/configurator"
    logpb "github.com/potix/log_monitor/logpb"
)

const (
    defaultPathFormat string = "${LABEL}/${HOST}_${ADDR}/${FILE_PATH}"
)

// LogStore is LogStore
type LogStore struct {
    config *configurator.LogRecieverConfig
}

// Save is save
func (l *LogStore)Save(ctx context.Context, addr string, request *logpb.TransferRequest) (error) {
    format := defaultPathFormat
    if l.config.PathFormat != "" {
        format = l.config.PathFormat
    } 
    r := strings.NewReplacer("${LABEL}", request.Label, "${HOST}", request.Host, "${ADDR}", addr, "${FILE_PATH}", request.Path)
    formatPath := r.Replace(format)
    filePath := filepath.Join(l.config.Path, formatPath)
    err := os.MkdirAll(path.Dir(filePath), 0755)
    if err != nil {
        return errors.Wrapf(err, "can not create directories (%v)", filePath)
    }
    file, err :=  os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return errors.Wrapf(err, "can not open file (%v)", filePath)
    }
    defer file.Close()
    _, err = file.Write(request.LogData)
    if err != nil {
        return errors.Wrapf(err, "can not write log data (%v)", filePath)
    }
    return nil
}

// NewLogStore is  create new log store
func NewLogStore(config *configurator.LogRecieverConfig) (*LogStore) {
     return &LogStore{
         config: config,
     }
}
