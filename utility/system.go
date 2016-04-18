// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package utility

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// Fetch ip address of the machine
func GetIpAddress(eth string) (map[string][]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var iplist map[string][]net.IP
	iplist = make(map[string][]net.IP)

	for _, i := range ifaces {
		//fetch specific interface ip address
		if eth != "" && eth != i.Name {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		var ip []net.IP
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = append(ip, v.IP)
			}
		}
		iplist[i.Name] = ip
	}

	return iplist, nil
}

// Does the machine owns the ip
func HasIpAddress(findip string) bool {
	iplist, err := GetIpAddress("")
	if err != nil {
		panic(err)
	}

	for _, value := range iplist {
		for _, ip := range value {
			if findip == ip.String() {
				return true
			}
		}
	}

	return false
}

// Find the real mask of the IP Adress
func MaskOFIpAddress(findip string) net.IPMask {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, i := range ifaces {

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.String() == findip {
					return v.Mask
				}
			}
		}
	}

	return nil
}

// Fetch the IP which is in the same intranet with findip
// Remove quote symbol
func GetLocalIPWithIntranet(findip string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ipobj := net.ParseIP(findip)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ipobj.Mask(v.Mask).String() == v.IP.Mask(v.Mask).String() {
					return v.IP, nil
				}
			}
		}
	}

	return nil, errors.New("Not found.")
}

func GetLocalIPWithIntranetIPMask(findip string, maskString string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	mask, err := BuildIPV4MaskFromString(maskString)
	if err != nil {
		return nil, err
	}

	ipobj := net.ParseIP(findip)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ipobj.Mask(mask).String() == v.IP.Mask(mask).String() {
					return v.IP, nil
				}
			}
		}
	}

	return nil, errors.New("Not found.")
}

func GetLocalIPWithInterfaceName(fname string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		//check the interface name
		if i.Name != fname {
			continue
		}

		//many ip fetch the the first one.
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP, nil
			}
		}
	}

	return nil, errors.New("Not found.")
}

func BuildIPV4MaskFromString(mask string) (net.IPMask, error) {
	list := strings.Split(mask, ".")
	if len(list) != 4 {
		return nil, errors.New("IP V4 mask format error.")
	}

	var p1, p2, p3, p4 byte
	i, err := strconv.ParseInt(list[0], 10, 32)
	if err != nil {
		return nil, err
	}
	p1 = byte(i) & 0xFF

	i, err = strconv.ParseInt(list[1], 10, 32)
	if err != nil {
		return nil, err
	}
	p2 = byte(i) & 0xFF

	i, err = strconv.ParseInt(list[2], 10, 32)
	if err != nil {
		return nil, err
	}
	p3 = byte(i) & 0xFF

	i, err = strconv.ParseInt(list[3], 10, 32)
	if err != nil {
		return nil, err
	}
	p4 = byte(i) & 0xFF

	result := net.IPv4Mask(p1, p2, p3, p4)
	return result, nil
}

// Improve for quoted string
func ParseCmdStringWithParams(cmd string) (path string, args []string, err error) {
	delimeter := " \t"
	cmd = strings.Trim(cmd, delimeter)

	// "'" means char in go. not support.
	// "`" has cmd exec meaning in shell. not support.
	quoted := `"`
	escaped := "\\"
	length := len(cmd)

	const (
		BLOCK_CMD = iota
		BLOCK_QUOTED
		BLOCK_ARG
		BLOCK_UNDEFINED
	)

	blockNow := BLOCK_UNDEFINED
	blockStart := -1

	escapePos := -1
	//escapeChar := byte(0) //null \t \n need special process

	// Func for process escaped things.
	doneEscapeFunc := func(str *string, start, end int) string {
		// Can safely use escapePos, escapeChar
		strNew := (*str)[start:end]

		if escapePos != -1 {
			if escapePos > length || escapePos < start {
				panic("Never can happen. escapePos > length || escapePos < start")

			} else {
				rEscapePos := escapePos - start
				strNew = strNew[:rEscapePos] + strNew[rEscapePos+1:]

			}

			// must reset here
			escapePos = -1
		}

		return strNew
	}

	for i := 0; i < length; i++ {
		if strings.ContainsAny(escaped, string(cmd[i])) {
			escapePos = i

			if blockNow == BLOCK_UNDEFINED {
				blockStart = i
				if i == 0 {
					blockNow = BLOCK_CMD
				} else {
					blockNow = BLOCK_ARG
				}
			}

			// ignore next
			i = i + 1
			// Can't be the laster char
			if i >= length {
				return "", nil, errors.New("Escape string can't be the laster char")
			}
			// escapeChar = cmd[i]

			continue
		}

		if i == 0 {
			blockNow = BLOCK_CMD
			blockStart = 0
			continue

		} else if strings.ContainsAny(delimeter, string(cmd[i])) { //end of slice[start:end] no need minus 1
			if blockNow == BLOCK_UNDEFINED {
				blockStart = i + 1
				blockNow = BLOCK_ARG

			} else if blockNow == BLOCK_ARG {
				args = append(args, doneEscapeFunc(&cmd, blockStart, i))
				//reset
				blockStart = -1
				blockNow = BLOCK_UNDEFINED

			} else if blockNow == BLOCK_CMD {
				path = doneEscapeFunc(&cmd, blockStart, i)
				//reset
				blockStart = -1
				blockNow = BLOCK_UNDEFINED

			} else if blockNow == BLOCK_QUOTED {
				//ignore blank in the text
			}

			continue

		} else if strings.ContainsAny(quoted, string(cmd[i])) {
			if blockNow == BLOCK_UNDEFINED {
				blockStart = i + 1
				blockNow = BLOCK_QUOTED

			} else if blockNow == BLOCK_QUOTED {
				if blockStart == -1 {
					return "", nil, errors.New("Run time error, blockStart")
				}
				//Fetch args
				args = append(args, doneEscapeFunc(&cmd, blockStart, i))
				//reset
				blockStart = -1
				blockNow = BLOCK_UNDEFINED

			} else if blockNow == BLOCK_CMD {
				return "", nil, errors.New("Command path should not start with Quoted string.")

			} else if blockNow == BLOCK_ARG {
				if blockStart == i {
					blockStart = i + 1
					blockNow = BLOCK_QUOTED
				} else {
					return "", nil, errors.New("Command arg could not has Quoted string in the middle.")
				}
			}

			continue

		} else if i == length-1 {
			if blockNow == BLOCK_UNDEFINED {
				return "", nil, errors.New("Reach end but block undefined.")

			} else if blockNow == BLOCK_ARG {
				args = append(args, doneEscapeFunc(&cmd, blockStart, length))

			} else if blockNow == BLOCK_CMD {
				path = doneEscapeFunc(&cmd, blockStart, length)

			} else {
				return "", nil, errors.New("BLOCK_QUOTED end exception.")
			}

		} else {
			//char
			if blockNow == BLOCK_UNDEFINED {
				blockStart = i //this char
				blockNow = BLOCK_ARG
			}
		}
	}

	return path, args, nil
}
