package log

import (
	"log"
	"os"
)

var(
	defaultLogger *log.Logger
)

func init (){
	defaultLogger= log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func Logger() (*log.Logger){
	return defaultLogger
}
