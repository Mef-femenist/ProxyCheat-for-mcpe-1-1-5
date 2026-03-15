package packet

import (
	"fmt"
	"mefproxy/pkg/math"
	"github.com/google/uuid"
	"mefproxy/pkg/item"
	"mefproxy/pkg/nbt"
	"reflect"
	"sort"

	"image/color"
	"io"
	"unsafe"
)

type PacketWriter struct {
	w interface {
		io.Writer
		io.ByteWriter
	}
	shieldID int32
}

var (
	entityDataByte        = EntityDataByte
	entityDataInt16       = EntityDataInt16
	entityDataInt32       = EntityDataInt32
	entityDataFloat32     = EntityDataFloat32
	entityDataString      = EntityDataString
	entityDataCompoundTag = EntityDataCompoundTag
	entityDataBlockPos    = EntityDataBlockPos
	entityDataInt64       = EntityDataInt64
	entityDataVec3        = EntityDataVec3
)

func NewWriter(w interface {
	io.Writer
	io.ByteWriter
}, shieldID int32) *PacketWriter {
	return &PacketWriter{w: w, shieldID: shieldID}
}
func (w *PacketWriter) EntityMetadata(x *map[uint32]interface{}) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	keys := make([]int, 0, l)
	for k := range *x {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		key := uint32(k)
		value := (*x)[key]
		w.Varuint32(&key)
		switch v := value.(type) {
		case byte:
			w.Varuint32(&entityDataByte)
			w.Uint8(&v)
		case int16:
			w.Varuint32(&entityDataInt16)
			w.Int16(&v)
		case int32:
			w.Varuint32(&entityDataInt32)
			w.Varint32(&v)
		case float32:
			w.Varuint32(&entityDataFloat32)
			w.Float32(&v)
		case string:
			w.Varuint32(&entityDataString)
			w.String(&v)
		case map[string]interface{}:
			w.Varuint32(&entityDataCompoundTag)
			w.NBT(&v, nbt.NetworkLittleEndian)
		case int64:
			w.Varuint32(&entityDataInt64)
			w.Varint64(&v)
		case mgl32.Vec3:
			w.Varuint32(&entityDataVec3)
			w.Vec3(&v)
		default:
			w.UnknownEnumOption(reflect.TypeOf(value), "entity metadata")
		}
	}
}

func (w *PacketWriter) Uint8(x *uint8) {
	_ = w.w.WriteByte(*x)
}

func (w *PacketWriter) ByteRotation(x *float32) {
	g := byte(*x / (float32(360) / float32(256)))
	_ = w.w.WriteByte(g)
}

func (w *PacketWriter) Bool(x *bool) {
	_ = w.w.WriteByte(*(*byte)(unsafe.Pointer(x)))
}

func (w *PacketWriter) StringUTF(x *string) {
	l := int16(len(*x))
	w.Int16(&l)
	_, _ = w.w.Write([]byte(*x))
}

func (w *PacketWriter) String(x *string) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	_, _ = w.w.Write([]byte(*x))
}

func (w *PacketWriter) ByteSlice(x *[]byte) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	_, _ = w.w.Write(*x)
}

func (w *PacketWriter) Bytes(x *[]byte) {
	_, _ = w.w.Write(*x)
}

func (w *PacketWriter) BytesCopy(x []byte) {
	_, _ = w.w.Write(x)
}

func (w *PacketWriter) ByteFloat(x *float32) {
	_ = w.w.WriteByte(byte(*x / (360.0 / 256.0)))
}

func (w *PacketWriter) Vec3(x *mgl32.Vec3) {
	w.Float32(&x[0])
	w.Float32(&x[1])
	w.Float32(&x[2])
}

func (w *PacketWriter) Vec2(x *mgl32.Vec2) {
	w.Float32(&x[0])
	w.Float32(&x[1])
}

func (w *PacketWriter) VarRGBA(x *color.RGBA) {
	val := uint32(x.R) | uint32(x.G)<<8 | uint32(x.B)<<16 | uint32(x.A)<<24
	w.Varuint32(&val)
}

func (w *PacketWriter) UUID(x *uuid.UUID) {
	b := append((*x)[8:], (*x)[:8]...)
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	_, _ = w.w.Write(b)
}

func (w *PacketWriter) Varint64(x *int64) {
	u := *x
	ux := uint64(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		_ = w.w.WriteByte(byte(ux) | 0x80)
		ux >>= 7
	}
	_ = w.w.WriteByte(byte(ux))
}

func (w *PacketWriter) Varuint64(x *uint64) {
	u := *x
	for u >= 0x80 {
		_ = w.w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	_ = w.w.WriteByte(byte(u))
}

func (w *PacketWriter) Varint32(x *int32) {
	u := *x
	ux := uint32(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		_ = w.w.WriteByte(byte(ux) | 0x80)
		ux >>= 7
	}
	_ = w.w.WriteByte(byte(ux))
}

func (w *PacketWriter) Varuint32(x *uint32) {
	u := *x
	for u >= 0x80 {
		_ = w.w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	_ = w.w.WriteByte(byte(u))
}

func (w *PacketWriter) Item(it *item.Item) {
	if it.GetID() == 0 {
		w.Varint32(&it.ID)
		return
	}
	w.Varint32(&it.ID)
	aux := ((it.GetDamage() & 0x7fff) << 8) | it.GetCount()
	w.Varint32(&aux)

	nbtd := it.NBT
	nlen := int16(len(nbtd))
	if nlen != 0 {
		w.Int16(&nlen)
		w.NBT(&nbtd, nbt.LittleEndian)
	} else {
		w.Int16(&nlen)
	}

	todo := int32(0)
	w.Varint32(&todo) 
	w.Varint32(&todo) 

}

func (w *PacketWriter) NBT(x *map[string]interface{}, encoding nbt.Encoding) {
	if err := nbt.NewEncoderWithEncoding(w.w, encoding).Encode(*x); err != nil {
		panic(err)
	}
}

func (w *PacketWriter) NBTList(x *[]interface{}, encoding nbt.Encoding) {
	if err := nbt.NewEncoderWithEncoding(w.w, encoding).Encode(*x); err != nil {
		panic(err)
	}
}

func (w *PacketWriter) BlockCoords(x *int32, y *uint32, z *int32) {
	w.Varint32(x)
	w.Varuint32(y)
	w.Varint32(z)
}

func (w *PacketWriter) UnknownEnumOption(value interface{}, enum string) {
	w.panicf("unknown value '%v' for enum type '%v'", value, enum)
}

func (w *PacketWriter) InvalidValue(value interface{}, forField, reason string) {
	w.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

func (w *PacketWriter) panicf(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a...))
}
