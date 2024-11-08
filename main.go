package main

import (
	"fmt"
	"strconv"
)

const SIMPLE_STRING_CH = "+"
const ERROR_STRING_CH = "-"
const BULK_STRING_CH = "$"
const ARRAY_STRING_CH = "*"
const CARRIAGE_RETURN_BYTE_NUMBER = 13 // \r
const LINE_FEED_BYTE_NUMBER = 10 // \n

func splitFromStartIndexToCRLF(startIndex int, message []byte) []byte {
    var deserialized []byte 
    for startIndex < len(message){
        if message[startIndex] == CARRIAGE_RETURN_BYTE_NUMBER || message[startIndex] == LINE_FEED_BYTE_NUMBER {
            startIndex += 1
            continue
        }
        deserialized = append(deserialized, message[startIndex])
        startIndex += 1
    }
    return deserialized
}

func deserializeSimpleString(startIndex int, message []byte) []string {
    return []string{string(splitFromStartIndexToCRLF(startIndex, message))}
}

func deserializeError(startIndex int, message []byte) []string {
    return []string{string(splitFromStartIndexToCRLF(startIndex, message))}
}

func deserializeBulkString(startIndex int, message []byte) []string {
    var deserialized []byte 
    for startIndex < len(message){
        _, err := strconv.Atoi(string(message[startIndex]))
        if err == nil {
            startIndex += 1
            continue
        }
        if message[startIndex] == CARRIAGE_RETURN_BYTE_NUMBER || message[startIndex] == LINE_FEED_BYTE_NUMBER {
            startIndex += 1
            continue
        }
        deserialized = append(deserialized, message[startIndex])
        startIndex += 1
    }
    return []string{string(deserialized)}
}

func deserializeArray(startIndex int, message []byte) []string {
    var deserializedArray []string

    for startIndex < len(message){
        if message[startIndex] == CARRIAGE_RETURN_BYTE_NUMBER {
            startIndex += 2
        }
        startIndex += 1
    }

    return deserializedArray
}

func deserialize(message []byte) []string {
    if string(message[0]) == SIMPLE_STRING_CH {
        return deserializeSimpleString(1, message)
    }
    if string(message[0]) == ERROR_STRING_CH {
        return deserializeError(1, message)
    }
    if string(message[0]) == BULK_STRING_CH {
        return deserializeBulkString(1, message)
    }
    if string(message[0]) == "*" {
        return deserializeArray(1, message)
    }
    return []string{""}
}

func main() {
    fmt.Print("lets do it!")    
}