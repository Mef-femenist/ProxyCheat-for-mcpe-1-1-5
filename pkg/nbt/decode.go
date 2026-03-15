package nbt

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"reflect"
	"strings"
	"sync"
)

type Decoder struct {
	
	Encoding Encoding

	r     *offsetReader
	depth int
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{Encoding: NetworkLittleEndian, r: newOffsetReader(r)}
}

func NewDecoderWithEncoding(r io.Reader, encoding Encoding) *Decoder {
	return &Decoder{Encoding: encoding, r: newOffsetReader(r)}
}

func (d *Decoder) Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return NonPointerTypeError{ActualType: val.Type()}
	}
	tagType, tagName, err := d.tag()
	if err != nil {
		return err
	}
	return d.unmarshalTag(val.Elem(), tagType, tagName)
}

func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalEncoding(data, v, NetworkLittleEndian)
}

func UnmarshalEncoding(data []byte, v interface{}, encoding Encoding) error {
	buf := bytes.NewBuffer(data)
	return (&Decoder{Encoding: encoding, r: &offsetReader{
		Reader:   buf,
		ReadByte: buf.ReadByte,
		Next:     buf.Next,
	}}).Decode(v)
}

var stringType = reflect.TypeOf("")
var byteType = reflect.TypeOf(byte(0))
var int32Type = reflect.TypeOf(int32(0))
var int64Type = reflect.TypeOf(int64(0))

var fieldMapPool = sync.Pool{
	New: func() interface{} {
		return map[string]reflect.Value{}
	},
}

