package server

import (
	"errors"
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

func makeSerializationBufferWithData(data []byte) *DeserializationBuffer {
	buffer := NewDeserializationBuffer()
	for i := 0; i < len(data) && i < len(buffer.data); i++ {
		buffer.data[i] = data[i]
	}
	return &buffer
}


func assertBufferData(t *testing.T, buffer *DeserializationBuffer, want []byte) {
	for i := 0; i < len(want); i++ {
		if buffer.data[i] != want[i] {
			t.Fatalf("got %q, wanted %q", buffer.data, want)
		}
	}
	fmt.Println("buffer.data", buffer.data)
	fmt.Println("want", want)

	for j := len(want); j < len(buffer.data); j++ {
		if buffer.data[j] != 0 {
			t.Fatalf("wanted 0 got %q on index %d", buffer.data[j], j)
		}
	}
}


func TestSerializationBufferSerializeSimpleString(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		extractedMessage []byte
		remainingBufferData []byte
	} {
		{"base case", []byte("+OK\r\n"), []byte("+OK\r\n"), []byte("")},
		{"noise in the begning", []byte("aa+OK\r\n"), []byte("+OK\r\n"), []byte("")},
		{"noise in the end", []byte("+OK\r\naa"), []byte("+OK\r\n"), []byte("aa")},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tt.bufferData)
			got, err := serializationBuffer.Dissipate()
			if err != nil {
				t.Errorf("error on extracting message %e", err)
			}
			want := tt.extractedMessage

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %q, wanted %q", got, want)
			}

			assertBufferData(t, serializationBuffer, tt.remainingBufferData)
		})
	}
}

func TestSerializationBufferSerializeBulkString(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		extractedMessage []byte
		remainingBufferData []byte
	} {
		{"base case", []byte("$5\r\nhello\r\n"), []byte("$5\r\nhello\r\n"), []byte("")},
		{"noise in the begning", []byte("aa$5\r\nhello\r\n"), []byte("$5\r\nhello\r\n"), []byte("")},
		{"noise in the end", []byte("$5\r\nhello\r\naa"), []byte("$5\r\nhello\r\n"), []byte("aa")},
		{"noise in the end crlf", []byte("$5\r\nhello\r\n\r\n"), []byte("$5\r\nhello\r\n"), []byte("\r\n")},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tt.bufferData)
			got, err := serializationBuffer.Dissipate()
			if err != nil {
				t.Errorf("error on extracting message %e", err)
			}
			want := tt.extractedMessage

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %q, wanted %q", got, want)
			}

			assertBufferData(t, serializationBuffer, tt.remainingBufferData)
		})
	}
}

func TestSerializationBufferSerializeArrayString(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		extractedMessage []byte
		remainingBufferData []byte
	} {
		{
			"hello world", 
			[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), 
			[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), 
			[]byte(""),
		},
		{
			"get key", 
			[]byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"), 
			[]byte("*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"), 
			[]byte(""),
		},
		{
			"set key", 
			[]byte("*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"), 
			[]byte("*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"), 
			[]byte(""),
		},
		{
			"noise in the begnning", 
			[]byte("aa*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), 
			[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), 
			[]byte(""),
		},
		{
			"noise in the end", 
			[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\naa"), 
			[]byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"), 
			[]byte("aa"),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tt.bufferData)
			got, err := serializationBuffer.Dissipate()
			if err != nil {
				t.Errorf("error on extracting message %e", err)
			}
			want := tt.extractedMessage

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %q, wanted %q", got, want)
			}

			assertBufferData(t, serializationBuffer, tt.remainingBufferData)
		})
	}
}

func TestSerializationBufferSerializeError(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		extractedMessage []byte
		remainingBufferData []byte
	} {
		{"base case", []byte("-SOMEERROR\r\n"), []byte("-SOMEERROR\r\n"), []byte("")},
		{"noise in the begning", []byte("aa-SOMEERROR\r\n"), []byte("-SOMEERROR\r\n"), []byte("")},
		{"noise in the end", []byte("-SOMEERROR\r\naa"), []byte("-SOMEERROR\r\n"), []byte("aa")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tc.bufferData)
			got, err := serializationBuffer.Dissipate()
			if err != nil {
				t.Errorf("error on extracting message %e", err)
			}
			want := tc.extractedMessage

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %q, wanted %q", got, want)
			}

			assertBufferData(t, serializationBuffer, tc.remainingBufferData)
		})
	}
}



func TestSerializationBufferSerializeSimpleStringMustFail(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		wantedError error
	} {
		{"No ending crlf", []byte("+OK"), errors.New("serialization errror: no crlf found")},
		// {"No first byte data type", []byte("OK\r\n"), errors.New("serialization error: unknown first byte data type")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tc.bufferData)
			_, err := serializationBuffer.Dissipate()
			if err.Error() != tc.wantedError.Error() {
				t.Errorf("wanted %q got %q", tc.wantedError, err)
			}
		})
	}
}

func TestSerializationBufferSerializeErrorMustFail(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		wantedError error
	} {
		{"No ending crlf", []byte("-SOMEERROR"), errors.New("serialization errror: no crlf found")},
		{"No first byte data type", []byte("SOMEERROR\r\n"), errors.New("serialization error: unknown first byte data type")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tc.bufferData)
			_, err := serializationBuffer.Dissipate()
			if err.Error() != tc.wantedError.Error() {
				t.Errorf("wanted %q got %q", tc.wantedError, err)
			}
		})
	}

}


func TestSerializationBufferSerializeBulkStringMustFail(t *testing.T) {
	var testcases = []struct {
		name string
		bufferData []byte
		wantedError error
	} {
		{"No ending crlf", []byte("$5\r\nhello"), errors.New("serialization errror: no crlf found")},
		{"No first byte data type", []byte("5\r\nhello"), errors.New("serialization error: unknown first byte data type")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			serializationBuffer := makeSerializationBufferWithData(tc.bufferData)
			_, err := serializationBuffer.Dissipate()
			if err.Error() != tc.wantedError.Error() {
				t.Errorf("wanted %q got %q", tc.wantedError, err)
			}
		})
	}
}