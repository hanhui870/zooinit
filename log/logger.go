package log

import (
	"errors"
	"io"
	"log"
	"os"
)

// TODO 20160223 Logger dispear without save if process terminate
type LoggerIOAdapter struct {
	logger *log.Logger
	prefix string
	suffix string
}

var (
	defaultLogger *log.Logger
)

func init() {
	defaultLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func Logger() *log.Logger {
	return defaultLogger
}

// Fetch a file based file service
func GetFileLogger(filename string) *log.Logger {
	file, err := NewFileLog(filename)
	if err != nil {
		log.Fatalln(err)
	}

	return log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

// Fetch a file based file service and write to os.Stdout at the same time
func GetConsoleFileMultiLogger(filename string) *log.Logger {
	file, err := NewFileLog(filename)
	if err != nil {
		log.Fatalln(err)
	}

	writer := io.MultiWriter(os.Stdout, file)

	return log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

// return new LoggerIOAdapter with Writer interface
func NewLoggerIOAdapter(logger *log.Logger) *LoggerIOAdapter {
	return &LoggerIOAdapter{logger: logger}
}

func (o *LoggerIOAdapter) SetPrefix(p string) {
	if o != nil {
		o.prefix = p
	}
}

func (o *LoggerIOAdapter) SetSuffix(s string) {
	if o == nil {
		o.suffix = s
	}
}

func (o *LoggerIOAdapter) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, errors.New("Object not exists.")
	}

	// Noneed add \n, caller process this
	o.logger.Print(o.prefix + string(p) + o.suffix)
	return len(p), nil
}
