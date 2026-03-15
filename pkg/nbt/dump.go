package nbt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Dump(data []byte, encoding Encoding) (string, error) {
	var m map[string]interface{}
	if err := UnmarshalEncoding(data, &m, encoding); err != nil {
		return "", fmt.Errorf("error decoding NBT: %v", err)
	}
	s := &dumpState{}
	return s.encodeTagType(m) + "(" + s.encodeTagValue(m) + ")", nil
}

type dumpState struct {
	
	currentIndent int
}

func (s *dumpState) indent() string {
	return strings.Repeat("	", s.currentIndent)
}

func (s *dumpState) encodeTagType(val interface{}) string {
	if val == nil {
		return "nil"
	}
	switch val.(type) {
	case byte:
		return "TAG_Byte"
	case int16:
		return "TAG_Short"
	case int32:
		return "TAG_Int"
	case int64:
		return "TAG_Long"
	case float32:
		return "TAG_Float"
	case float64:
		return "TAG_Double"
	case string:
		return "TAG_String"
	}
	t := reflect.TypeOf(val)
	switch t.Kind() {
	case reflect.Map:
		return "TAG_Compound"
	case reflect.Slice:
		elemType := reflect.New(t.Elem()).Elem().Interface()

		v := reflect.ValueOf(val)
		if v.Len() != 0 && elemType == nil {
			elemType = v.Index(0).Elem().Interface()
		}
		return "TAG_List<" + s.encodeTagType(elemType) + ">"
	case reflect.Array:
		switch t.Elem().Kind() {
		case reflect.Uint8, reflect.Bool:
			return "TAG_ByteArray"
		case reflect.Int32:
			return "TAG_IntArray"
		case reflect.Int64:
			return "TAG_LongArray"
		}
	}
	panic("should not happen")
}

func (s *dumpState) encodeTagValue(val interface{}) string {
	
	const hexTable = "0123456789abcdef"

	switch v := val.(type) {
	case byte:
		return "0x" + string([]byte{hexTable[v>>4], hexTable[v&0x0f]})
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case string:
		return v
	}
	t := reflect.TypeOf(val)
	reflectVal := reflect.ValueOf(val)
	switch t.Kind() {
	case reflect.Map:
		b := strings.Builder{}
		b.WriteString("{\n")
		for _, k := range reflectVal.MapKeys() {
			v := reflectVal.MapIndex(k)
			actualVal := v.Interface()

			s.currentIndent++
			b.WriteString(fmt.Sprintf("%v'%v': %v(%v),\n", s.indent(), k.String(), s.encodeTagType(actualVal), s.encodeTagValue(actualVal)))
			s.currentIndent--
		}
		b.WriteString(s.indent() + "}")
		return b.String()
	case reflect.Slice:
		b := strings.Builder{}
		b.WriteString("{\n")
		for i := 0; i < reflectVal.Len(); i++ {
			v := reflectVal.Index(i)
			actualVal := v.Interface()

			s.currentIndent++
			b.WriteString(fmt.Sprintf("%v%v,\n", s.indent(), s.encodeTagValue(actualVal)))
			s.currentIndent--
		}
		b.WriteString(s.indent() + "}")
		return b.String()
	case reflect.Array:
		switch t.Elem().Kind() {
		case reflect.Uint8:
			b := strings.Builder{}
			for i := 0; i < reflectVal.Len(); i++ {
				v := reflectVal.Index(i).Uint()
				b.WriteString("0x")
				b.WriteString(string([]byte{hexTable[v>>4], hexTable[v&0x0f]}))
				if i != reflectVal.Len()-1 {
					b.WriteByte(' ')
				}
			}
			return b.String()
		case reflect.Int32, reflect.Int64:
			b := strings.Builder{}
			for i := 0; i < reflectVal.Len(); i++ {
				v := reflectVal.Index(i).Int()
				b.WriteString(strconv.FormatInt(v, 10))
				if i != reflectVal.Len()-1 {
					b.WriteByte(' ')
				}
			}
			return b.String()
		}
	}
	panic("should not happen")
}
