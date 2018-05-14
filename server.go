package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"net"
	"strings"
)

const MSG_SIZE = 512

type SocketServer struct {
	port    string
	uuid    string
	servers map[string]net.IP
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (s *SocketServer) Setup(port string) {
	s.port = port
	s.uuid = uuid.Must(uuid.NewV4()).String()
	s.servers = make(map[string]net.IP)
}

func (s *SocketServer) Serve() {
	LocalAddr, err := net.ResolveUDPAddr("udp", s.port)
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", LocalAddr)
	CheckError(err)

	go func() {
		buf := make([]byte, MSG_SIZE)
		defer ServerConn.Close()
		for {
			n, addr, err := ServerConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			id := string(buf[0:n])
			// do not answer to myself
			if strings.Compare(s.uuid, id) == 0 {
				continue
			}
			_, exists := s.servers[id]
			if !exists {
				fmt.Println("Received ", id, " from ", addr)
				s.servers[id] = addr.IP
				_, err = ServerConn.WriteToUDP([]byte(s.uuid), addr)
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}

		}
	}()

	// now Discover
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255"+s.port)
	CheckError(err)
	buf2 := []byte(s.uuid)
	_, err = ServerConn.WriteToUDP(buf2, ServerAddr)
	if err != nil {
		fmt.Println(s.uuid, err)
	}

}
