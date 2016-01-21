package utility

import (
	"net"
	"errors"
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
func HasIpAddress(findip string) (bool) {
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

// Find the mask of the IP Adress
func MaskOFIpAddress(findip string) (net.IPMask) {
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
func GetLocalIPWithIntranet(findip string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ipobj:=net.ParseIP(findip)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if ipobj.Mask(v.Mask).String()==v.IP.Mask(v.Mask).String() {
					return v.IP, nil
				}
			}
		}
	}

	return nil, errors.New("Not found.")
}