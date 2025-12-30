package main

import (
	"testing"
	"reflect"
	"math"
)


// Create a new reflection using struct with a schema
type TestMessageReflection struct {
	m map[int]interface{}
	s map[int]ProtoType
}

type TestMessage struct {
	r TestMessageReflection
}

func (t TestMessageReflection) Put(i int, v ProtoValue) bool {
	t.m[i] = v.v
	t.s[i] = v.t
	return true
}

func (t TestMessageReflection) PutType(i int, v ProtoType) bool {
	t.s[i] = v
	return true
}


func (t TestMessageReflection) GetDescriptors() map[int]ProtoValue {
	merged_m := make(map[int]ProtoValue, 0)
	for k,v := range(t.m) {
		merged_m[k] = ProtoValue{v, t.s[k]}
	}
	return merged_m
}

func (t TestMessageReflection) GetValue(i int) ProtoValue {
	return ProtoValue{ t.m[i], t.s[i] }
}

func (t TestMessageReflection) GetFieldType(i int) ProtoType {
	return t.s[i]
}


func (t *TestMessage) GetReflection() Reflection {
	return t.r
}

func initTestMessage() *TestMessage {
	return &TestMessage{
		TestMessageReflection{
			m: map[int]interface{}{},
			s: map[int]ProtoType{},
		},
	}
}

func TestReadSingleRecord(t *testing.T) {
	data := []byte{0x80, 0x80, 0x80, 0x08, 0x01}
	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	expected_tm := initTestMessage()
	value := ProtoValue { int32(1), Int32 }
	expected_tm.GetReflection().Put(1, value)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestReadSingleRecordMultiBytePayload(t *testing.T) {
	data := []byte{0x80, 0x80, 0x80, 0x08, 0x81, 0x01}
	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	expected_tm := initTestMessage()
	value := ProtoValue{int32(129), Int32}
	expected_tm.GetReflection().Put(1, value)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestReadMultipleRecords(t *testing.T) {
	data := []byte{0x08, 0x81, 0x01, 0x10, 0x01}
	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	actual_tm.GetReflection().PutType(2, Int32)
	expected_tm := initTestMessage()
	value_1 := ProtoValue{int32(129), Int32}
	value_2 := ProtoValue{int32(1), Int32}
	expected_tm.GetReflection().Put(1, value_1)
	expected_tm.GetReflection().Put(2, value_2)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestWriteSingleRecord(t *testing.T) {
	expected_data := []byte{0x08, 0x01}
	tm := initTestMessage()
	value_1 := ProtoValue{int32(1), Int32}
	tm.GetReflection().Put(1, value_1)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}
}

func TestWriteMultipleRecords(t *testing.T) {
	expected_data := []byte{
		0x08, 0x01,
		0x10, 0x01 }
	tm := initTestMessage()
	value := ProtoValue{int32(1), Int32}
	tm.GetReflection().Put(1,value)
	tm.GetReflection().Put(2,value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}
}

func TestWriteThenReadSingleRecord(t *testing.T) {
	expected_data := []byte{
		0x08, 0x01,
		0x10, 0x01 }
	tm := initTestMessage()
	value := ProtoValue{int32(1), Int32}
	tm.GetReflection().Put(1,value)
	tm.GetReflection().Put(2,value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	actual_tm.GetReflection().PutType(2, Int32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(actual_tm, tm) {
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}
}

func TestWriteLenRecord(t *testing.T) {
	expected_data := []byte{
		0x09, 0x07,
		0x74, 0x65,
		0x73, 0x74,
		0x69, 0x6e, 0x67 }

	tm := initTestMessage()
	value := ProtoValue{ string("testing"), String}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}
}

func TestReadLenRecord(t *testing.T) {
	actual_data := []byte{
		0x09, 0x07,
		0x74, 0x65,
		0x73, 0x74,
		0x69, 0x6e, 0x67 }
	actual_tm := initTestMessage()
	expected_tm := initTestMessage()
	value := ProtoValue{"testing", String}
	expected_tm.GetReflection().Put(1,value)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(expected_tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestWriteMultipleTypesRecord(t *testing.T) {
	expected_data := []byte{
		0x09, 0x07,
		0x74, 0x65,
		0x73, 0x74,
		0x69, 0x6e, 0x67,
		0x10, 0x01 }

	tm := initTestMessage()
	value_1 := ProtoValue{"testing", String}
	tm.GetReflection().Put(1,value_1)
	value_2 := ProtoValue{int32(1), Int32}
	tm.GetReflection().Put(2, value_2)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}
}

func TestReadMultipleTypesRecord(t *testing.T) {
	actual_data := []byte{
		0x09, 0x07,
		0x74, 0x65,
		0x73, 0x74,
		0x69, 0x6e, 0x67,
		0x10, 0x01 }
	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, String)
	actual_tm.GetReflection().PutType(2, Int32)
	expected_tm := initTestMessage()
	value_1 := ProtoValue{"testing", String}
	expected_tm.GetReflection().Put(1,value_1)
	value_2 := ProtoValue{int32(1), Int32}
	expected_tm.GetReflection().Put(2, value_2)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(expected_tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestReadWriteMultipleTypesRecord(t *testing.T) {
	expected_data := []byte{
		0x09, 0x07,
		0x74, 0x65,
		0x73, 0x74,
		0x69, 0x6e, 0x67,
		0x10, 0x01 }
	tm := initTestMessage()
	value_1 := ProtoValue{"testing", String}
	tm.GetReflection().Put(1, value_1)
	value_2 := ProtoValue{int32(1), Int32}
	tm.GetReflection().Put(2, value_2)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, String)
	actual_tm.GetReflection().PutType(2, Int32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}
}

func TestWriteThenReadNegativeSignedIntegerRecord(t *testing.T) {
	expected_data := []byte{
		0x08, 0x01 }

	tm := initTestMessage()
	value := ProtoValue{ int32(-1), Sint32}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Sint32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}

}

func TestWriteThenReadSignedIntegerRecord(t *testing.T) {
	expected_data := []byte{
		0x08, 
		0x8f, 0xff, 0xff, 0xff, 0x7f,
		}

	tm := initTestMessage()
	value := ProtoValue{ int32(-1), Int32}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}

}

func TestWriteZeroInteger(t *testing.T) {
	expected_data := []byte{
		0x08, 0x00,
		}

	tm := initTestMessage()
	value := ProtoValue{ int32(0), Int32}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Int32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}
}

func TestWriteEmptyString(t *testing.T) {
	expected_data := []byte{
		0x09, 0x00,
		}

	tm := initTestMessage()
	value := ProtoValue{ "", String}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, String)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}
}

func TestWriteThenReadNegativeMinSignedIntegerRecord(t *testing.T) {
	expected_data := []byte{
		0x08, 0x01 }

	tm := initTestMessage()
	value := ProtoValue{ int32(math.MinInt32), Sint32}
	tm.GetReflection().Put(1, value)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	actual_tm.GetReflection().PutType(1, Sint32)
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}

}

