package main

import "fmt"

const SIMPLE_STRING_CH = "+"
const BACHSLASH_N_BYTE_NUMBER = 13
const BACHSLASH_R_BYTE_NUMBER = 10


func deserializeSimpleString(message []byte) string {
    var deserialized []byte 
    var index = 1;
    for index < len(message){
        if message[index] == BACHSLASH_N_BYTE_NUMBER || message[index] == BACHSLASH_R_BYTE_NUMBER {
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
    return ""
}

func main() {
    fmt.Print("lets do it!")    
}