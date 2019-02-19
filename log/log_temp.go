package log

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

type LogObject struct {
	content  bytes.Buffer
	fileName string
}

func NewLogObject(fileName string) *LogObject {
	return &LogObject{
		fileName: fileName,
	}
}

func (logObject *LogObject) Content() string {
	return logObject.content.String()
}

func (logObject *LogObject) Log(format string, a ...interface{}) string {
	if logObject.content.Len() > 1024*100 {
		logObject.content.Reset()
	}
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("Info <%s>: %s\n", time.Now().Format(time.ANSIC), logStr)
	logObject.content.WriteString(logStr)

	return logStr
}

func (logObject *LogObject) Save() {
	// generate file path
	if rootDirectory == "" {
		return
	}
	filePath := fmt.Sprintf("%s/conf/log/%s", rootDirectory, logObject.fileName)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		LogSerious("log to file %s fail %s, log content %s", logObject.fileName, err.Error(), logObject.content)
	}

	defer f.Close()

	if _, err = f.WriteString(logObject.content.String()); err != nil {
		LogSerious("log to file %s fail %s, log content %s", logObject.fileName, err.Error(), logObject.content)
	}
}

func (logObject *LogObject) Clear() {
	logObject.content.Reset()
}
