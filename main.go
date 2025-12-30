package main

import (
	"fmt"
	"slices"
	"errors"
	"strings"
)

type WireType int
type ParseOp int
type ProtoType int

const (
	String ProtoType = iota
	Uint32
	Uint64
	Sint32
	Sint64
	Int32
	Int64
)

const (
	VarInt WireType = iota
	LEN
)

const (
	Tag ParseOp = iota
	Payload
)

var globalMap = [2]WireType{ VarInt, LEN }

func main() {
}

type Message interface {
	GetReflection() Reflection
}

type Reflection interface {
	Put(i int, v ProtoValue) bool
	PutType(i int, v ProtoType) bool
	GetDescriptors() map[int]ProtoValue
	GetValue(i int) ProtoValue
	GetFieldType(i int) ProtoType
}

type Reader interface {
	Read(buffer []byte, m Message)
}

type ProtoReader struct {}

type Writer interface {
	Write(m Message) []byte
}

type ProtoWriter struct {}

type ProtoValue struct {
	v interface{}
	t ProtoType
}

func (w ProtoWriter) Write(m Message) []byte {
	reflection := m.GetReflection()
	descriptors := reflection.GetDescriptors()
	keys := make([]int, len(descriptors))
	i := 0
	for k := range(descriptors) {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	buf := make([]byte, 0)
	for _, index := range(keys) {
		val := descriptors[index]
		b := ToVarInts(index << 3)
		buf = append(buf, b...)
		tag_byte, err := ToTagType(val)
		if err != nil {
			panic(err)
		}
		last_byte_index := len(buf) - 1
		buf[last_byte_index] = buf[last_byte_index] | tag_byte
		tag_type := globalMap[tag_byte]
		payload := reflection.GetValue(index)
		switch tag_type {
			case VarInt:
				p_b, err := ToVarIntsGeneric(payload)
				if err != nil {
					panic(fmt.Sprintf("Payload does not match tag type %v", tag_type))
				}
				buf = append(buf, p_b...)
			case LEN:
				p_b, err := ToLen(payload)
				if err != nil {
					panic(fmt.Sprintf("Payload does not match tag type %v", tag_type))
				}
				buf = append(buf, p_b...)
		}


	}
	return buf
}

func ToTagType(v ProtoValue) (byte, error) {
	switch v.t {
		case Sint32:
			fallthrough
		case Sint64:
			fallthrough
		case Uint32:
			fallthrough
		case Uint64:
			fallthrough
		case Int32:
			fallthrough
		case Int64:
			return 0, nil
		case String:
			return 1, nil
	}
	return 0, errors.New("Tag Type is not supported")
}

func ToLen(v ProtoValue) ([]byte, error) {
	s, ok := v.v.(string)
	if !ok { return nil, errors.New("Payload has mismatched type. Expected string") }
	if s == "" {
		return []byte{0}, nil
	}
	
	record_length := len(s)
	length := ToVarInts(record_length)
	buf := make([]byte, record_length + len(length))
	index := 0
	for _,v := range(length) {
		buf[index] = v
		index++
	}
	for _, c := range(s) {
		buf[index] = byte(c)
		index++
	}
	return buf, nil
}

func ToVarIntsGeneric(v ProtoValue) ([]byte, error) {
	switch v.t {
		case Uint32:
			i, ok := v.v.(uint32)
			if !ok { return nil,
			errors.New("Payload has mismatched type. Expected uint32") }
			return ToVarInts(i), nil
		case Uint64:
			i, ok := v.v.(uint64)
			if !ok { return nil,
				errors.New("Payload has mismatched type. Expected uint32") }
			return ToVarInts(i), nil
		case Int32:
			i, ok := v.v.(int32)
			new_integer := uint32(i)
			if !ok { return nil, 
				errors.New("Payload has mismatched type. Expected uint32") }
			return ToVarInts(new_integer), nil
		case Int64:
			i, ok := v.v.(int64)
			new_integer := uint64(i)
			if !ok { return nil, 
			errors.New("Payload has mismatched type. Expected uint32") }
			return ToVarInts(new_integer), nil
		case Sint32:
			i, ok := v.v.(int32)
			if !ok { return nil, 
				errors.New("Payload has mismatched type. Expected uint32") }
			i = (i << 1) ^ (i >> 31)
			new_integer := uint32(i)
			return ToVarInts(new_integer), nil
		case Sint64:
			i, ok := v.v.(int64)
			if !ok { return nil,
				errors.New("Payload has mismatched type. Expected uint32") }
			i = (i << 1) ^ (i >> 63)
			new_integer := uint64(i)
			return ToVarInts(new_integer), nil
	}
	return nil,
		errors.New("Payload type does not match any of the supported types")
}

func ToVarInts [T ~uint32| ~uint64 | ~int32 | ~int64 | ~int] (i T) []byte {
	buf := make([]byte, 0)
	if i == 0 {
		return []byte{0}
	}
	for i != 0 {
		b := byte(i & 0x7F)
		if len(buf) >= 1 {
			b = b | 0x80
		}
		buf = append([]byte{b}, buf...)
		i = i >> 7
	}
	return buf
}

func (r ProtoReader) Read(buffer []byte, m Message) {
	reflection := m.GetReflection()
	next_index := 0
	for next_index < len(buffer) {
		tag, new_index := parseVarInts(buffer, next_index)
		next_index = new_index
		wire_type := tag & 0x07
		field_number := tag >> 3

		payload_wire_type := globalMap[int(wire_type)]
		switch payload_wire_type {
			case VarInt:
				int_type := reflection.GetFieldType(field_number)
				val, new_index := parseVarInts(buffer, next_index)
				next_index = new_index
				var value ProtoValue
				switch int_type {
					case Uint32:
						value = ProtoValue{ uint32(val), int_type }
					case Uint64:
						value = ProtoValue { uint64(val), int_type }
					case Int32:
						value = ProtoValue { int32(val), int_type }
					case Int64:
						value = ProtoValue { int64(val), int_type }
					case Sint32:
						parsed_val := uint32(val)
						converted_val := int32((parsed_val >> 1) ^ -(parsed_val & 1))
						value = ProtoValue { converted_val , int_type }
					case Sint64:
						parsed_val := int64(val)
						converted_val := int64((parsed_val >> 1) ^ -(parsed_val & 1))
						value = ProtoValue { converted_val, int_type }
				}
				reflection.Put(field_number, value)
			case LEN:
				val, new_index := parseLen(buffer, next_index)
				next_index = new_index
				reflection.Put(field_number, ProtoValue{ val, String } )
		}
	}
}

func getWireType(tag byte) int {
	return int(tag & 0x03)
}

func parseVarInts(buffer []byte, starting_index int) (int, int) {
	continuation := true
	val := 0
	for continuation {
		val = val << 7
		b := buffer[starting_index]
		continuation = ((b & 0x80) == 0x80)
		if starting_index == len(buffer) && continuation {
			panic("Reached the end of byte stream but there are still more bytes to decode")
		}
		starting_index += 1
		val = val | int(b & 0x7F)
	}
	return val, starting_index
}

func parseLen(buffer []byte, starting_index int) (string, int) {
	record_length, next_index := parseVarInts(buffer, starting_index)
	var sb strings.Builder
	for i := 0; i < record_length; i++ {
		sb.WriteRune(rune(buffer[next_index]))
		next_index++
	}
	return sb.String(), next_index
}


