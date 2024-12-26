package tcp_senders

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func SendSetCommandWithGiantValue() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	giant_value := strings.Repeat("A", 200000)
	message := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$%d\r\n%s\r\n", len(giant_value), giant_value)
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	// Read the response from the server
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response from server:", response)
}
