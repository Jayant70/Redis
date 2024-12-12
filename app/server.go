package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	//Listen for incoming connections
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	//Ensure we tear down the server when the program exists
	defer listener.Close()

	fmt.Println("Server is listening on port 6379")
	fmt.Println("Test commit")

	for {
		//Block untill we recive an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		//Handle client connection
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	//Ensure we close the connection after we're done
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed or error:", err)
			return
		}
		//Parsed RESP input
		input := string(buf[:n])
		command, message, err := parseRESP(input)
		if err != nil {
			fmt.Println("Invalid RESP format\r\n")
			conn.Write([]byte("-ERR invalid RESP format\r\n"))
			continue
		}
		//process the commad
		switch strings.ToUpper(command) {
		case "ECHO":
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)))
		case "PING":
			conn.Write([]byte("+PONG/r/n"))
		default:
			conn.Write([]byte("-ERR Unknown Command\r\n"))
		}
	}
}

func parseRESP(input string) (string, string, error) {
	reader := bufio.NewReader(strings.NewReader(input))

	//read array header
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(line, "*") {
		return "", "", fmt.Errorf("invalid header format")
	}
	// Read the number of elements
	numElements, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(line), "*"))
	if err != nil || numElements < 2 {
		return "", "", fmt.Errorf("invalid array element count")
	}
	//Read the command first bulk string
	command, err := readBulkString(reader)
	if err != nil {
		return "", "", err
	}
	//Read the message Seconf bulk string
	message, err := readBulkString(reader)
	if err != nil {
		return "", "", err
	}
	return command, message, nil
}

func readBulkString(reader *bufio.Reader) (string, error) {
	//Read bulk string length line
	lengthLine, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(lengthLine, "$") {
		return "", fmt.Errorf("invalid bulk string format")
	}
	//parse the length
	length, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(lengthLine), "$"))
	if err != nil {
		return "", err
	}
	// Handle cases where the length is negative (RESP null bulk string)
	if length < 0 {
		return "", nil // RESP null bulk strings are valid and return nil
	}
	//Read the bulk string
	data := make([]byte, length)
	_, err = reader.Read(data)
	if err != nil {
		return "", err
	}
	//consume the trailing \r\n
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return string(data), nil
}
