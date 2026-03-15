package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ParseIP4(ips string) string {
	ipp := net.ParseIP(ips)
	ret := ips
	if ipp != nil {
		if ipp.To4().String() == ret {
			return ret
		}
	}
	ipss, _ := net.LookupIP(ips)
	for _, ip := range ipss {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func StringToIPPort(addr string) (string, int) {
	address := strings.Split(addr, ":")
	port, _ := strconv.Atoi(address[1])
	return address[0], port
}

func GetLANIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				return ipv4.String()
			}
		}
	}
	return "127.0.0.1"
}

func GetAllLANIPs() []string {
	var result []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"127.0.0.1"}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				result = append(result, fmt.Sprintf("%s (iface: %s)", ipv4.String(), iface.Name))
			}
		}
	}
	if len(result) == 0 {
		return []string{"127.0.0.1"}
	}
	return result
}
