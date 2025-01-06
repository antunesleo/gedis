package tcp_senders

import (
	"fmt"
	"net"
)

func SendManySetCommands() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	for i := 0; i < 20; i++ {
		message := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$29\r\nmylonglonglongsodamnlongvalue\r\n"

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
	}
}
