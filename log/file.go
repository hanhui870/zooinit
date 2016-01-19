package log

import (
	"os"
	"bytes"
	"time"
	"io"
	"strings"
	"errors"
)

// 有缓存的FileLog适配器
type FileLog struct {
	file    *os.File
	buf     *bytes.Buffer
	// 预留多少储存空间
	reseved int
	// 最多保留多少缓存
	max     int
	// 超出多少时间刷到磁盘
	ttl     time.Duration
	//channel for background worker sync
	syncChan chan int
}

const (
	DEFAULT_RESERVED = 1024 * 100
// 默认缓存中最多保留1M缓存
	DEFAULT_MAX = 1024 * 1024
// 缓存对最多多久刷新到磁盘
	DEFAULT_TTL = 3 * time.Second
)

func NewFileLog(filepath string) (*FileLog, error) {
	filepath = strings.Trim(filepath, "")
	if filepath == "" {
		return nil, &os.PathError{"open", filepath, errors.New("File path empty.")}
	}

	file, err := os.OpenFile(filepath, os.O_CREATE | os.O_APPEND | os.O_RDWR | os.O_SYNC, 0660)
	if err != nil {
		return nil, &os.PathError{"open", filepath, err}
	}
	// Colose File needed.
	defer file.Close()

	logger := &FileLog{file:file, buf:bytes.NewBufferString(""), reseved:DEFAULT_RESERVED, max:DEFAULT_MAX, ttl:DEFAULT_TTL, syncChan:make(chan int)}
	logger.buf.Grow(DEFAULT_RESERVED)

	go logger.backgroundSaveWorker()

	return logger, nil
}

func (f *FileLog) Write(b []byte) (n int, err error) {
	if f == nil {
		return 0, os.ErrInvalid
	}
	n, _ = f.buf.Write(b)

	if n < 0 {
		n = 0
	}
	if n != len(b) {
		err = io.ErrShortWrite
	}

	return n, err
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

	f.ttl = ttlNew

	return nil
}

// sync buffer to persist storage
func (f *FileLog) sync() (n int, err error) {
	if f == nil {
		return 0, os.ErrInvalid
	}
	n, e := f.file.Write(f.buf.Bytes())
	if n < 0 {
		n = 0
	}
	if n != len(f.buf.Bytes()) {
		err = io.ErrShortWrite
	}

	if e != nil {
		err = &os.PathError{"write", f.file.Name(), e}
	}
	return n, err
}

func (f *FileLog) backgroundSaveWorker() (error) {
	if f == nil {
		return os.ErrInvalid
	}

	return nil
}