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
	cfgpath := filepath.Dir(dir) + "/config/config.ini"

	ini := config.Ini(cfgpath)
	logPath := ini.Section("production").Key("log.path").String()

	f, err := os.OpenFile(GenerateFileLogPathName(logPath, "test/zoomuiltloger"), os.O_CREATE | os.O_APPEND | os.O_RDWR | os.O_SYNC, 0660)
	if err != nil {
		t.Error(err)
	}

	//必须调用不然会丢失日志
	defer f.Close()

	writer := io.MultiWriter(os.Stdout, f)
	logger := log.New(writer, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	logger.Println("hello world from TestMuiltWriterLogger")
}

