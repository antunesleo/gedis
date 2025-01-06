package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
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
var INCR_BYTE_ARRAY = []byte("INCR")


var cache sync.Map


func saveSnapshot(cache *sync.Map) error {
    fi, err := os.Create("snapshot.gedis")
    if err != nil {
        return err
    }
    defer fi.Close()

    cache.Range(func(key, value interface{}) bool {
        fi.Write([]byte(key.(string)))
        fi.Write([]byte("\n"))
        fi.Write(value.([]byte))
        fi.Write([]byte("\n"))
        return true
    })

    return nil
}

func restoreSnapshot() (error, map[string][]byte) {
    innerCache := map[string][]byte{}
    fil, err := os.Open("snapshop.gedis")
    if err != nil {
        return err, nil
    }
    buffer := make([]byte, 10000)
    _, err = fil.Read(buffer)
    if err != nil {
        return err, nil
    }

    var index = 0
    for index < len(buffer) {
        var key []byte
        var value []byte

        var keyFinished = false
        for  index < len(buffer) && !keyFinished  {
            if buffer[index] == LINE_FEED_BYTE_NUMBER {
                keyFinished = true
            } else {
                key = append(key, buffer[index])
            }
            index += 1
        }

        var valueFinished = false
        for  index < len(buffer) && !valueFinished  {
            if buffer[index] == LINE_FEED_BYTE_NUMBER {
                valueFinished = true
            } else {
                value = append(value, buffer[index])
            }
            index += 1
        }

        if len(key) != 0 && len(value) != 0 {
            innerCache[string(key)] = value
        }
    }

    return nil, innerCache
}

