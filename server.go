package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/satori/go.uuid"
)

const MSG_SIZE = 512

type Message struct {
	Action string
	Source string
}

type Packet struct {
	ID       string
	Response string
	Content  []byte
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
	s.uuid = uuid.NewV4().String()
	s.servers = make(map[string]net.IP)
	s.ready = make(chan struct{})
}

func (s *SocketServer) Init(readPort string, writePort string) (<-chan Packet, chan<- Packet) {
	receive := make(chan Packet, 10)
	send := make(chan Packet, 10)
	go s.listen(receive, readPort)
	go s.broadcast(send, writePort)
	return receive, send
}

func (s *SocketServer) listen(receive chan Packet, port string) {
	localAddress, _ := net.ResolveUDPAddr("udp", port)
	connection, err := net.ListenUDP("udp", localAddress)
	CheckError(err)
	defer connection.Close()
	var message Packet
	for {
		inputBytes := make([]byte, 4096)
		length, _, _ := connection.ReadFromUDP(inputBytes)
		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := gob.NewDecoder(buffer)
		decoder.Decode(&message)
		//Filters out all messages not relevant for the system
		if message.ID == "toto" {
			receive <- message
		}
	}
}

func (s *SocketServer) broadcast(send chan Packet, port string) {
	localAddress, _ := net.ResolveUDPAddr("udp", port)
	destinationAddress, _ := net.ResolveUDPAddr("udp", "255.255.255.255"+port)
	connection, err := net.DialUDP("udp", localAddress, destinationAddress)
	CheckError(err)
	defer connection.Close()
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	for {
		message := <-send
		encoder.Encode(message)
		connection.Write(buffer.Bytes())
		buffer.Reset()
	}
}

func (s *SocketServer) Serve() {
	inBuffer := new(bytes.Buffer)
	LocalAddr, err := net.ResolveUDPAddr("udp", s.port)
	CheckError(err)

	conn, err := net.ListenUDP("udp", LocalAddr)
	CheckError(err)
	defer conn.Close()

	remoteAddr := conn.LocalAddr()
	fmt.Println("Remote Addr: ", remoteAddr)

	// Initialize the encoder and decoder. Normally enc and dec would be
	// bound to network connections and the encoder and decoder would
	// run in different processes
	enc := gob.NewEncoder(inBuffer) // Will write to network.

	msg := Message{Action: "ASK", Source: s.uuid}
	enc.Encode(&msg)
	fmt.Println("Send message: ", msg, " buffer len ", inBuffer.Len())
	// now Discover
	broadCastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255"+s.port)
	CheckError(err)
	_, err = conn.WriteToUDP(inBuffer.Bytes(), broadCastAddr)
	if err != nil {
		log.Println(s.uuid, err)
	}

	for {
		var input Message
		inputBytes := make([]byte, MSG_SIZE)
		lenght, addr, err := conn.ReadFromUDP(inputBytes)
		if err != nil {
			log.Println("Error: ", err)
		}
		buffer := bytes.NewBuffer(inputBytes[:lenght])
		decoder := gob.NewDecoder(buffer)
		decoder.Decode(&input)
		fmt.Println("Receive message: ", input, " size ", lenght, " buffer len ", buffer.Len())
		// do not answer to myself
		if strings.Compare(s.uuid, msg.Source) == 0 {
			continue
		}
		_, exists := s.servers[msg.Source]
		if !exists {
			fmt.Println("Received ", msg, " from ", addr)
			s.servers[msg.Source] = addr.IP
			_, err = conn.WriteToUDP([]byte(s.uuid), addr)
			if err != nil {
				log.Println("Error: ", err)
			}
		}

	}

}
