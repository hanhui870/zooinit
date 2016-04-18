// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package etcd

import (
	"strconv"
	"strings"
)

func CheckInOrderKeyFormat(path string) bool {
	lastpart := path[strings.LastIndex(path, "/")+1:]

	_, err := strconv.ParseInt(lastpart, 10, 64)
	if err == nil {
		return true
	}

	return false
}

// get the order value of createInOrderKey
func GetInOrderKeyValue(path string) int64 {
	var lastpart string
	if strings.LastIndex(path, "/") == -1 {
		lastpart = path
	} else {
		lastpart = path[strings.LastIndex(path, "/")+1:]
	}

	value, err := strconv.ParseInt(lastpart, 10, 64)
	if err == nil {
		return value
	}

	return 0
}
