package log

import (
	"testing"
	"os"
	"zooinit/config"
	"path/filepath"
	"time"
	"bytes"
)

func TestFileLoggerNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	cfgpath := filepath.Dir(dir) + "/config/config.ini"
	t.Log("Working dir now:", dir, " Config path:", cfgpath)

	ini := config.Ini(cfgpath)
	logPath:=ini.Section("production").Key("log.path").String()
	t.Log("Log path:", logPath)

	date:=time.Now();
	t.Log("DateTime Now:", date.Format(time.RFC3339))
	//这个日期是固定的
	t.Log("Date Now:", date.Format("2006-01-02"))
	t.Log("filelognane:", GenerateFileLogPathName(logPath, "test/zooinit"))

	log, err:=NewFileLog(GenerateFileLogPathName(logPath, "test/zooinit"))
	if err!=nil {
		t.Error(err)
	}

	log.Write(bytes.NewBufferString("hello world").Bytes())

}
