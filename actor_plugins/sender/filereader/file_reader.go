package filereader

type fileInfo struct {
    fileID string
    fileNameMutex sync.Mutex
    fileName string
    pos int64
}

func (f *fileInfo) getID() {
}

func (f *fileInfo) getFileName() {
}

func (f *fileInfo) setFileName() {
}

func (f *fileInfo) getPos() {
}

func (f *fileInfo) setPos() {
}

type FileReader struct {
    callers string
    config config *configurator.Config
    fileInfo *fileInfo
}

func (f *FileReader)Read(fileID string, filename string) (byte[], error) {
    needInfoFlush := false
    if f.fileInfo != nil {
       fileInfo, err := loadFileInfo(fileID)
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
    }
    st := stat(filename)
    if st.size > f.fileInfo.pos {
        f = f.Open(filename)
        defer f.Close()
        reader := NewReader(f) 
        data, err := reader.ReadBytes('\n')
        if err != nil {
            XXXXXX
        }
        f.fileInfo.pos + len(data) 
        needInfoFlush = true
    }
    if needInfoFlush {
        saveFileInfo(fileID)
    }
}

func (f *FileReader)Rename(fileID string, filename string) (error) {
    f.fileInfo.Rename(filename)
}

func NewFileReader(callers string, config *configurator.Config) {
    return &FileReader {
        callers: callers,
        config: config,
        fileInfo: nil,
    }
}
