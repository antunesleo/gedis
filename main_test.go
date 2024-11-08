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
