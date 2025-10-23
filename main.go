package main

import (
	"log"
	"net"
)

const PORT_NUMBER string = ":8080"

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Enter read/respond loop
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}
		message := string(buf[:n])
		log.Println(message)

		// Send OK response
		_, err = conn.Write([]byte("OK\r\n"))
		if err != nil {
			log.Println(err)
		}
	}

}

func main() {
	log.Printf("Starting server on localhost%s", PORT_NUMBER)

	ln, err := net.Listen("tcp", PORT_NUMBER)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		log.Printf("Connected to %s", conn.RemoteAddr().String())
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(conn)
	}
}
