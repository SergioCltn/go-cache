package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter command: ")
		scanner.Scan()
		commandStr := scanner.Text()

		parts := strings.Fields(commandStr)
		if len(parts) == 0 {
			fmt.Println("Invalid command format")
			continue
		}

		sendCommand(conn, commandStr)

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading response:", err)
			break
		}

		fmt.Println("Server response:", response)
	}
}

func sendCommand(conn net.Conn, command string) {
	_, err := conn.Write(append([]byte(command), '\n'))
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}
}
