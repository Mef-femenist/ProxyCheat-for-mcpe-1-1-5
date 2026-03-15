package utils

import (
	"encoding/hex"
	"strconv"
	"strings"
)

func RemoveColor(ss string) string {
	for i := 0; i <= 9; i++ {
		ss = strings.ReplaceAll(ss, "§"+strconv.Itoa(i), "")
	}
	ss = strings.ReplaceAll(ss, "§a", "")
	ss = strings.ReplaceAll(ss, "§b", "")
	ss = strings.ReplaceAll(ss, "§c", "")
	ss = strings.ReplaceAll(ss, "§d", "")
	ss = strings.ReplaceAll(ss, "§e", "")
	ss = strings.ReplaceAll(ss, "§f", "")
	ss = strings.ReplaceAll(ss, "§g", "")
	ss = strings.ReplaceAll(ss, "§k", "")
	ss = strings.ReplaceAll(ss, "§o", "")
	ss = strings.ReplaceAll(ss, "§r", "")
	ss = strings.ReplaceAll(ss, "§l", "")
	
	return ss
}

func StrpadLeft(str, pad string, count int) string {
	if len(str) >= count {
		return str
	}
	count = count - len(str)
	str = strings.Repeat(pad, count) + str
	return str
}

func Dump(d []byte) string {
	tmp := hex.EncodeToString(d)
	out := ""
	for i := 0; i < len(tmp); i += 2 {
		out += "\\x" + tmp[i:i+1] + tmp[i:i+2]
	}
	return out
}
