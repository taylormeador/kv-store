package main

import (
	"log"
	"net"
)

const PORT_NUMBER string = ":8080"

func handleConnection(conn net.Conn) {

	data := "hello world\r\n"
	_, err := conn.Write([]byte(data))
	if err != nil {
		log.Println(err)
	}
}

func main() {
	log.Printf("Starting server on localhost%s", PORT_NUMBER)

	ln, err := net.Listen("tcp", PORT_NUMBER)
	if err != nil {
		log.Println(err)
		return
	}
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
