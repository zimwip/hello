package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"net"
	"strings"
)

const MSG_SIZE = 512

type message struct {
	action string
	source string
}

type SocketServer struct {
	port    string
	uuid    string
	servers map[string]net.IP
	ready   chan struct{}
}

func CheckError(err error) {
	if err != nil {
		log.Println("Error: ", err)
	}
}

func (s *SocketServer) Setup(port string) {
	s.port = port
	s.uuid = uuid.Must(uuid.NewV4()).String()
	s.servers = make(map[string]net.IP)
	s.ready = make(chan struct{})
}

func (s *SocketServer) Serve() {
	buffer := new(bytes.Buffer)
	LocalAddr, err := net.ResolveUDPAddr("udp", s.port)
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", LocalAddr)
	CheckError(err)
	defer ServerConn.Close()

	// Initialize the encoder and decoder. Normally enc and dec would be
	// bound to network connections and the encoder and decoder would
	// run in different processes
	enc := gob.NewEncoder(buffer) // Will write to network.
	dec := gob.NewDecoder(buffer)     // Will read from network.

	msg := message{action: "ASK", source: s.uuid}
	enc.Encode(&msg)
	fmt.Println("Send message: ", msg, " buffer len ", buffer.Len())
	// now Discover
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255"+s.port)
	CheckError(err)
	_, err = ServerConn.WriteToUDP(buffer.Bytes(), ServerAddr)
	if err != nil {
		log.Println(s.uuid, err)
	}
	localBuf := make([]byte, MSG_SIZE)
	for {
		var input message
		n, addr, err := ServerConn.ReadFromUDP(localBuf)
		if err != nil {
			log.Println("Error: ", err)
		}
		buffer.Reset()
		buffer.Write(localBuf[:n])
		fmt.Println("Buffer len: ", buffer.Len())
		dec.Decode(&input)
		fmt.Println("Receive message: ", input, " size ", n, " buffer len ", buffer.Len())
		// do not answer to myself
		if strings.Compare(s.uuid, msg.source) == 0 {
			continue
		}
		_, exists := s.servers[msg.source]
		if !exists {
			fmt.Println("Received ", msg, " from ", addr)
			s.servers[msg.source] = addr.IP
			_, err = ServerConn.WriteToUDP([]byte(s.uuid), addr)
			if err != nil {
				log.Println("Error: ", err)
			}
		}

	}

}
