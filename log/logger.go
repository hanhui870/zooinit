package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type BufferedFileLogger struct {
	logger *log.Logger
	file   *FileLog
}

type LoggerIOAdapter struct {
	logger *BufferedFileLogger
	prefix string
	suffix string
}

var (
	defaultLogger *BufferedFileLogger
)

const (
	DEFAULT_LOGGER_FLAGS = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
)

func init() {
	defaultLogger = GetBufferedLogger()
}

func Logger() *BufferedFileLogger {
	return defaultLogger
}

// Fetch a file based file service
func GetBufferedLogger() *BufferedFileLogger {
	logger := log.New(os.Stdout, "", DEFAULT_LOGGER_FLAGS)
	return &BufferedFileLogger{logger: logger}
}

// Fetch a file based file service
func GetFileLogger(filename string) *BufferedFileLogger {
	file, err := NewFileLog(filename)
	if err != nil {
		log.Fatalln(err)
	}

	logger := log.New(file, "", DEFAULT_LOGGER_FLAGS)
	return &BufferedFileLogger{logger: logger, file: file}
}

// Fetch a file based file service and write to os.Stdout at the same time
func GetConsoleFileMultiLogger(filename string) *BufferedFileLogger {
	file, err := NewFileLog(filename)
	if err != nil {
		log.Fatalln(err)
	}

	writer := io.MultiWriter(os.Stdout, file)

	logger := log.New(writer, "", DEFAULT_LOGGER_FLAGS)
	return &BufferedFileLogger{logger: logger, file: file}
}

// return new LoggerIOAdapter with Writer interface
func NewLoggerIOAdapter(logger *BufferedFileLogger) *LoggerIOAdapter {
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

// Arguments are handled in the manner of fmt.Printf.
func (l *BufferedFileLogger) Printf(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *BufferedFileLogger) Print(v ...interface{}) { l.logger.Output(2, fmt.Sprint(v...)) }

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *BufferedFileLogger) Println(v ...interface{}) { l.logger.Output(2, fmt.Sprintln(v...)) }

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *BufferedFileLogger) Fatal(v ...interface{}) {
	l.logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *BufferedFileLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *BufferedFileLogger) Fatalln(v ...interface{}) {
	l.logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *BufferedFileLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.logger.Output(2, s)
	panic(s)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *BufferedFileLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.logger.Output(2, s)
	panic(s)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *BufferedFileLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.logger.Output(2, s)
	panic(s)
}

// Flags returns the output flags for the logger.
func (l *BufferedFileLogger) Flags() int {
	return l.logger.Flags()
}

// SetFlags sets the output flags for the logger.
func (l *BufferedFileLogger) SetFlags(flag int) {
	l.logger.SetFlags(flag)
}

// Prefix returns the output prefix for the logger.
func (l *BufferedFileLogger) Prefix() string {
	return Logger().Prefix()
}

// SetPrefix sets the output prefix for the logger.
func (l *BufferedFileLogger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}

// Log sync protocal
func (l *BufferedFileLogger) Sync() (n int, err error) {
	if l.file == nil {
		return 0, nil
	}

	return l.file.Sync()
}
