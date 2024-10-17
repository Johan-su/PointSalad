package main

import (
	"log"
	"net"
	"fmt"
)

type TCPServer struct {
	conn      []net.Conn
	out       map[int]chan []byte
	in        map[int]chan []byte
	quitRead  map[int]chan bool
	quitWrite map[int]chan bool

	serverMaxReceiveSize int
	listener          net.Listener
}

func (server *TCPServer) Init(port string, playerNum int, serverMaxReceiveSize int) error {
	server.serverMaxReceiveSize = serverMaxReceiveSize
	log.Printf("listening on port %v\n", port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	server.listener = ln
	id := 0

	server.out = make(map[int]chan []byte)
	server.in = make(map[int]chan []byte)
	server.quitRead = make(map[int]chan bool)
	server.quitWrite = make(map[int]chan bool)

	for len(server.conn) < playerNum {
		log.Printf("Waiting for %d player(s)\n", playerNum-len(server.conn))
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection %s", err)
			continue
		}
		addr := conn.RemoteAddr()
		log.Printf("%s connected\n", addr.String())
		server.conn = append(server.conn, conn)
		server.out[id] = make(chan []byte)
		server.in[id] = make(chan []byte)
		server.quitRead[id] = make(chan bool)
		server.quitWrite[id] = make(chan bool)


		buf := make([]byte, 4, 4)
		_, err = conn.Read(buf)
		if err != nil {
			return err
		}
		if string(buf) != "ping" {
			return fmt.Errorf("expected ping pong test\n")
		}
		buf = make([]byte, 4, 4)
		buf = []byte("pong")
		_, err = conn.Write(buf)
		if err != nil {
			return err
		}

		go handleRead(server, id)
		go handleWrite(server, id)
		id += 1
	}
	return nil
}

func (s *TCPServer) Close() {
	log.Printf("Closing server\n")
	for k := range s.in {
		_, err := s.conn[k].Write([]byte("quit"))
		if err != nil {
			log.Fatalf("Failed to close server correctly")
		}
		s.conn[k].Close()
	}
	for k := range s.in {
		s.quitRead[k] <- true
		s.quitWrite[k] <- true
	}
	s.listener.Close()
}

func (s *TCPServer) GetReadChannels() map[int]chan []byte {
	return s.in
}

func (s *TCPServer) GetWriteChannels() map[int]chan []byte {
	return s.out
}

func handleRead(s *TCPServer, connId int) {
	defer close(s.in[connId])
	loop:
	for {
		buf := make([]byte, s.serverMaxReceiveSize, s.serverMaxReceiveSize)
		read, err := s.conn[connId].Read(buf)
		if err != nil {
			// log.Printf("ERROR: connId = %d, read = %v, err = %s\n", connId, read, err)
			break loop
		}
		select {
		case <-s.quitRead[connId]:
			return

		case s.in[connId] <- buf[:read]:
		}
	}
	<-s.quitRead[connId]
}

func handleWrite(s *TCPServer, connId int) {
	var buf []byte
	loop:
	for {
		select {
		case <-s.quitWrite[connId]:
			return
		case buf = <-s.out[connId]:
		}
		_, err := s.conn[connId].Write(buf)
		if err != nil {
			// log.Printf("ERROR: connId = %d, written = %v, err = %s\n", connId, written, err)
			break loop
		}
	}
	<-s.quitWrite[connId]
}
