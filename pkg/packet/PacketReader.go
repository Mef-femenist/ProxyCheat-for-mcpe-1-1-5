package packet

import (
	"errors"
	"fmt"
	"mefproxy/pkg/math"
	"github.com/google/uuid"
	"mefproxy/pkg/item"
	"mefproxy/pkg/nbt"
	"log"

	"image/color"
	"io"
	
	"math"
	"unsafe"
)

const (
	EntityDataByte uint32 = iota
	EntityDataInt16
	EntityDataInt32
	EntityDataFloat32
	EntityDataString
	EntityDataCompoundTag
	EntityDataBlockPos
	EntityDataInt64
	EntityDataVec3
)

type PacketReader struct {
	r interface {
		io.Reader
		io.ByteReader
	}
	shieldID int32
}

func NewReader(r interface {
	io.Reader
	io.ByteReader
}, shieldID int32) *PacketReader {
	return &PacketReader{r: r, shieldID: shieldID}
}

func (r *PacketReader) Uint8(x *uint8) {
	var err error
	*x, err = r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
}

func (r *PacketReader) Bool(x *bool) {
	u, err := r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
	*x = *(*bool)(unsafe.Pointer(&u))
}

var errStringTooLong = errors.New("string length overflows a 32-bit integer")

func (r *PacketReader) StringUTF(x *string) {
	var length uint16
	r.Uint16(&length)
	log.Println("StringUTF", length)
	l := int(length)
	if l > math.MaxInt16 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

func (r *PacketReader) String(x *string) {
	var length uint32
	r.Varuint32(&length)
	l := int(length)
	if l > math.MaxInt32 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

func (r *PacketReader) ByteSlice(x *[]byte) {
	var length uint32
	r.Varuint32(&length)
	l := int(length)
	if l > math.MaxInt32 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = data
}

func (r *PacketReader) Vec3(x *mgl32.Vec3) {
	r.Float32(&x[0])
	r.Float32(&x[1])
	r.Float32(&x[2])
}

func (r *PacketReader) Vec2(x *mgl32.Vec2) {
	r.Float32(&x[0])
	r.Float32(&x[1])
}

func (r *PacketReader) ByteFloat(x *float32) {
	var v uint8
	r.Uint8(&v)
	*x = float32(v) * (360.0 / 256.0)
}

func (r *PacketReader) VarRGBA(x *color.RGBA) {
	var v uint32
	r.Varuint32(&v)
	*x = color.RGBA{
		R: byte(v),
		G: byte(v >> 8),
		B: byte(v >> 16),
		A: byte(v >> 24),
	}
}

func (r *PacketReader) Bytes(p *[]byte) {
	var err error
	*p, err = io.ReadAll(r.r)
	if err != nil {
		r.panic(err)
	}
}

func (r *PacketReader) BytesLength(p []byte) {
	var err error
	_, err = r.r.Read(p)
	if err != nil {
		r.panic(err)
	}
}

func (r *PacketReader) BlockCoords(x *int32, y *uint32, z *int32) {
	r.Varint32(x)
	r.Varuint32(y)
	r.Varint32(z)
}

func (r *PacketReader) Item() *item.Item {

	var iid int32
	r.Varint32(&iid)
	if iid <= 0 {
		return item.NewItem(0, 0, 0, map[string]interface{}{})
	}
	var auxv int32
	r.Varint32(&auxv)
	data := auxv >> 8
	if data == 0x7fff {
		data = -1
	}
	cnt := auxv & 0xff
	
	var nbtlen int16
	var nbtdata map[string]interface{}
	r.Int16(&nbtlen)
	if nbtlen > 0 {
		r.NBT(&nbtdata, nbt.LittleEndian)
	}
	var canp int32
	r.Varint32(&canp)
	zz := ""
	if canp > 0 {
		for i := int32(0); i < canp; i++ {
			r.String(&zz)
		}
	}
	var cand int32
	r.Varint32(&cand)
	if cand > 0 {
		for i := int32(0); i < canp; i++ {
			r.String(&zz)
		}
	}
	return item.NewItem(iid, data, cnt, nbtdata)
}

func (r *PacketReader) EntityMetadata(x *map[uint32]interface{}) {
	*x = map[uint32]interface{}{}

	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, 256)
	for i := uint32(0); i < count; i++ {
		var key, dataType uint32
		r.Varuint32(&key)
		r.Varuint32(&dataType)
		switch dataType {
		case EntityDataByte:
			var v byte
			r.Uint8(&v)
			(*x)[key] = v
		case EntityDataInt16:
			var v int16
			r.Int16(&v)
			(*x)[key] = v
		case EntityDataInt32:
			var v int32
			r.Varint32(&v)
			(*x)[key] = v
		case EntityDataFloat32:
			var v float32
			r.Float32(&v)
			(*x)[key] = v
		case EntityDataString:
			var v string
			r.String(&v)
			(*x)[key] = v
		case EntityDataCompoundTag:
			
		case EntityDataBlockPos:
			
		case EntityDataInt64:
			var v int64
			r.Varint64(&v)
			(*x)[key] = v
		case EntityDataVec3:
			
		default:
			r.UnknownEnumOption(dataType, "entity metadata")

		}
		(*x)[key] = []interface{}{dataType, (*x)[key]}
	}
}

func (r *PacketReader) NBT(m *map[string]interface{}, encoding nbt.Encoding) {
	if err := nbt.NewDecoderWithEncoding(r.r, encoding).Decode(m); err != nil {
		r.panic(err)
	}
}

func (r *PacketReader) ByteRotation(x *float32) {
	g, _ := r.r.ReadByte()
	a := float32(g * (360 / 256))
	*x = a
}

func (r *PacketReader) NBTList(m *[]interface{}, encoding nbt.Encoding) {
	if err := nbt.NewDecoderWithEncoding(r.r, encoding).Decode(m); err != nil {
		r.panic(err)
	}
}

func (r *PacketReader) UUID(x *uuid.UUID) {
	b := make([]byte, 16)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}

	b = append(b[8:], b[:8]...)
	var arr [16]byte
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = b[j], b[i]
	}
	*x = arr
}

