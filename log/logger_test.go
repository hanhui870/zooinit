package log

import (
	"testing"
	"log"
	"io"
	"os"
	"path/filepath"
	"zooinit/config"
)

func TestMuiltWriterLogger(t *testing.T) {
	Logger().Println("Test Log muilter writers")

	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	cfgpath := filepath.Dir(dir) + "/config/config_for_test.ini"

	ini := config.Ini(cfgpath)
	logPath := ini.Section("production").Key("log.path").String()

	f, err := NewFileLog(GenerateFileLogPathName(logPath, "test/zoomuiltloger"))
	if err != nil {
		t.Error(err)
	}

	//必须调用不然会丢失日志
	defer f.Close()

	writer := io.MultiWriter(os.Stdout, f)
	logger := log.New(writer, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	logger.Println("hello world from TestMuiltWriterLogger")
}

