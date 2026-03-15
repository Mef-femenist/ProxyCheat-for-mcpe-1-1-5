package utils

func RemoveIndexByteArray(s []byte, index int) []byte {
	ret := make([]byte, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