func (r *PacketReader) LimitUint32(value uint32, max uint32) {
	if max == math.MaxUint32 {
		
		max = 0
	}
	if value > max {
		r.panicf("uint32 %v exceeds maximum of %v", value, max)
	}
}

func (r *PacketReader) LimitInt32(value int32, min, max int32) {
	if value < min {
		r.panicf("int32 %v exceeds minimum of %v", value, min)
	} else if value > max {
		r.panicf("int32 %v exceeds maximum of %v", value, max)
	}
}

func (r *PacketReader) UnknownEnumOption(value interface{}, enum string) {
	r.panicf("unknown value '%v' for enum type '%v'", value, enum)
}

func (r *PacketReader) InvalidValue(value interface{}, forField, reason string) {
	r.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

var errVarIntOverflow = errors.New("varint overflows integer")

func (r *PacketReader) Varint64(x *int64) {
	var ux uint64
	for i := 0; i < 70; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		ux |= uint64(b&0x7f) << uint32(i)
		if b&0x80 == 0 {
			*x = int64(ux >> 1)
			if ux&1 != 0 {
				*x = ^*x
			}
			return
		}
	}
	
}

func (r *PacketReader) Varuint64(x *uint64) {
	var v uint64
	for i := 0; i < 70; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		v |= uint64(b&0x7f) << uint32(i)
		if b&0x80 == 0 {
			*x = v
			return
		}
	}
	
}

func (r *PacketReader) Varint32(x *int32) {
	var ux uint32
	for i := 0; i < 35; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		ux |= uint32(b&0x7f) << uint32(i)
		if b&0x80 == 0 {
			*x = int32(ux >> 1)
			if ux&1 != 0 {
				*x = ^*x
			}
			return
		}
	}
	
}

func (r *PacketReader) Varuint32(x *uint32) {
	var v uint32
	for i := 0; i < 35; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		v |= uint32(b&0x7f) << uint32(i)
		if b&0x80 == 0 {
			*x = v
			return
		}
	}
	
}

func (r *PacketReader) panicf(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a...))
}

func (r *PacketReader) panic(err error) {
	panic(err)
}
