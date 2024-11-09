package main

import (
	"reflect"
	"testing"
)

func TestDeserializeSimpleString(t *testing.T) {
	strMessage := "+hello world\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := [][]byte{[]byte("hello world")}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDeserializeError(t *testing.T) {
	strMessage := "-Error message\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := [][]byte{[]byte("Error message")}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}


var bulkstringtests = []struct {
	in []byte
	out [][]byte
}{
	{[]byte("$5\r\nhello𒓸\r\n"), [][]byte{[]byte("hello𒓸")}},
	{[]byte("$0\r\n\r\n"), [][]byte{}},
	{[]byte("$-1\r\n"), [][]byte{}},
}
func TestDeserializeBulkSring(t *testing.T) {
	for _, tt := range bulkstringtests {
		got := deserialize(tt.in)
		want := tt.out

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}

var arraytests = []struct {
	in []byte
	out [][]byte
}{
	{[]byte("*1\r\n$4\r\nping\r\n"), [][]byte{[]byte("ping")}},
	{[]byte("*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n"), [][]byte{[]byte("echo"), []byte("hello world")}},
	{[]byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"), [][]byte{[]byte("get"), []byte("key")}},
}
func TestDeserializeArray(t *testing.T) {
	for _, tt := range arraytests {
		got := deserialize(tt.in)
		want := tt.out

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}
