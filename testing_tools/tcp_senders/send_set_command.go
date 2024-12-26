package tcp_senders

import (
	"bufio"
	"fmt"
	"net"
)

func SendSetCommand() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	message := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$29\r\nmylonglonglongsodamnlongvalue\r\n"
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
