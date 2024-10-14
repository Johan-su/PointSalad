package main

import (
	"log"
	"net"
)

type TCPServer struct {
	conn      []net.Conn
	out       map[int]chan []byte
	in        map[int]chan []byte
	quitRead  map[int]chan bool
	quitWrite map[int]chan bool

	serverReceiveSize int
	serverSendSize    int
	listener          net.Listener
}

func (server *TCPServer) Init(port string, playerNum int) error {
	log.Printf("listening on port %v\n", port)
	log.Printf("Waiting for %d player(s)\n", playerNum-len(server.conn))
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	server.listener = ln
	id := 0

	server.out = make(map[int]chan []byte)
	server.in = make(map[int]chan []byte)

	for len(server.conn) < playerNum {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection %s", err)
			continue
		}
		addr := conn.RemoteAddr()
		log.Printf("%s connected\n", addr.String())
		log.Printf("Waiting for %d player(s)\n", playerNum-len(server.conn))
		server.conn = append(server.conn, conn)
		server.out[id] = make(chan []byte)
		server.in[id] = make(chan []byte)
		go handleRead(server, id)
		go handleWrite(server, id)
		id += 1
	}
	return nil
}

func (s *TCPServer) Close() {
	for k := range s.in {
		s.conn[k].Close()
	}
	s.listener.Close()
	for k := range s.in {
		s.quitRead[k] <- true
		s.quitWrite[k] <- true
		close(s.in[k])
		// closing out is probably not necessary
		// close(s.out[k])
	}
}

func (s *TCPServer) GetReadChannels() map[int]chan []byte {
	return s.in
}

func (s *TCPServer) GetWriteChannels() map[int]chan []byte {
	return s.out
}

func handleRead(s *TCPServer, connId int) {
	running := true
	for running {
		buf := make([]byte, s.serverReceiveSize, s.serverReceiveSize)
		read, err := s.conn[connId].Read(buf)
		if err != nil {
			running = false
			log.Printf("ERROR: connId = %d, read = %v, err = %s\n", connId, read, err)
		}
		select {
		case <-s.quitRead[connId]:
			{
				running = false
			}

		case s.in[connId] <- buf:
		}
	}
}

func handleWrite(s *TCPServer, connId int) {
	running := true
	var buf []byte
	for running {
		select {
		case <-s.quitWrite[connId]:
			{
				running = false
			}
		case buf = <-s.out[connId]:
		}
		written, err := s.conn[connId].Write(buf)
		if err != nil {
			log.Printf("ERROR: connId = %d, written = %v, err = %s\n", connId, written, err)
			break
		}
	}
}