func isNumeric(c byte) bool {
    return (c >= '0' && c <= '9')
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

func serializerInteger64(intToSerialize int64) []byte {
    return []byte(fmt.Sprintf(":%d\r\n", intToSerialize))
}

type Command interface {
    execute(message[][]byte) []byte
}

type CommandPing struct {}
func (command CommandPing) execute(message[][]byte) []byte {
    return serializeSimpleStringFromString("PONG")
}

type CommandEcho struct {}
func (command CommandEcho) execute(message[][]byte) []byte {
    return serializeSimpleStringFromByteArray(message[1])
}

type CommandSet struct {}
func (command CommandSet) execute(message[][]byte) []byte {
    cache.Store(string(message[1]), message[2])
    return serializeSimpleStringFromString("OK")
}

type CommandGet struct {}
func (command CommandGet) execute(message[][]byte) []byte {
    value, ok := cache.Load(string(message[1]))
    if ok {
        return serializeSimpleStringFromByteArray(value.([]byte))
    }
    return serializeError("doesn't exist")
}

type CommandExists struct {}
func (command CommandExists) execute(message[][]byte) []byte {
    var existsCount = 0
    var itemIndex = 1
    for itemIndex < len(message) {
        _, ok := cache.Load(string(message[1]))
        if ok {
            existsCount += 1
        }
        itemIndex += 1
    }

    return serializerInteger(existsCount)
}


func getCommand(messageFirstArgument *[]byte) (Command, error) {
    if bytes.Equal(*messageFirstArgument, SET_BYTE_ARRAY) {
        return &CommandSet{}, nil
    } else if bytes.Equal(*messageFirstArgument, GET_BYTE_ARRAY) {
        return &CommandGet{}, nil
    } else if bytes.Equal(*messageFirstArgument, PING_BYTE_ARRAY) {
        return &CommandPing{}, nil
    } else if bytes.Equal(*messageFirstArgument, ECHO_BYTE_ARRAY) {
        return &CommandEcho{}, nil
    } else if bytes.Equal(*messageFirstArgument, EXISTS_BYTE_ARRAY) {
        return &CommandExists{}, nil
    } else {
        return nil, errors.New("no command found for message")
    }
}

func periodicallySaveSnapshot() {
    for {
        time.Sleep(5 * time.Second)
        saveSnapshot(&cache)
    }
}

func Start() {
    // restoreErr, newCache := restoreSnapshot()
    // if restoreErr == nil {
    //     cache = newCache
    // }

    listerner, listenErr := net.Listen("tcp", "localhost:6379")
    if listenErr != nil {
        fmt.Println(listenErr)
        return
    }
    // go periodicallySaveSnapshot()
    for {
        conn, acceptErr := listerner.Accept()
        if acceptErr != nil {
            fmt.Println(acceptErr)
            continue
        }
        go handleConnection(conn)
    }    
}

func cmdSet(message[][]byte) []byte {
    cache.Store(string(message[1]), message[2])
    return serializeSimpleStringFromString("OK")
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    deserializationBuffer := NewDeserializationBuffer()

    for {
        connBuffer := make([]byte, 8192)
        bytesRead, connReadErr := conn.Read(connBuffer)
        if connReadErr != nil {
            fmt.Println("Error:", connReadErr)
            return
        }

        err := deserializationBuffer.Absorb(connBuffer[:bytesRead])
        if err != nil {
            fmt.Println("Error:", err)
        }
        theResult, dissipateErr := deserializationBuffer.Dissipate()
        if dissipateErr != nil {
            continue
        }

        message := theResult.Arguments
        command, getCommandErr := getCommand(&message[0])

        var result []byte
        if getCommandErr != nil {
            result = serializeError("not implemented")
        } else {
            result = command.execute(message)
        }

        totalWritten := 0
        for totalWritten < len(result) {
            bytesWrittenNumbers, connWriteErr := conn.Write(result[totalWritten:])
            if connWriteErr != nil {
                fmt.Println("Error:", connWriteErr)
                return
            }
            totalWritten += bytesWrittenNumbers
        }
    }
}

func ByteSliceToInteger(byteSlice []byte) (int, error) {
    str := string(byteSlice)
    num, err := strconv.Atoi(str)
    return num, err
}

func FindIndexAfterCrlf(data []byte, startIndex int) (int, error) {
    crlfFound := false
    currIndex := startIndex
    for !crlfFound && currIndex < len(data)-2 {
        if data[currIndex+1] == CARRIAGE_RETURN_BYTE_NUMBER && data[currIndex+2] == LINE_FEED_BYTE_NUMBER {
            crlfFound = true
            break
        }
        currIndex += 1
    }
    if crlfFound {
        return currIndex+2, nil
    }
    return -1, errors.New("serialization errror: no crlf found")
}

func GetEndLenghtIndex(data []byte, startLengthIndex int) (int, error) {
	endLengthIndex := startLengthIndex
	hasCrfl := false
	for endLengthIndex < len(data)-2 {
		if data[endLengthIndex+1] == CARRIAGE_RETURN_BYTE_NUMBER && data[endLengthIndex+2] == LINE_FEED_BYTE_NUMBER {
			hasCrfl = true
			break
		}
		endLengthIndex += 1
	}
	if !hasCrfl {
		return 0, errors.New("serialization errror: no crlf found")
	}
	return endLengthIndex, nil
}

func ValidateNumberOfElements(data []byte, startLengthIndex int) (int, int, error) {
    if startLengthIndex >= len(data) {
        return -1, -1, errors.New("serialization error: no length found")
    }

    endLengthIndex, err := GetEndLenghtIndex(data, startLengthIndex)
    if err != nil {
    	return -1, -1, err
    }

    length, err := ByteSliceToInteger(data[startLengthIndex:endLengthIndex+1])
    if err != nil {
        return -1, -1, errors.New("serialization error: no length found")
    }

    return length, endLengthIndex, nil
}

func  ValidateCarriageReturnAndLineFeed(data []byte, carriageReturnIndex int) (int, error) {
    if carriageReturnIndex >= len(data) {
        return -1, errors.New("serialization error: no carriage return found")
    }

    lineFeedIndex := carriageReturnIndex + 1
    if lineFeedIndex >= len(data) {
        return -1, errors.New("serialization error: no line feed found")
    }
    return lineFeedIndex, nil
}

func CopyBytesFromBuffer(data []byte, startIndex int, endIndex int) []byte {
	newData := []byte{}
	for i := startIndex; i <= endIndex; i++ {
		newData = append(newData, data[i])
	}
    return newData
}

type DeserializationResult struct {
    EndIndex int
    Arguments [][]byte
}

func Deserialize(data []byte, startIndex int) (DeserializationResult, error) {
    item := data[startIndex]
    if item == SIMPLE_STRING_BYTE_NUMBER {
        serializer := SimpleStringDeserializer2{}
        return serializer.Deserialize(data, startIndex)
    } else if item == ERROR_STRING_BYTE_NUMBER {
        serializer := ErrorDeserializer2{}
        return serializer.Deserialize(data, startIndex)
    } else if item == BULK_STRING_BYTE_NUMBER {
        serializer := BulkStringDeserializer2{}
        return serializer.Deserialize(data, startIndex)
    } else if item == ARRAY_STRING_BYTE_NUMBER {
        serializer := ArrayDeserializer2{}
        return serializer.Deserialize(data, startIndex)
    }

    return DeserializationResult{}, errors.New("serialization error: unknown first byte data type")
}

type Deserializer2 interface {
    Deserialize(data []byte, startIndex int) (int, error)
}

type SimpleStringDeserializer2 struct {}
func (s SimpleStringDeserializer2) Deserialize(data []byte, startIndex int) (DeserializationResult, error) {
    endIndex, err := FindIndexAfterCrlf(data, startIndex+1)
    if err != nil {
        return DeserializationResult{}, err
    }
    argument := CopyBytesFromBuffer(data, startIndex+1, endIndex-2)
    return DeserializationResult{
        EndIndex: endIndex, 
        Arguments: [][]byte{argument},
    }, nil
}

type ErrorDeserializer2 struct {}
func (s ErrorDeserializer2) Deserialize(data []byte, startIndex int) (DeserializationResult, error) {
    endIndex, err := FindIndexAfterCrlf(data, startIndex+1)
    if err != nil {
        return DeserializationResult{}, err
    }
    return DeserializationResult{
        EndIndex: endIndex, 
        Arguments: [][]byte{data[startIndex+1:endIndex-2]},
    }, nil
}

type BulkStringDeserializer2 struct {}
func (s BulkStringDeserializer2) Deserialize(data []byte, startIndex int) (DeserializationResult, error) {
    startLengthIndex := startIndex + 1
    length, endLengthIndex, err := ValidateNumberOfElements(data, startLengthIndex)

    if err != nil {
        return DeserializationResult{}, err
    }

    lineFeedIndex, err := ValidateCarriageReturnAndLineFeed(data, endLengthIndex+1)
    if err != nil {
        return DeserializationResult{}, err
    }

    endIndex := lineFeedIndex + length

    if endIndex >= len(data)-2 {
        return DeserializationResult{}, errors.New("serialization errror: no crlf found") 
    }


    if data[endIndex+1] == CARRIAGE_RETURN_BYTE_NUMBER && data[endIndex+2] == LINE_FEED_BYTE_NUMBER {
        argument := CopyBytesFromBuffer(data, lineFeedIndex+1, endIndex)
        return DeserializationResult{
            EndIndex: endIndex+2,
            Arguments: [][]byte{argument},
        }, nil
    }

    return DeserializationResult{}, errors.New("serialization errror: no crlf found")      
}

type ArrayDeserializer2 struct {}
func (s ArrayDeserializer2) Deserialize(data []byte, startIndex int) (DeserializationResult, error) {

    startLengthIndex := startIndex + 1
    length, endLengthIndex, err := ValidateNumberOfElements(data, startLengthIndex)

    if err != nil {
        return DeserializationResult{}, err
    }

    lineFeedIndex, err := ValidateCarriageReturnAndLineFeed(data, endLengthIndex+1)
    if err != nil {
        return DeserializationResult{}, err
    }

    arguments := [][]byte{}
    endIndex := lineFeedIndex
    for i := 0; i < length; i++ {
        result, err := Deserialize(data, endIndex+1)
        if err != nil {
            return DeserializationResult{}, err
        }
        endIndex = result.EndIndex
        arguments = append(arguments, result.Arguments...)
    }

    return DeserializationResult{EndIndex: endIndex, Arguments: arguments}, nil
}

type DeserializationBuffer struct {
    data []byte // Possible could be implemented as a linked list
}

func (c *DeserializationBuffer) rearrengeBuffer(endIndex int) {
	for i := 0; i <= endIndex; i++ {
		c.data[i] = 0
	}
	emptyIndex := 0
	for leftBehindIndex := endIndex + 1; leftBehindIndex < len(c.data); leftBehindIndex++ {
		if c.data[leftBehindIndex] != 0 {
			c.data[emptyIndex] = c.data[leftBehindIndex]
			c.data[leftBehindIndex] = 0
			emptyIndex += 1
		} else {
			break
		}
	}
}

func (c *DeserializationBuffer) Absorb(bytes []byte) error {
    emptyIndex := -1
    for i, _byte := range c.data {
        if _byte == 0 {
            emptyIndex = i
            break
        }
    }
    if emptyIndex < 0 {
        return errors.New("serialization error: buffer is full")
    }
    
    availableSlots := len(c.data) - emptyIndex
    if len(bytes) > availableSlots {
        return errors.New("serialization error: buffer is full")
    }

    for _, _byte := range bytes {
        c.data[emptyIndex] = _byte
        emptyIndex += 1
    }
    return nil
}

func (sb *DeserializationBuffer) Dissipate() (DeserializationResult, error) {
	if len(sb.data) == 0 {
		return DeserializationResult{}, errors.New("serialization error: no data in buffer")
	}

	for i, _byte := range sb.data {
        knownFirstBytes := []byte{
            SIMPLE_STRING_BYTE_NUMBER,
            ERROR_STRING_BYTE_NUMBER,
            BULK_STRING_BYTE_NUMBER,
            ARRAY_STRING_BYTE_NUMBER,
        }
        if slices.Contains(knownFirstBytes, _byte) {
            result, err := Deserialize(sb.data, i)
            if err != nil {
                return DeserializationResult{}, err
            }
            sb.rearrengeBuffer(result.EndIndex)
            return result, nil           
        }
	}

	return DeserializationResult{}, errors.New("serialization error: unknown first byte data type")
}

func NewDeserializationBuffer() DeserializationBuffer {
    return DeserializationBuffer{make([]byte, 100)}
}
