package log

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
}
