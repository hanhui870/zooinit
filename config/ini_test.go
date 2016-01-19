package config

import (
	"testing"
	"os"
)

func TestFileLoggerNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err!=nil {
		t.Error(err)
	}
	t.Log("Working dir now:", dir)

	file:=dir+"/config.ini"

	ini:=Ini(file)
	sec, err:=ini.GetSection("")
	if err!=nil {
		t.Error(err)
	}

	key, err:=sec.GetKey("log.path")
	if err!=nil {
		t.Error(err)
	}
	if key.String()!="/Users/bruce/project/godev/src/zooinit/log" {
		t.Error("Get log.path error:", key.String())
	}
	t.Log("Get log.path:", key.String())
}


