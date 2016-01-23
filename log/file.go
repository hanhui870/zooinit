package log

import (
	"os"
	"bytes"
	"time"
	"io"
	"strings"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
)

// 有缓存的FileLog适配器
type FileLog struct {
	file     *os.File
	buf      *bytes.Buffer
	// 预留多少储存空间
	reseved  int
	// 最多保留多少缓存
	max      int
	// 超出多少时间刷到磁盘
	ttl      time.Duration
	//channel for background worker sync
	syncChan chan int
	//排它锁
	mu       sync.Mutex
}

const (
	DEFAULT_RESERVED = 1024 * 100
// 默认缓存中最多保留1M缓存
	DEFAULT_MAX = 1024 * 1024
// 缓存对最多多久刷新到磁盘
	DEFAULT_TTL = 3 * time.Second
// File mode for log
	DEFAULT_LOGFILE_MODE = 0660
// Dir mode for log, Need exec to read
	DEFAULT_LOGDIR_MODE = 0770
// 写入channel信号
	WRITE_SIGNAL = 1
)

// Notice: 调用函数处必须使用 defer log.Close()
func NewFileLog(logpath string) (*FileLog, error) {
	logpath = strings.Trim(logpath, "")
	if logpath == "" {
		return nil, &os.PathError{"open", logpath, errors.New("File path empty.")}
	}

	err := os.MkdirAll(filepath.Dir(logpath), DEFAULT_LOGDIR_MODE)
	if err != nil {
		return nil, &os.PathError{"create dir", logpath, err}
	}

	file, err := os.OpenFile(logpath, os.O_CREATE | os.O_APPEND | os.O_RDWR | os.O_SYNC, DEFAULT_LOGFILE_MODE)
	if err != nil {
		return nil, &os.PathError{"open", logpath, err}
	}

	logger := &FileLog{file:file, buf:bytes.NewBufferString(""), reseved:DEFAULT_RESERVED, max:DEFAULT_MAX, ttl:DEFAULT_TTL, syncChan:make(chan int)}
	logger.buf.Grow(DEFAULT_RESERVED)

	// Colose File needed. !!!!defer不能用来当析构函数, 这个是在函数结束的时候调用的.
	//defer logger.Close()

	go logger.backgroundSaveWorker()

	return logger, nil
}

func GenerateFileLogPathName(path, service string) (string) {
	date := time.Now()

	fname := fmt.Sprintf("%s/%s.%v.log", path, service, date.Format("2006-01-02"))

	return fname
}

func (f *FileLog) Write(b []byte) (n int, err error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	n, _ = f.buf.Write(b)

	if n < 0 {
		n = 0
	}
	if n != len(b) {
		err = io.ErrShortWrite
	}

	f.syncChan <- WRITE_SIGNAL

	return n, err
}

func (f *FileLog) Close() (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	count, err := f.sync()

	f.file.Close()

	return count, err
}

func (f *FileLog) GetBufferReseved() (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}
	return f.reseved, nil
}

func (f *FileLog) SetBufferReseved(reservedNew int) (error) {
	if f == nil {
		return os.ErrInvalid
	}

	if reservedNew < 0 {
		return errors.New("ReservedNew buffer lenth must>0 ")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.reseved = reservedNew
	f.buf.Grow(f.reseved)

	return nil
}

func (f *FileLog) GetBufferMax() (int, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}
	return f.max, nil
}

func (f *FileLog) SetBufferMax(maxNew int) (error) {
	if f == nil {
		return os.ErrInvalid
	}

	if maxNew < 0 {
		return errors.New("MaxNew buffer lenth must>0 ")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.max = maxNew

	return nil
}

func (f *FileLog) GetBufferTtl() (time.Duration, error) {
	if f == nil {
		return 0, os.ErrInvalid
	}
	return f.ttl, nil
}

func (f *FileLog) SetBufferTtl(ttlNew time.Duration) (error) {
	if f == nil {
		return os.ErrInvalid
	}

	if ttlNew < time.Microsecond {
		return errors.New("MaxNew buffer lenth must>time.Microsecond ")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.ttl = ttlNew

	return nil
}

// sync buffer to persist storage
func (f *FileLog) sync() (n int, err error) {
	if f == nil {
		return 0, os.ErrInvalid
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	n, e := f.file.Write(f.buf.Bytes())
	//reset to blank
	f.buf.Reset()

	if n < 0 {
		n = 0
	}
	if n != len(f.buf.Bytes()) {
		err = io.ErrShortWrite
	}

	if e == nil {
		e = f.file.Sync()
	}

	err = e

	return n, err
}

func (f *FileLog) backgroundSaveWorker() (error) {
	if f == nil {
		return os.ErrInvalid
	}

	for {
		//实现每多少时间保存一次
		timeChan := make(chan int)
		go func() {
			time.Sleep(f.ttl)

			timeChan <- 1
		}()

		select {
		case writeSignal := <-f.syncChan:
			if writeSignal == WRITE_SIGNAL {
				if f.buf.Len() > f.max {
					_, err := f.sync()
					if err != nil {
						panic(err)
					}
				}
			}

		case writeSignal := <-timeChan:
			if writeSignal == WRITE_SIGNAL {
				fmt.Println("receive write signal:", writeSignal)
				_, err := f.sync()
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return nil
}