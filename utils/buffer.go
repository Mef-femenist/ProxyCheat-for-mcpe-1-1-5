package utils

import (
	"bytes"
	"compress/zlib"
	"mefproxy/pkg/packet"
	"unsafe"
)

func ReadString(buffer *bytes.Buffer) string {
	length := ReadUnsignedVarInt(buffer)
	l := length
	data := make([]byte, l)
	_, _ = buffer.Read(data)
	return *(*string)(unsafe.Pointer(&data))
}

func ReadBytesLen(buffer *bytes.Buffer) []byte {
	length := ReadUnsignedVarInt(buffer)
	data := make([]byte, length)
	_, _ = buffer.Read(data)
	return data
}

func ReadBytesLen014(buffer *bytes.Buffer) []byte {
	bb := packet.NewReader(buffer, 3)
	var length int32
	bb.Int32(&length)
	data := make([]byte, length)
	_, _ = buffer.Read(data)
	return data
}

func ReadUnsignedVarInt(buffer *bytes.Buffer) int {
	var v int
	for i := 0; i < 35; i += 7 {
		b, _ := buffer.ReadByte()
		v |= int(b&0x7f) << i
		if b&0x80 == 0 {
			return v
		}
	}
	return 0
}

func WriteVarUInt32(dst *zlib.Writer, x uint32, b []byte) error {
	for i := 0; i <= 4; i++ {
		b[i] = 0
	}
	i := 0
	for x >= 0x80 {
		b[i] = byte(x) | 0x80
		i++
		x >>= 7
	}
	b[i] = byte(x)
	_, err := dst.Write(b[:i+1])
	return err
}
