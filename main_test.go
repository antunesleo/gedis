package main

import "testing"

func TestDeserializeSimpleString(t *testing.T) {
	strMessage := "+hello world\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := "hello world"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDeserializeError(t *testing.T) {
	strMessage := "-Error message\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := "Error message"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}


func TestDeserializeBulkSring(t *testing.T) {
	strMessage := "$5\r\nhello\r\n"
	byteArrMessage := []byte(strMessage)
	got := deserialize(byteArrMessage)
	want := "hello"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
