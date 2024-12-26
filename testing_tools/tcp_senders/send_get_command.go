package tcp_senders

import (
	"bufio"
	"fmt"
	"net"
)

func SendGetCommand() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send the string to the server
	message := "*2\r\n$3\r\nget\r\n$3\r\nkey\r\n"
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
