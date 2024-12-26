package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDeserializeSimpleString(t *testing.T) {
	strMessage := "+hello world\r\n"
	byteArrMessage := []byte(strMessage)
	deserializer, _ := getDeserializer(byteArrMessage)
	got := deserializer.deserialize(byteArrMessage)
	want := [][]byte{[]byte("hello world")}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDeserializeError(t *testing.T) {
	strMessage := "-Error message\r\n"
	byteArrMessage := []byte(strMessage)
	deserializer, _ := getDeserializer(byteArrMessage)
	got := deserializer.deserialize(byteArrMessage)
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
		deserializer, _ := getDeserializer(tt.in)
		got := deserializer.deserialize(tt.in)
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
		deserializer, _ := getDeserializer(tt.in)
		got := deserializer.deserialize(tt.in)
		want := tt.out

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}

func TestBinaryAdventures(t *testing.T) {
    letter := 'a'
    fmt.Printf("Binary representation of '%c': %08b\n", letter, letter)

	number := 97
    fmt.Printf("Binary representation of %d: %08b\n", number, number)
	
    stringNumber := "97"
    fmt.Printf("Binary representation of \"%s\": ", stringNumber)
    for _, char := range stringNumber {
        fmt.Printf("%08b ", char)
    }
    fmt.Println()
}

func TestSaveAndRestoreSnapshot(t *testing.T) {
	var want = map[string][]byte{
		"a": []byte("2"),
		"b": []byte("3"),
	}
	err := saveSnapshot(want)
	if err != nil {
		t.Errorf("error on saving snapshot %e", err)
	}
	err, got := restoreSnapshot()
	if err != nil {
		t.Errorf("error on restoring snapshot %e", err)		
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}

}

// func TestCommandBufferExtractMessage(t *testing.T) {
// 	data := []byte("*1\r\n$4\r\nping\r\n")
// 	buffer := NewCommandBuffer()
// 	buffer.data = data
// 	err, got := buffer.extractMessage()
// 	if err != nil {
// 		t.Errorf("error on extracting message %e", err)
// 	}
// 	want := []byte("*1\r\n$4\r\nping\r\n")
// 	if !reflect.DeepEqual(got, want) {
// 		t.Errorf("got %q, wanted %q", got, want)
// 	}
// }

// var unserializableMessages = [][]byte {
// 	[]byte("1\r\n$4\r\nping\r\n"),
// 	[]byte("*2\r\n$4\r\nping\r\n"),
// 	[]byte("*1\r\n4\r\nping\r\n"),
// 	[]byte("*1\r\n$5\r\nping\r\n"),
// 	[]byte("*1$4\r\nping\r\n"),
// 	[]byte("*1\r\n$4ping\r\n"),
// 	[]byte("*1\r\n$4\r\nping"),
// }
// func TestCommandBufferExtractMessagePingReturnSerializationError(t *testing.T) {
// 	for _, unserializableMessage := range unserializableMessages {
// 		buffer := NewCommandBuffer()
// 		buffer.data = unserializableMessage
// 		err, _ := buffer.extractMessage()
// 		if err == nil {
// 			t.Error("serialization should have failed")
// 		}
// 	}
// }

func makeBufferWithData(data []byte) *CommandBuffer {
	buffer := NewCommandBuffer()
	for i := 0; i < len(data) && i < len(buffer.data); i++ {
		buffer.data[i] = data[i]
	}
	return &buffer
}


var messages = []struct {
	bufferData []byte
	extractedMessage []byte
} {
	{[]byte("+OK\r\n"), []byte("+OK\r\n")},
	{[]byte("aa+OK\r\n"), []byte("+OK\r\n")},
	{[]byte("+OK\r\naa"), []byte("+OK\r\n")},
}
func TestCommandBufferExtractSimpleStringMessage(t *testing.T) {
	for _, message := range messages {
		buffer := makeBufferWithData(message.bufferData)
		err, got := buffer.extractMessage()
		if err != nil {
			t.Errorf("error on extracting message %e", err)
		}
		want := message.extractedMessage
		fmt.Println("got", got)
		fmt.Println("want", want)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, wanted %q", got, want)
		}
	}
}


var unserializableMessages = [][]byte {
	[]byte("+OK"),
	[]byte("OK\r\n"),
}
func TestCommandBufferExtractSimpleStringMessageMustFailSerialization(t *testing.T) {
	for _, unserializableMessage := range unserializableMessages {
		buffer := NewCommandBuffer()
		buffer.data = unserializableMessage
		err, _ := buffer.extractMessage()
		if err == nil {
			t.Error("serialization should have failed")
		}
	}
}