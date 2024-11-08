package main

import (
	"fmt"
	"strconv"
)

const SIMPLE_STRING_CH = "+"
const ERROR_STRING_CH = "-"
const BULK_STRING_CH = "$"
const CARRIAGE_RETURN_BYTE_NUMBER = 13 // \r
const LINE_FEED_BYTE_NUMBER = 10 // \n

func removeFirstByteAndBackslashes(message []byte) []byte {
    var deserialized []byte 
    var index = 1;
    for index < len(message){
        if message[index] == CARRIAGE_RETURN_BYTE_NUMBER || message[index] == LINE_FEED_BYTE_NUMBER {
            index += 1
            continue
        }
        deserialized = append(deserialized, message[index])
        index += 1
    }
    return deserialized
}

func deserializeSimpleString(message []byte) string {
    return string(removeFirstByteAndBackslashes(message))
}

func deserializeError(message []byte) string {
    return string(removeFirstByteAndBackslashes(message)) 
}

func deserializeBulkString(message []byte) string {
    var deserialized []byte 
    var index = 1;
    for index < len(message){
        _, err := strconv.Atoi(string(message[index]))
        if err == nil {
            index += 1
            continue
        }
        if message[index] == CARRIAGE_RETURN_BYTE_NUMBER || message[index] == LINE_FEED_BYTE_NUMBER {
            index += 1
            continue
        }
        deserialized = append(deserialized, message[index])
        index += 1
    }
    return string(deserialized)
}

func deserialize(message []byte) string {
    if string(message[0]) == SIMPLE_STRING_CH {
        return deserializeSimpleString(message)
    }
    if string(message[0]) == ERROR_STRING_CH {
        return deserializeError(message)
    }
    if string(message[0]) == BULK_STRING_CH {
        return deserializeBulkString(message)
    }
    return ""
}

func main() {
    fmt.Print("lets do it!")    
}