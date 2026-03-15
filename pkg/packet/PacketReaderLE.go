
package packet

import (
	"encoding/binary"
	"unsafe"
)

func (r *PacketReader) Uint16(x *uint16) {
	b := make([]byte, 2)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*uint16)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) Int16(x *int16) {
	b := make([]byte, 2)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*int16)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) Uint32(x *uint32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*uint32)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) Int32(x *int32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*int32)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) BEInt32(x *int32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = int32(binary.BigEndian.Uint32(b))
}

func (r *PacketReader) Uint64(x *uint64) {
	b := make([]byte, 8)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*uint64)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) Int64(x *int64) {
	b := make([]byte, 8)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*int64)(unsafe.Pointer(&b[0]))
}

func (r *PacketReader) Float32(x *float32) {
	b := make([]byte, 4)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}
	*x = *(*float32)(unsafe.Pointer(&b[0]))
}
