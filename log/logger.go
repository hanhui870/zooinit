package log

import (
	"log"
	"io"
	"os"
)

var(
	defaultLogger *log.Logger
)

func init (){
	writer:=io.MultiWriter(os.Stdout)

	defaultLogger= log.New(writer, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func Logger() (*log.Logger){
	return defaultLogger
}
