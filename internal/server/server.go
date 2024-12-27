package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
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


var cache = make(map[string][]byte)


func saveSnapshot(cache map[string][]byte) error {
    fi, err := os.Create("snapshop.gedis")
    if err != nil {
        return err
    }
    for key, value := range cache {
        fi.Write([]byte(key))
        fi.Write([]byte("\n"))
        fi.Write(value)
        fi.Write([]byte("\n"))
    }
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


type Deserializer interface {
    deserialize(message []byte) [][]byte
}

type ArrayDeserializer struct {}
func (d ArrayDeserializer) deserialize(message []byte) [][]byte {
    startIndex := 1
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

type SimpleStringDeserializer struct{}
func (d SimpleStringDeserializer) deserialize(message []byte) [][]byte {
    return [][]byte{splitFromStartIndexToCRLF(1, message)}
}

type ErrorDeserializer struct{}
func (d ErrorDeserializer) deserialize(message []byte) [][]byte {
    return  [][]byte{splitFromStartIndexToCRLF(1, message)}
}

type BulkStringDeserializer struct{}
func (d BulkStringDeserializer) deserialize(message []byte) [][]byte {
    return deserializeBulkString(1, message)
}

func getDeserializer(message []byte) (Deserializer, error) {
    if message[0] == ARRAY_STRING_BYTE_NUMBER {
        return ArrayDeserializer{}, nil
    }
    if message[0] == SIMPLE_STRING_BYTE_NUMBER {
        return SimpleStringDeserializer{}, nil
    }
    if message[0] == ERROR_STRING_BYTE_NUMBER {
        return ErrorDeserializer{}, nil
    }
    if message[0] == BULK_STRING_BYTE_NUMBER {
        return BulkStringDeserializer{}, nil
    }
    return nil, fmt.Errorf("no deserializer for")
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
    cache[string(message[1])] = message[2]
    return serializeSimpleStringFromString("OK")
}

type CommandGet struct {}
func (command CommandGet) execute(message[][]byte) []byte {
    value, ok := cache[string(message[1])]
    if ok {
        return serializeSimpleStringFromByteArray(value)
    }
    return serializeError("doesn't exist")
}

type CommandExists struct {}
func (command CommandExists) execute(message[][]byte) []byte {
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

type CommandDel struct {}
func (command CommandDel) execute(message[][]byte) []byte {
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

type CommandIncr struct {}
func (command CommandIncr) execute(message[][]byte) []byte {
    var key = string(message[1])
    value, ok := cache[key]

    if ok {
        value, err := strconv.ParseInt(string(value), 10, 64)
        if err != nil {
            return serializeError("ERR value is not an integer or out of range")
        }
        newValue := value + 1
        cache[key] = []byte(strconv.FormatInt(newValue, 10))
        return serializerInteger64(newValue)
    }

    var newValue int64 = 1 

    cache[key] = []byte(strconv.FormatInt(newValue, 10))
    return serializerInteger64(newValue)
}

func getCommand(message[][]byte) (Command, error) {
    if bytes.Equal(message[0], SET_BYTE_ARRAY) {
        return CommandSet{}, nil
    } else if bytes.Equal(message[0], GET_BYTE_ARRAY) {
        return CommandGet{}, nil
    } else if bytes.Equal(message[0], PING_BYTE_ARRAY) {
        return CommandPing{}, nil
    } else if bytes.Equal(message[0], ECHO_BYTE_ARRAY) {
        return CommandEcho{}, nil
    } else if bytes.Equal(message[0], EXISTS_BYTE_ARRAY) {
        return CommandExists{}, nil
    } else if bytes.Equal(message[0], DEL_BYTE_ARRAY) {
        return CommandDel{}, nil
    } else if bytes.Equal(message[0], INCR_BYTE_ARRAY) {
        return CommandIncr{}, nil
    } else {
        return nil, errors.New("no command found for message")
    }
}

func periodicallySaveSnapshot() {
    for {
        time.Sleep(5 * time.Second)
        saveSnapshot(cache)
    }
}

func Start() {
    restoreErr, newCache := restoreSnapshot()
    if restoreErr == nil {
        cache = newCache
    }

    listerner, listenErr := net.Listen("tcp", "localhost:6379")
    if listenErr != nil {
        fmt.Println(listenErr)
        return
    }
    for {
        conn, acceptErr := listerner.Accept()
        if acceptErr != nil {
            fmt.Println(acceptErr)
            continue
        }
        go handleConnection(conn)
        go periodicallySaveSnapshot()
    }    
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 8192)
    for {
        // #TODO this read op is not safe as we don´t know where the message ends. this needs to be refactored.
        bytesNumber, connReadErr := conn.Read(buffer)
        if connReadErr != nil {
            fmt.Println("Error:", connReadErr)
            return
        }

        deserializer, getDesError := getDeserializer(buffer[:bytesNumber])
        if getDesError != nil {
            continue
        }
        message := deserializer.deserialize(buffer[:bytesNumber])
        command, getCommandErr := getCommand(message)

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
    return DeserializationResult{
        EndIndex: endIndex, 
        Arguments: [][]byte{data[startIndex+1:endIndex-2]},
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
        return DeserializationResult{
            EndIndex: endIndex+2,
            Arguments: [][]byte{data[startIndex+1:endIndex]},
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
        for _, arg := range result.Arguments {
            arguments = append(arguments, arg)
        }
    }

    return DeserializationResult{EndIndex: endIndex, Arguments: arguments}, nil
}

type DeserializationBuffer struct {
    data []byte
}

func (sb *DeserializationBuffer) Dissipate() ([]byte, error) {
	if len(sb.data) == 0 {
		return []byte{}, errors.New("serialization error: no data in buffer")
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
                return []byte{}, err
            }
            newData := sb.copyBytesFromBuffer(i, result.EndIndex)
            sb.rearrengeBuffer(result.EndIndex)
            return newData, nil           
        }
	}

	return []byte{}, errors.New("serialization error: unknown first byte data type")
}

func (c *DeserializationBuffer) copyBytesFromBuffer(startIndex int, endIndex int) []byte {
	newData := []byte{}
	for i := startIndex; i <= endIndex; i++ {
		newData = append(newData, c.data[i])
	}
    return newData
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

func NewDeserializationBuffer() DeserializationBuffer {
    return DeserializationBuffer{make([]byte, 100)}
}
