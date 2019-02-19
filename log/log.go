package log

import (
	"database/sql"
	"fmt"
	"gopkg.in/gomail.v1"
	"os"
	"runtime"
	"time"
)

var rootDirectory string
var httpRootUrl string
var shouldSendAdminEmail bool
var shouldLogToFile bool
var mailer = gomail.NewMailer("smtp.gmail.com", "daominah@gmail.com", "talaTung208", 587)

// Z2lhLmRhbmduZ3V5ZW5ob2FuZ0BnbWFpbC5jb20=
// YmdnbGd3dWl4emF3Y3Rvdw==
func EnableAdminEmail() {
	shouldSendAdminEmail = true
}

func EnableLogToFile() {
	shouldLogToFile = true
}

func SetRootDirectory(aRootDirectory string) {
	rootDirectory = aRootDirectory
}

func SetHttpRootUrl(aHttpRootUrl string) {
	httpRootUrl = aHttpRootUrl
}

func Log(format string, a ...interface{}) string {
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("Info <%s>: %s", time.Now().Format(time.ANSIC), logStr)
	fmt.Println(logStr)

	return logStr
}

func GetStack() string {
	trace := make([]byte, 8192)
	count := runtime.Stack(trace, false)
	content := fmt.Sprintf("Dump (%d bytes):\n %s \n", count, trace)
	return content
}

func LogSerious(format string, a ...interface{}) string {
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("SERIOUS <%s>: %s", time.Now().Format(time.ANSIC), logStr)
	fmt.Println(logStr)

	if shouldSendAdminEmail {
		// fmt.Println("SENDADMINMAIL")
		trace := make([]byte, 8192)
		count := runtime.Stack(trace, false)
		content := fmt.Sprintf("%s \n Dump (%d bytes):\n %s \n", logStr, count, trace)
		go sendMail(content)
	}

	return logStr
}

func LogWithAttach(title string, attachFilePath string) {
	if shouldSendAdminEmail {
		msg := gomail.NewMessage()
		msg.SetHeader("From", "daominah@gmail.com")
		msg.SetHeader("To", "daominah@gmail.com")
		msg.SetHeader("Subject", fmt.Sprintf("%s %s", title, httpRootUrl))
		msg.SetBody("text/plain", "attachment")
		file, err := gomail.OpenFile(attachFilePath)
		if err != nil {
			fmt.Printf("SERIOUS <%s>: get attach file error %s \n", time.Now().Format(time.ANSIC), err.Error())
			return
		}
		msg.Attach(file)
		if err := mailer.Send(msg); err != nil {
			fmt.Printf("SERIOUS <%s>: Mail send error %s \n", time.Now().Format(time.ANSIC), err.Error())
		}
	}
}

func LogSeriousWithStack(format string, a ...interface{}) string {
	logStr := fmt.Sprintf(format, a...)
	trace := make([]byte, 8192)
	count := runtime.Stack(trace, false)
	content := fmt.Sprintf("Dump (%d bytes):\n %s \n", count, trace)
	logStr = fmt.Sprintf("SERIOUS <%s>: %s\n%s", time.Now().Format(time.ANSIC), logStr, content)
	fmt.Println(logStr)

	if shouldSendAdminEmail {
		trace := make([]byte, 8192)
		count := runtime.Stack(trace, false)
		content := fmt.Sprintf("%s \n Dump (%d bytes):\n %s \n", logStr, count, trace)
		go sendMail(content)
	}

	return logStr
}

func CreateFileAndLog(fileName string, format string, a ...interface{}) (content string, filePath string) {
	if !shouldLogToFile {
		return "", ""
	}
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("Info <%s>: %s\n", time.Now().Format(time.ANSIC), logStr)

	// generate file path
	filePath = fmt.Sprintf("%s/conf/log/%s", rootDirectory, fileName)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
		return "", ""
	}

	defer f.Close()

	if _, err = f.WriteString(logStr); err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
	}
	return logStr, filePath
}

func LogFile(fileName string, format string, a ...interface{}) string {
	if !shouldLogToFile {
		return ""
	}
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("Info <%s>: %s\n", time.Now().Format(time.ANSIC), logStr)

	// generate file path
	filePath := fmt.Sprintf("%s/conf/log/%s", rootDirectory, fileName)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
	}

	defer f.Close()

	if _, err = f.WriteString(logStr); err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
	}
	return logStr
}

func LogSeriousFile(fileName string, format string, a ...interface{}) string {
	if !shouldLogToFile {
		return ""
	}
	logStr := fmt.Sprintf(format, a...)
	logStr = fmt.Sprintf("SERIOUS <%s>: %s\n", time.Now().Format(time.ANSIC), logStr)

	// generate file path
	filePath := fmt.Sprintf("%s/conf/log/%s", rootDirectory, fileName)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
	}

	defer f.Close()

	if _, err = f.WriteString(logStr); err != nil {
		LogSerious("log to file %s fail %s,log content %s", fileName, err.Error(), logStr)
	}
	return logStr
}

func SendMailWithCurrentStack(message string) string {
	trace := make([]byte, 8192)
	count := runtime.Stack(trace, false)
	content := fmt.Sprintf("%s \n Dump (%d bytes):\n %s \n", message, count, trace)
	sendMail(content)
	Log(content)
	return content
}

func sendMail(content string) {
	if shouldSendAdminEmail {
		// fmt.Println("SENDADMINMAIL1")
		msg := gomail.NewMessage()
		msg.SetHeader("From", "daominah@gmail.com")
		msg.SetHeader("To", "daominah@gmail.com")
		msg.SetHeader("Subject", fmt.Sprintf("Casino admin message %s", httpRootUrl))
		msg.SetBody("text/plain", content)
		if err := mailer.Send(msg); err != nil {
			// fmt.Println("SENDADMINMAIL ERROR", err)
			fmt.Printf("SERIOUS <%s>: Mail send error %s \n", time.Now().Format(time.ANSIC), err.Error())
		} else {
			fmt.Println("SENDADMINMAIL OK")
		}
	}
}

func DumpRows(rows *sql.Rows) {
	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return
	}

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
		}

		fmt.Printf("%#v\n", result)
	}
}
