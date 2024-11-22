package main

import (
	"fmt"
	"net"
	"os"
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

	for {
		//Block untill we recive an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		//Handle client connection
		handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	//Ensure we close the connection after we're done
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		fmt.Println("Received data", buf[:n])
		// Write the same data back
		conn.Write([]byte("+PONG\r\n"))
	}
}
