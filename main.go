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
            break
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
    moveForward := true
    
    var numberBuffer []byte
    index := startIndex
    for moveForward {
        _, err := strconv.Atoi(string(message[index]))
        if err == nil {
            numberBuffer = append(numberBuffer, message[index])
            index += 1
        } else {
            moveForward = false
        }
    }
    
    number, err := strconv.Atoi(string(numberBuffer))
    if err == nil && number == 0 {
        // empty
        return []string{}
    }


    if message[index] == CARRIAGE_RETURN_BYTE_NUMBER && message[index+1] == LINE_FEED_BYTE_NUMBER {
        index += 2
    } else {
        // invalid
        return []string{}
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
    return []string{string(stringBuffer)}
}


func deserializeArray(startIndex int, message []byte) []string {
    var deserializedArray []string

    for startIndex < len(message) {
        if string(message[startIndex]) == "$" {
            startIndex += 1
            moveForward := true
            for moveForward {
                _, err := strconv.Atoi(string(message[startIndex]))
                if err == nil {
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