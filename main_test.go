package main

import (
	"reflect"
	"testing"
)

func TestDeserializeSimpleString(t *testing.T) {
	strMessage := "+hello world\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := []string{"hello world"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDeserializeError(t *testing.T) {
	strMessage := "-Error message\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := []string{"Error message"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}


func TestDeserializeBulkSring(t *testing.T) {
	strMessage := "$5\r\nhello𒓸\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := []string{"hello𒓸"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}


func TestDeserializeArray(t *testing.T) {
	strMessage := "*2\r\n$4\r\necho\r\n$11\r\nhello world\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := []string{"echo", "hello world"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
