package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func getBootedFlagFile() string {
	return "/tmp/test/booted"
}

func makeClusterBooted() (bool, error) {
	file, err := os.OpenFile(getBootedFlagFile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return false, &os.PathError{"open", getBootedFlagFile(), err}
	}

	defer file.Close()
	n, err := file.Write(bytes.NewBufferString(time.Now().Format(time.RFC3339)).Bytes())
	if err != nil {
		return false, err
	} else if n == 0 {
		err = errors.New("Write 0 length content for makeClusterBooted()")
		return false, err
	}

	return true, nil
}

func IsClusterBootedBefore() bool {
	file, err := os.Open(getBootedFlagFile())
	if err == nil {
		content, err := ioutil.ReadAll(file)
		if err == nil {
			ti, err := time.Parse(time.RFC3339, string(content))
			if err == nil {
				fmt.Println("Cluster ever booted at:", ti.Format(time.RFC3339))
				return true
			}
		}
	}

	return false
}

func main() {
	fmt.Println("is Booted:", IsClusterBootedBefore())
	fmt.Println("mark Booted...")
	makeClusterBooted()
	fmt.Println("is Booted:", IsClusterBootedBefore())
}