func (d *Decoder) unmarshalTag(val reflect.Value, tagType byte, tagName string) error {
	switch tagType {
	default:
		return UnknownTagError{Off: d.r.off, TagType: tagType, Op: "Match"}
	case tagEnd:
		return UnexpectedTagError{Off: d.r.off, TagType: tagEnd}

	case tagByte:
		value, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Byte"}
		}
		if val.Kind() != reflect.Uint8 {
			if val.Kind() == reflect.Bool {
				if value != 0 {
					val.SetBool(true)
				}
				return nil
			}
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetUint(uint64(value))

	case tagInt16:
		value, err := d.Encoding.Int16(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int16 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(int64(value))

	case tagInt32:
		value, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(int64(value))

	case tagInt64:
		value, err := d.Encoding.Int64(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(value)

	case tagFloat32:
		value, err := d.Encoding.Float32(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Float32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetFloat(float64(value))

	case tagFloat64:
		value, err := d.Encoding.Float64(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Float64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetFloat(value)

	case tagString:
		value, err := d.Encoding.String(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.String {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetString(value)

	case tagByteArray:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		data, err := consumeN(int(length), d.r)
		if err != nil {
			return BufferOverrunError{Op: "ByteArray"}
		}
		value := reflect.New(reflect.ArrayOf(int(length), byteType)).Elem()
		for i := int32(0); i < length; i++ {
			value.Index(int(i)).SetUint(uint64(data[i]))
		}
		if val.Kind() != reflect.Array {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "ByteArray", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagInt32Array:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		value := reflect.New(reflect.ArrayOf(int(length), int32Type)).Elem()
		for i := int32(0); i < length; i++ {
			v, err := d.Encoding.Int32(d.r)
			if err != nil {
				return err
			}
			value.Index(int(i)).SetInt(int64(v))
		}
		if val.Kind() != reflect.Array || val.Type().Elem().Kind() != reflect.Int32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "Int32Array", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagInt64Array:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		value := reflect.New(reflect.ArrayOf(int(length), int64Type)).Elem()
		for i := int32(0); i < length; i++ {
			v, err := d.Encoding.Int64(d.r)
			if err != nil {
				return err
			}
			value.Index(int(i)).SetInt(v)
		}
		if val.Kind() != reflect.Array || val.Type().Elem().Kind() != reflect.Int64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "Int64Array", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagSlice:
		d.depth++
		listType, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "List"}
		}
		if !tagExists(listType) {
			return UnknownTagError{Off: d.r.off, TagType: listType, Op: "Slice"}
		}
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		valType := val.Type()
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Interface {
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Kind() == reflect.Interface {
			valType = reflect.SliceOf(valType)
		}
		v := reflect.MakeSlice(valType, int(length), int(length))
		if length != 0 {
			for i := 0; i < int(length); i++ {
				if err := d.unmarshalTag(v.Index(i), listType, ""); err != nil {
					
					if _, ok := err.(InvalidTypeError); ok {
						return InvalidTypeError{Off: d.r.off, FieldType: valType.Elem(), Field: fmt.Sprintf("%v[%v]", tagName, i), TagType: listType}
					}
					return err
				}
			}
		}
		val.Set(v)
		d.depth--

	case tagStruct:
		d.depth++
		switch val.Kind() {
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		case reflect.Struct:
			
			fields := fieldMapPool.Get().(map[string]reflect.Value)
			d.populateFields(val, fields)
			for {
				nestedTagType, nestedTagName, err := d.tag()
				if err != nil {
					return err
				}
				if nestedTagType == tagEnd {
					
					break
				}
				if !tagExists(nestedTagType) {
					return UnknownTagError{Off: d.r.off, Op: "Struct", TagType: nestedTagType}
				}
				field, ok := fields[nestedTagName]
				if ok {
					if err = d.unmarshalTag(field, nestedTagType, nestedTagName); err != nil {
						return err
					}
					continue
				}
				
				return UnexpectedNamedTagError{Off: d.r.off, TagName: tagName + "." + nestedTagName, TagType: nestedTagType}
			}
			
			for k := range fields {
				delete(fields, k)
			}
			fieldMapPool.Put(fields)
		case reflect.Interface, reflect.Map:
			if vk := val.Kind(); vk == reflect.Interface && val.NumMethod() != 0 {
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
			}
			valType := val.Type()
			if val.Kind() == reflect.Map {
				valType = valType.Elem()
			}
			m := reflect.MakeMap(reflect.MapOf(stringType, valType))
			for {
				nestedTagType, nestedTagName, err := d.tag()
				if err != nil {
					return err
				}
				if !tagExists(nestedTagType) {
					return UnknownTagError{Off: d.r.off, Op: "Map", TagType: nestedTagType}
				}
				if nestedTagType == tagEnd {
					
					break
				}
				value := reflect.New(valType).Elem()
				if err := d.unmarshalTag(value, nestedTagType, nestedTagName); err != nil {
					return err
				}
				m.SetMapIndex(reflect.ValueOf(nestedTagName), value)
			}
			val.Set(m)
		}
		d.depth--
	}
	return nil
}

func (d *Decoder) populateFields(val reflect.Value, m map[string]reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		if !ast.IsExported(fieldType.Name) {
			
			continue
		}
		field := val.Field(i)
		name := fieldType.Name
		if fieldType.Anonymous {
			
			d.populateFields(field, m)
			continue
		}
		if tag, ok := fieldType.Tag.Lookup("nbt"); ok {
			if tag == "-" {
				continue
			}
			tag = strings.TrimSuffix(tag, ",omitempty")
			if tag != "" {
				name = tag
			}
		}
		m[name] = field
	}
}

func (d *Decoder) tag() (tagType byte, tagName string, err error) {
	if d.depth >= maximumNestingDepth {
		return 0, "", MaximumDepthReachedError{}
	}
	if d.r.off >= maximumNetworkOffset && d.Encoding == NetworkLittleEndian {
		return 0, "", MaximumBytesReadError{}
	}
	tagType, err = d.r.ReadByte()
	if err != nil {
		return 0, "", BufferOverrunError{Op: "ReadTag"}
	}
	if tagType != tagEnd {
		
		tagName, err = d.Encoding.String(d.r)
	}
	return
}
