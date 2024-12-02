package main

import (
	"bytes"
	"fmt"
	"net"
)

const SIMPLE_STRING_BYTE_NUMBER = 43 // +
const ERROR_STRING_BYTE_NUMBER = 45 // -
const BULK_STRING_BYTE_NUMBER = 36 // $
const ARRAY_STRING_BYTE_NUMBER = 42 // *
const CARRIAGE_RETURN_BYTE_NUMBER = 13 // \r
const LINE_FEED_BYTE_NUMBER = 10 // \n
var PING_BYTE_ARRAY = []byte("PING")
var ECHO_BYTE_ARRAY = []byte("ECHO")
var GET_BYTE_ARRAY = []byte("GET")
var EXISTS_BYTE_ARRAY = []byte("EXISTS")
var SET_BYTE_ARRAY = []byte("SET")
var DEL_BYTE_ARRAY = []byte("DEL")


var cache = make(map[string][]byte)

func splitFromStartIndexToCRLF(startIndex int, message []byte) []byte {
    var deserialized []byte 
    for startIndex < len(message){
        if message[startIndex] == CARRIAGE_RETURN_BYTE_NUMBER || message[startIndex] == LINE_FEED_BYTE_NUMBER {
            break
        }
        deserialized = append(deserialized, message[startIndex])
        startIndex += 1
    }
    return deserialized
}

func deserializeSimpleString(startIndex int, message []byte) [][]byte {
    return [][]byte{splitFromStartIndexToCRLF(startIndex, message)}
}

func deserializeError(startIndex int, message []byte) [][]byte {
    return  [][]byte{splitFromStartIndexToCRLF(startIndex, message)}
}


func deserializeBulkString(startIndex int, message []byte) [][]byte {
    moveForward := true

    index := startIndex
    for moveForward {
        if  isNumeric(message[index]) {
            index += 1
        } else {
            moveForward = false
        }
    }

    if message[index] == CARRIAGE_RETURN_BYTE_NUMBER && message[index+1] == LINE_FEED_BYTE_NUMBER {
        index += 2
    } else {
        // invalid
        return [][]byte{}
    }

    var stringBuffer []byte
    moveForward = true
    for moveForward {
        if message[index] == CARRIAGE_RETURN_BYTE_NUMBER && message[index+1] == LINE_FEED_BYTE_NUMBER {
            moveForward = false
        } else {
            stringBuffer = append(stringBuffer, message[index])
            index += 1
        }
    }
    if len(stringBuffer) == 0 {
        return [][]byte{}
    }
    return [][]byte{stringBuffer}
}

func isNumeric(c byte) bool {
    return (c >= '0' && c <= '9')
}


func deserializeArray(startIndex int, message []byte) [][]byte {
    var deserializedArray [][]byte

    for startIndex < len(message) {
        if message[startIndex] == BULK_STRING_BYTE_NUMBER {
            startIndex += 1
            moveForward := true
            for moveForward {
                if isNumeric(message[startIndex]) {
                    startIndex += 1
                } else {
                    moveForward = false
                }
            }
            deserializedArrayItem := deserializeBulkString(startIndex, message)[0]
            deserializedArray = append(deserializedArray, deserializedArrayItem)
        }
        startIndex += 1
    }

    return deserializedArray
}

func deserialize(message []byte) [][]byte {
    if message[0] == ARRAY_STRING_BYTE_NUMBER {
        return deserializeArray(1, message)
    }
    if message[0] == SIMPLE_STRING_BYTE_NUMBER {
        return deserializeSimpleString(1, message)
    }
    if message[0] == ERROR_STRING_BYTE_NUMBER {
        return deserializeError(1, message)
    }
    if message[0] == BULK_STRING_BYTE_NUMBER {
        return deserializeBulkString(1, message)
    }
    return [][]byte{}
}

func serializeSimpleStringFromByteArray(message []byte) []byte {
    return []byte(fmt.Sprintf("+%s\r\n", message))
}

func serializeSimpleStringFromString(message string) []byte {
    return []byte(fmt.Sprintf("+%s\r\n", message))
}

func serializeError(message string) []byte {
    return []byte(fmt.Sprintf("-%s\r\n", message))
}

func serializerInteger(intToSerialize int) []byte {
    return []byte(fmt.Sprintf(":%d\r\n", intToSerialize))
}

func cmdPing() []byte {
    return serializeSimpleStringFromString("PONG")
}

func cmdEcho(message[][]byte) []byte {
    return serializeSimpleStringFromByteArray(message[1])
}

func cmdSet(message[][]byte) []byte {
    cache[string(message[1])] = message[2]
    return serializeSimpleStringFromString("OK")
}

func cmdGet(message[][]byte) []byte {
    value, ok := cache[string(message[1])]
    if ok {
        return serializeSimpleStringFromByteArray(value)
    }
    return serializeError("doesn't exist")
}

func cmdExists(message[][]byte) []byte {
    var existsCount = 0
    var itemIndex = 1
    for itemIndex < len(message) {
        _, ok := cache[string(message[itemIndex])]
        if ok {
            existsCount += 1
        }
        itemIndex += 1
    }

    return serializerInteger(existsCount)
}

func cmdDel(message[][]byte) []byte {
    var existsCount = 0
    var itemIndex = 1
    for itemIndex < len(message) {
        var key = string(message[itemIndex])
        _, ok := cache[key]
        if ok {
            delete(cache, key)
            existsCount += 1
        }
        itemIndex += 1
    }

    return serializerInteger(existsCount)   
}

func main() {
    listerner, err := net.Listen("tcp", "localhost:6379")
    if err != nil {
        fmt.Println(err)
        return
        // handle error
    }
    for {
        conn, err := listerner.Accept()
        if err != nil {
            // handle error
        }
        go handleConnection(conn)
    }    
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 8192)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        result := []byte{}
        message := deserialize(buffer[:n])

        if bytes.Equal(message[0], SET_BYTE_ARRAY) {
            result = cmdSet(message)
        } else if bytes.Equal(message[0], GET_BYTE_ARRAY) {
            result = cmdGet(message)
        } else if bytes.Equal(message[0], PING_BYTE_ARRAY) {
            result = cmdPing()
        } else if bytes.Equal(message[0], ECHO_BYTE_ARRAY) {
            result = cmdEcho(message)
        } else if bytes.Equal(message[0], EXISTS_BYTE_ARRAY) {
            result = cmdExists(message)
        } else if bytes.Equal(message[0], DEL_BYTE_ARRAY) {
            result = cmdDel(message)
        }
        
        if len(result) == 0 {
            result = serializeError("not implemented")
        }

        _, err = conn.Write(result)
        if err != nil {
            fmt.Println("Error:", err)
        }
    }
}