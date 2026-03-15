package packet

import (
	"encoding/binary"
	"unsafe"
)

func (w *PacketWriter) Uint16(x *uint16) {
	data := *(*[2]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) Int16(x *int16) {
	data := *(*[2]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) Uint32(x *uint32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) Int32(x *int32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) BEInt32(x *int32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(*x))
	_, _ = w.w.Write(data)
}

func (w *PacketWriter) Uint64(x *uint64) {
	data := *(*[8]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) Int64(x *int64) {
	data := *(*[8]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}

func (w *PacketWriter) Float32(x *float32) {
	data := *(*[4]byte)(unsafe.Pointer(x))
	_, _ = w.w.Write(data[:])
}
