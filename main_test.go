package main

import (
	"testing"
	"reflect"
)


// Create a new reflection using struct with a schema
type TestMessageReflection struct {
	m map[int]interface{}
}

type TestMessage struct {
	r TestMessageReflection
}

func (t TestMessageReflection) Put(i int, v interface{}) bool {
	t.m[i] = v
	return true
}

func (t TestMessageReflection) GetDescriptors() map[int]interface{} {
	return t.m
}

func (t TestMessageReflection) GetValue(i int) interface{} {
	return t.m[i]
}

func (t *TestMessage) GetReflection() Reflection {
	return t.r
}

func initTestMessage() *TestMessage {
	return &TestMessage{
		TestMessageReflection{
			m: map[int]interface{}{},
		},
	}
}

func TestReadSingleRecord(t *testing.T) {
	data := []byte{0x80, 0x80, 0x80, 0x08, 0x01}
	actual_tm := initTestMessage()
	expected_tm := initTestMessage()
	expected_tm.GetReflection().Put(1, 1)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestReadSingleRecordMultiBytePayload(t *testing.T) {
	data := []byte{0x80, 0x80, 0x80, 0x08, 0x81, 0x01}
	actual_tm := initTestMessage()
	expected_tm := initTestMessage()
	expected_tm.GetReflection().Put(1, 129)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestReadMultipleRecords(t *testing.T) {
	data := []byte{0x08, 0x81, 0x01, 0x10, 0x01}
	actual_tm := initTestMessage()
	expected_tm := initTestMessage()
	expected_tm.GetReflection().Put(1, 129)
	expected_tm.GetReflection().Put(2, 1)
	r := ProtoReader{}
	r.Read(data, actual_tm)

	if !reflect.DeepEqual(actual_tm, expected_tm) {
		t.Errorf("Expected %v. Actual %v", expected_tm, actual_tm)
	}
}

func TestWriteSingleRecord(t *testing.T) {
	expected_data := []byte{0x08, 0x01}
	tm := initTestMessage()
	tm.GetReflection().Put(1,1)
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
	tm.GetReflection().Put(1,1)
	tm.GetReflection().Put(2,1)
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
	tm.GetReflection().Put(1,1)
	tm.GetReflection().Put(2,1)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
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
	tm.GetReflection().Put(1,"testing")
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
	expected_tm.GetReflection().Put(1,"testing")
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
	tm.GetReflection().Put(1,"testing")
	tm.GetReflection().Put(2, 1)
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
	expected_tm := initTestMessage()
	expected_tm.GetReflection().Put(1,"testing")
	expected_tm.GetReflection().Put(2, 1)
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
	tm.GetReflection().Put(1,"testing")
	tm.GetReflection().Put(2, 1)
	w := ProtoWriter{}
	actual_data := w.Write(tm)

	if !reflect.DeepEqual(expected_data, actual_data) {
		t.Errorf("Expected %v. Actual %v", expected_data, actual_data)
	}

	actual_tm := initTestMessage()
	r := ProtoReader{}
	r.Read(actual_data, actual_tm)

	if !reflect.DeepEqual(tm, actual_tm){
		t.Errorf("Expected %v. Actual %v", tm, actual_tm)
	}
}

