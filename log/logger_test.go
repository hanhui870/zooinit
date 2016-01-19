package log

import (
	"testing"
)

func TestLoggerNormal(t *testing.T)  {
	Logger().Println("hello world")
}

