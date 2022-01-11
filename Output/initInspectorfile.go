package Output

import (
	"errors"
	"fmt"
	"mongostatus/files"
	"mongostatus/sys"
	"os"
	"path/filepath"
	"time"
)

type logcontent struct {
	logText string
}

func addLogContant(content string) *logcontent {
	return &logcontent{
		content,
	}
}

func (this *logcontent) printmessage() {
	//打印在屏幕上
	t := time.Now()
	header := fmt.Sprintf("%d-%d-%d %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	fmt.Printf("%s %s\n", header, this.logText)

}
func (this *logcontent) writeinspectorfile(f string) {
	//写文件方法
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	t := time.Now()
	header := fmt.Sprintf("%d-%d-%d %d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	text := header + this.logText + "\n"
	_, _ = file.WriteString(text)
	_ = file.Close()
}

func (this *logcontent) writefile(f string) {
	//写文件方法
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}
	text := this.logText
	_, _ = file.WriteString(text)
	_ = file.Close()
}

func (this *logcontent) truncateFile(f string) {
	file, err := os.OpenFile(f, os.O_RDWR|os.O_TRUNC, 0660)
	if err != nil {
		panic(err)
	}
	file.Close()
}

func Initresultfile(fn string) {
	var u logPrint
	u = addLogContant("Init....")
	u.truncateFile(fn)
}
func DoResult(longtext string, fn string) {
	var u logPrint
	u = addLogContant(longtext)
	u.writeinspectorfile(fn)
	u.printmessage()
}

func Writeins(longtext, fn string) {
	var u logPrint
	u = addLogContant(longtext + "\n")
	u.writefile(fn)
}

const inspectorfiledirname = "inspector"

func InitInsepectorFile(isf string) (ISFName string, err error) {
	// 检测是否存在 ./inspector这个目录
	var flagStorePath string
	flagStorePath = filepath.Join(sys.CurrentDirectory(), inspectorfiledirname)
	if err := files.MkdirIfNecessary(flagStorePath); err != nil {
		return "", errors.New(fmt.Sprintf("create boltdb store : %s", err.Error()))
	}
	// 检测是否存在 ./inspector/$mongo_conn这个文件 此文件用户存储巡检报告
	var FlagFilePath string
	FlagFilePath = filepath.Join(flagStorePath, isf)
	files.CreateFileIfNecessary(FlagFilePath)
	return FlagFilePath, nil
}
