package nbt

import (
	"log"
	"math"
	"unsafe"
)

type Encoding interface {
	Int16(r *offsetReader) (int16, error)
	Int32(r *offsetReader) (int32, error)
	Int64(r *offsetReader) (int64, error)
	Float32(r *offsetReader) (float32, error)
	Float64(r *offsetReader) (float64, error)
	String(r *offsetReader) (string, error)

	WriteInt16(w *offsetWriter, x int16) error
	WriteInt32(w *offsetWriter, x int32) error
	WriteInt64(w *offsetWriter, x int64) error
	WriteFloat32(w *offsetWriter, x float32) error
	WriteFloat64(w *offsetWriter, x float64) error
	WriteString(w *offsetWriter, x string) error
}

var NetworkLittleEndian networkLittleEndian

var LittleEndian littleEndian

var BigEndian bigEndian

var _ = BigEndian
var _ = LittleEndian
var _ = NetworkLittleEndian

type networkLittleEndian struct{ littleEndian }

func (networkLittleEndian) WriteInt32(w *offsetWriter, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt32", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

func (networkLittleEndian) WriteInt64(w *offsetWriter, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteInt64", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

func (networkLittleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		log.Println("max int16 nbt encoding 341")
		
	}
	ux := uint32(len(x))
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return FailedWriteError{Op: "WriteString", Off: w.off}
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

func (networkLittleEndian) Int32(r *offsetReader) (int32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int32"}
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int32(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

func (networkLittleEndian) Int64(r *offsetReader) (int64, error) {
	var ux uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, BufferOverrunError{Op: "Int64"}
		}
		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

func (e networkLittleEndian) String(r *offsetReader) (string, error) {
	var length uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return "", BufferOverrunError{Op: "String"}
		}
		length |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	if length > math.MaxInt16 {
		log.Println("max int16 nbt encoding 341")
		
	}
	data, err := consumeN(int(length), r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

type littleEndian struct{}

func (littleEndian) WriteInt16(w *offsetWriter, x int16) error {
	if _, err := w.Write([]byte{byte(x), byte(x >> 8)}); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

func (littleEndian) WriteInt32(w *offsetWriter, x int32) error {
	if _, err := w.Write([]byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

func (littleEndian) WriteInt64(w *offsetWriter, x int64) error {
	if _, err := w.Write([]byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24),
		byte(x >> 32), byte(x >> 40), byte(x >> 48), byte(x >> 56)}); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

func (littleEndian) WriteFloat32(w *offsetWriter, x float32) error {
	bits := math.Float32bits(x)
	if _, err := w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24)}); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

func (littleEndian) WriteFloat64(w *offsetWriter, x float64) error {
	bits := math.Float64bits(x)
	if _, err := w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
		byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56)}); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

func (littleEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		log.Println("max int16 nbt encoding 341")
		
	}
	length := int16(len(x))
	if _, err := w.Write([]byte{byte(length), byte(length >> 8)}); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

func (littleEndian) Int16(r *offsetReader) (int16, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), nil
}

func (littleEndian) Int32(r *offsetReader) (int32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

func (littleEndian) Int64(r *offsetReader) (int64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

func (littleEndian) Float32(r *offsetReader) (float32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

func (littleEndian) Float64(r *offsetReader) (float64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

func (littleEndian) String(r *offsetReader) (string, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	stringLength := int(uint16(b[0]) | uint16(b[1])<<8)
	data, err := consumeN(stringLength, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

type bigEndian struct{}

func (bigEndian) WriteInt16(w *offsetWriter, x int16) error {
	if _, err := w.Write([]byte{byte(x >> 8), byte(x)}); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	return nil
}

func (bigEndian) WriteInt32(w *offsetWriter, x int32) error {
	if _, err := w.Write([]byte{byte(x >> 24), byte(x >> 16), byte(x >> 8), byte(x)}); err != nil {
		return FailedWriteError{Op: "WriteInt32", Off: w.off}
	}
	return nil
}

func (bigEndian) WriteInt64(w *offsetWriter, x int64) error {
	if _, err := w.Write([]byte{byte(x >> 56), byte(x >> 48), byte(x >> 40), byte(x >> 32),
		byte(x >> 24), byte(x >> 16), byte(x >> 8), byte(x)}); err != nil {
		return FailedWriteError{Op: "WriteInt64", Off: w.off}
	}
	return nil
}

func (bigEndian) WriteFloat32(w *offsetWriter, x float32) error {
	bits := math.Float32bits(x)
	if _, err := w.Write([]byte{byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)}); err != nil {
		return FailedWriteError{Op: "WriteFloat32", Off: w.off}
	}
	return nil
}

func (bigEndian) WriteFloat64(w *offsetWriter, x float64) error {
	bits := math.Float64bits(x)
	if _, err := w.Write([]byte{byte(bits >> 56), byte(bits >> 48), byte(bits >> 40), byte(bits >> 32),
		byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)}); err != nil {
		return FailedWriteError{Op: "WriteFloat64", Off: w.off}
	}
	return nil
}

func (bigEndian) WriteString(w *offsetWriter, x string) error {
	if len(x) > math.MaxInt16 {
		log.Println("max int16 nbt encoding 341")
		
	}
	length := int16(len(x))
	if _, err := w.Write([]byte{byte(length >> 8), byte(length)}); err != nil {
		return FailedWriteError{Op: "WriteInt16", Off: w.off}
	}
	
	if _, err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return FailedWriteError{Op: "WriteString", Off: w.off}
	}
	return nil
}

func (bigEndian) Int16(r *offsetReader) (int16, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int16"}
	}
	return int16(uint16(b[0])<<8 | uint16(b[1])), nil
}

func (bigEndian) Int32(r *offsetReader) (int32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Int32"}
	}
	return int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

func (bigEndian) Int64(r *offsetReader) (int64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return int64(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

func (bigEndian) Float32(r *offsetReader) (float32, error) {
	b, err := consumeN(4, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float32"}
	}
	return math.Float32frombits(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

func (bigEndian) Float64(r *offsetReader) (float64, error) {
	b, err := consumeN(8, r)
	if err != nil {
		return 0, BufferOverrunError{Op: "Float64"}
	}
	return math.Float64frombits(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

func (bigEndian) String(r *offsetReader) (string, error) {
	b, err := consumeN(2, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	stringLength := int(uint16(b[0])<<8 | uint16(b[1]))
	data, err := consumeN(stringLength, r)
	if err != nil {
		return "", BufferOverrunError{Op: "String"}
	}
	return string(data), nil
}

func consumeN(n int, r *offsetReader) ([]byte, error) {
	if n < 0 {
		return nil, InvalidArraySizeError{Off: r.off, Op: "Consume", NBTLength: n}
	}
	data := r.Next(n)
	if len(data) != n {
		return nil, BufferOverrunError{Op: "Consume"}
	}
	return data, nil
}
