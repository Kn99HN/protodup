package main

import (
	"fmt"
	"slices"
	"errors"
	"strings"
)

type WireType int
type ParseOp int

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
	Put(i int, v interface{}) bool
	GetDescriptors() map[int]interface{}
	GetValue(i int) interface{}
}

type Reader interface {
	Read(buffer []byte, m Message)
}

type ProtoReader struct {}

type Writer interface {
	Write(m Message) []byte
}

type ProtoWriter struct {}

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

func ToTagType(v interface{}) (byte, error) {
	switch v.(type) {
		case int:
			return 0, nil
		case string:
			return 1, nil
	}
	return 0, errors.New("Tag Type is not supported")
}

func ToLen(v interface{}) ([]byte, error) {
	s, ok := v.(string)
	if !ok { return nil, errors.New("Payload has mismatched type. Expected string") }
	
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

func ToVarIntsGeneric(v interface{}) ([]byte, error) {
	i, ok := v.(int)
	if !ok { return nil, errors.New("Payload has mismatched type. Expected int") }
	return ToVarInts(i), nil
}


func ToVarInts(i int) []byte {
	buf := make([]byte, 0)
	for i > 0 {
		b := byte(i & 0x7F)
		if len(buf) > 1 {
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
				val, new_index := parseVarInts(buffer, next_index)
				next_index = new_index
				reflection.Put(field_number, val)
			case LEN:
				val, new_index := parseLen(buffer, next_index)
				next_index = new_index
				reflection.Put(field_number, val)
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
		val = int(byte(val) | (b & 0x7F))
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


