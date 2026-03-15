
package machineid

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net"
	"strings"
)

type Result struct {
	Phone   string
	ID      string
	IsLinux bool
}

var prev *Result

func SendNotify(head, msg string) {}
func SendToast(msg string)        {}
func SaveAddress(addr string)     {}
func MaybeRestart()               {}

var OnRestartCallback func(addr string)

func Get() (Result, error) {
	if prev != nil {
		return *prev, nil
	}
	ifas, err := net.Interfaces()
	if err != nil {
		return Result{}, errors.New("не удалось определить сетевые интерфейсы")
	}
	for _, ifa := range ifas {
		mac := ifa.HardwareAddr.String()
		if mac == "" {
			continue
		}
		if strings.Contains(ifa.Flags.String(), "loopback") {
			continue
		}
		hd := md5.Sum([]byte(mac + "gopisa"))
		res := Result{
			ID:      hex.EncodeToString(hd[:])[:10] + "w",
			IsLinux: false,
		}
		prev = &res
		return res, nil
	}
	hd := md5.Sum([]byte("fallback_gopisa"))
	res := Result{ID: hex.EncodeToString(hd[:])[:10] + "w"}
	prev = &res
	return res, nil
}
