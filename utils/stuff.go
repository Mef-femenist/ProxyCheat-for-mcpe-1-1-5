package utils

import "strconv"

func FixClientIDLen(cid int64) int64 {
	strid := strconv.FormatInt(cid, 10)
	if len(strid) > 19 { 
		strid = strid[:19]
	}
	out, _ := strconv.ParseInt(strid, 10, 64)
	return out
}
