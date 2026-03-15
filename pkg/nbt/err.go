package nbt

import (
	"fmt"
	"reflect"
)

type InvalidTypeError struct {
	Off       int64
	Field     string
	TagType   byte
	FieldType reflect.Type
}

func (err InvalidTypeError) Error() string {
	return fmt.Sprintf("nbt: invalid type for tag '%v' at offset %v: cannot unmarshalTag %v into %v", err.Field, err.Off, tagName(err.TagType), err.FieldType)
}

type UnknownTagError struct {
	Off     int64
	Op      string
	TagType byte
}

func (err UnknownTagError) Error() string {
	return fmt.Sprintf("nbt: unknown tag '%v' at offset %v during op '%v'", err.TagType, err.Off, err.Op)
}

type UnexpectedTagError struct {
	Off     int64
	TagType byte
}

func (err UnexpectedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected tag %v at offset %v: tag is not valid in its context", tagName(err.TagType), err.Off)
}

type NonPointerTypeError struct {
	ActualType reflect.Type
}

func (err NonPointerTypeError) Error() string {
	return fmt.Sprintf("nbt: expected ptr type to decode into, but got '%v'", err.ActualType)
}

type BufferOverrunError struct {
	Op string
}

func (err BufferOverrunError) Error() string {
	return fmt.Sprintf("nbt: unexpected buffer end during op: '%v'", err.Op)
}

type InvalidArraySizeError struct {
	Off       int64
	Op        string
	GoLength  int
	NBTLength int
}

func (err InvalidArraySizeError) Error() string {
	return fmt.Sprintf("nbt: mismatched array size at %v during op '%v': expected size %v, found %v in NBT", err.Off, err.Op, err.GoLength, err.NBTLength)
}

type UnexpectedNamedTagError struct {
	Off     int64
	TagName string
	TagType byte
}

func (err UnexpectedNamedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected named tag '%v' with type %v at offset %v: not present in struct to be decoded into", err.TagName, tagName(err.TagType), err.Off)
}

type FailedWriteError struct {
	Off int64
	Op  string
	Err error
}

func (err FailedWriteError) Error() string {
	return fmt.Sprintf("nbt: failed write during op '%v' at offset %v: %v", err.Op, err.Off, err.Err)
}

type IncompatibleTypeError struct {
	ValueName string
	Type      reflect.Type
}

func (err IncompatibleTypeError) Error() string {
	return fmt.Sprintf("nbt: value type %v (%v) cannot be translated to an NBT tag", err.Type, err.ValueName)
}

type InvalidStringError struct {
	Off    int64
	Err    error
	String string
}

func (err InvalidStringError) Error() string {
	return fmt.Sprintf("nbt: string at offset %v is not valid: %v (%v)", err.Off, err.Err, err.String)
}

const maximumNestingDepth = 512

type MaximumDepthReachedError struct {
}

func (err MaximumDepthReachedError) Error() string {
	return fmt.Sprintf("nbt: maximum nesting depth of %v was reached", maximumNestingDepth)
}

const maximumNetworkOffset = 4 * 1024 * 1024

type MaximumBytesReadError struct {
}

func (err MaximumBytesReadError) Error() string {
	return fmt.Sprintf("nbt: limit of bytes read %v with NetworkLittleEndian format exhausted", maximumNetworkOffset)
}
