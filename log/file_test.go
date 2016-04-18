// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package log

import (
	"bytes"
	stdlog "log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"zooinit/config"
)

func TestFileLoggerNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	cfgpath := filepath.Dir(dir) + "/config/config_for_test.ini"
	t.Log("Working dir now:", dir, " Config path:", cfgpath)

	ini := config.Ini(cfgpath)
	logPath := ini.Section("production").Key("log.path").String()
	t.Log("Log path:", logPath)

	date := time.Now()
	t.Log("DateTime Now:", date.Format(time.RFC3339))
	//这个日期是固定的
	t.Log("Date Now:", date.Format("2006-01-02"))
	t.Log("filelognane:", GenerateFileLogPathName(logPath, "test/zooinit"))

	log, err := NewFileLog(GenerateFileLogPathName(logPath, "test/zooinit"))
	if err != nil {
		t.Fatal(err)
	}

	//必须调用不然会丢失日志
	defer log.Close()

	t.Log("Log write count:", []string{"hello", "world"})
	asa := []string{"hello", "world"}
	echo(t, asa...)
	echoInterface(t, asa, "more value")
	t.Log(asa)
	t.Log(log.Write(bytes.NewBufferString("hello world").Bytes()))
	t.Log("log buffer:", log.buf.String())
	t.Log("log buffer2:", log.buf.String())

	//need time
	testLookWrite(t, log)
}

func TestFileLoggerWrap(t *testing.T) {
	log := GetFileLogger(GenerateFileLogPathName("test", "TestFileLoggerWrap"))
	defer log.Sync()
	if log.Flags()&stdlog.LUTC != 0 {
		t.Log("log.Flags()&stdlog.LUTC is set")
	} else {
		t.Log("log.Flags()&stdlog.LUTC is not set")
	}

	log.Println("This is a helloworld from TestFileLoggerWrap.")
}

func echo(t *testing.T, args ...string) {
	t.Log("Call from echo test:", args)
}

func echoInterface(t *testing.T, args ...interface{}) {
	//数组中的数组
	t.Log("Call from echoInterface test:", args)
}

func testLookWrite(t *testing.T, log *FileLog) {
	go func() {
		for i := 0; i < 2; i++ {
			log.Write(bytes.NewBufferString("Test testLookWrite" + strconv.Itoa(i) + "\n").Bytes())
			time.Sleep(5 * time.Second)
		}

	}()

	for i := 0; i < 100; i++ {
		//receive write signal: 1
		//receive write signal: 1
		//receive write signal: 1
		time.Sleep(100 * time.Millisecond)
	}
}
