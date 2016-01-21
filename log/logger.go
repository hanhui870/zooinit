package log

import (
	"log"
	"os"
	"io"
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

// Fetch a file based file service
func GetFileLogger(filename string)(*log.Logger){
	file, err:=os.OpenFile(filename, os.O_CREATE | os.O_APPEND | os.O_RDWR | os.O_SYNC, 0660)
	if err != nil {
		log.Fatalln()
	}

	return log.New(file, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

// Fetch a file based file service and write to os.Stdout at the same time
func GetConsoleFileMultiLogger(filename string)(*log.Logger){
	file, err:=os.OpenFile(filename, os.O_CREATE | os.O_APPEND | os.O_RDWR | os.O_SYNC, 0660)
	if err != nil {
		log.Fatalln()
	}

	writer := io.MultiWriter(os.Stdout, file)

	return log.New(writer, "", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}


