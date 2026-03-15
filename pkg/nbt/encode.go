package nbt

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
)

type Encoder struct {
	
	Encoding Encoding

	w     *offsetWriter
	depth int
}

func NewEncoder(w io.Writer) *Encoder {
	var writer *offsetWriter
	if byteWriter, ok := w.(io.ByteWriter); ok {
		writer = &offsetWriter{Writer: w, WriteByte: byteWriter.WriteByte}
	} else {
		writer = &offsetWriter{Writer: w, WriteByte: func(b byte) error {
			_, err := w.Write([]byte{b})
			return err
		}}
	}
	return &Encoder{w: writer, Encoding: NetworkLittleEndian}
}

func NewEncoderWithEncoding(w io.Writer, encoding Encoding) *Encoder {
	enc := NewEncoder(w)
	enc.Encoding = encoding
	return enc
}

func (e *Encoder) Encode(v interface{}) error {
	val := reflect.ValueOf(v)
	return e.marshal(val, "")
}

func Marshal(v interface{}) ([]byte, error) {
	return MarshalEncoding(v, NetworkLittleEndian)
}

func MarshalEncoding(v interface{}, encoding Encoding) ([]byte, error) {
	b := bufferPool.Get().(*bytes.Buffer)
	err := (&Encoder{w: &offsetWriter{Writer: b, WriteByte: b.WriteByte}, Encoding: encoding}).Encode(v)
	data := append([]byte(nil), b.Bytes()...)

	b.Reset()
	bufferPool.Put(b)
	return data, err
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

func (e *Encoder) marshal(val reflect.Value, tagName string) error {
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	tagType := tagFromType(val.Type())
	if tagType == math.MaxUint8 {
		return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
	}
	if err := e.writeTag(tagType, tagName); err != nil {
		return err
	}
	return e.encode(val, tagName)
}

func (e *Encoder) encode(val reflect.Value, tagName string) error {
	kind := val.Kind()
	if kind == reflect.Interface {
		val = val.Elem()
		kind = val.Kind()
	}
	switch vk := kind; vk {
	case reflect.Uint8:
		return e.w.WriteByte(byte(val.Uint()))

	case reflect.Bool:
		if val.Bool() {
			return e.w.WriteByte(1)
		}
		return e.w.WriteByte(0)

	case reflect.Int16:
		return e.Encoding.WriteInt16(e.w, int16(val.Int()))

	case reflect.Int32:
		return e.Encoding.WriteInt32(e.w, int32(val.Int()))

	case reflect.Int64:
		return e.Encoding.WriteInt64(e.w, val.Int())

	case reflect.Float32:
		return e.Encoding.WriteFloat32(e.w, float32(val.Float()))

	case reflect.Float64:
		return e.Encoding.WriteFloat64(e.w, val.Float())

	case reflect.Array:
		switch val.Type().Elem().Kind() {

		case reflect.Uint8:
			n := val.Cap()
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			data := make([]byte, n)
			for i := 0; i < n; i++ {
				data[i] = byte(val.Index(i).Uint())
			}
			if _, err := e.w.Write(data); err != nil {
				return FailedWriteError{Op: "WriteByteArray", Off: e.w.off}
			}
			return nil

		case reflect.Int32:
			n := val.Cap()
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Encoding.WriteInt32(e.w, int32(val.Index(i).Int())); err != nil {
					return err
				}
			}

		case reflect.Int64:
			n := val.Cap()
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Encoding.WriteInt64(e.w, val.Index(i).Int()); err != nil {
					return err
				}
			}
		}

	case reflect.String:
		return e.Encoding.WriteString(e.w, val.String())

	case reflect.Slice:
		e.depth++
		elemType := val.Type().Elem()
		if elemType.Kind() == reflect.Interface {
			if val.Len() == 0 {
				
				elemType = byteType
			} else {
				
				elemType = val.Index(0).Elem().Type()
			}
		}

		listType := tagFromType(elemType)
		if listType == math.MaxUint8 {
			return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
		}
		if err := e.w.WriteByte(listType); err != nil {
			return FailedWriteError{Off: e.w.off, Op: "WriteSlice", Err: err}
		}
		if err := e.Encoding.WriteInt32(e.w, int32(val.Len())); err != nil {
			return err
		}
		for i := 0; i < val.Len(); i++ {
			nestedValue := val.Index(i)
			if err := e.encode(nestedValue, ""); err != nil {
				return err
			}
		}
		e.depth--

	case reflect.Struct:
		e.depth++
		if err := e.writeStructValues(val); err != nil {
			return err
		}
		e.depth--
		return e.w.WriteByte(tagEnd)

	case reflect.Map:
		e.depth++
		if val.Type().Key().Kind() != reflect.String {
			return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
		}
		iter := val.MapRange()
		for iter.Next() {
			if err := e.marshal(iter.Value(), iter.Key().String()); err != nil {
				return err
			}
		}
		e.depth--
		return e.w.WriteByte(tagEnd)
	}
	return nil
}

func (e *Encoder) writeStructValues(val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		fieldValue := val.Field(i)
		tag := fieldType.Tag.Get("nbt")
		if fieldType.PkgPath != "" || tag == "-" {
			
			continue
		}
		if fieldType.Anonymous {
			
			if err := e.writeStructValues(fieldValue); err != nil {
				return err
			}
			continue
		}
		tagName := fieldType.Name
		if strings.HasSuffix(tag, ",omitempty") {
			tag = strings.TrimSuffix(tag, ",omitempty")
			if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()) {
				
				continue
			}
		}
		if tag != "" {
			tagName = tag
		}
		if err := e.marshal(fieldValue, tagName); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeTag(tagType byte, tagName string) error {
	if e.depth >= maximumNestingDepth {
		return MaximumDepthReachedError{}
	}
	if err := e.w.WriteByte(tagType); err != nil {
		return err
	}
	return e.Encoding.WriteString(e.w, tagName)
}
