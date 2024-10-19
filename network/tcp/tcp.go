package tcp

import (
	"fmt"
	"log"
	"net"
)

const (
	pingMagic = "ABCZ"
	pongMagic = "ZCBA"
)

type Client struct {
	conn                 net.Conn
	in                   chan []byte
	out                  chan []byte
	quitRead             chan bool
	quitWrite            chan bool
	clientMaxReceiveSize int
}

func (c *Client) Connect(hostname string, port string, clientMaxReceiveSize int) error {
	c.clientMaxReceiveSize = clientMaxReceiveSize
	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		return err
	}

	buf := []byte(pingMagic)
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}
	buf = make([]byte, len(pongMagic))
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}
	if string(buf) != pongMagic {
		return fmt.Errorf("Failed ping pong test\n")
	}

	c.conn = conn
	c.in = make(chan []byte)
	c.out = make(chan []byte)
	c.quitRead = make(chan bool)
	c.quitWrite = make(chan bool)

	go c.handleRead()
	go c.handleWrite()
	return nil
}

func (c *Client) Close() {
	log.Printf("Closing client\n")
	c.conn.Close()
	c.quitRead <- true
	c.quitWrite <- true
}

func (c *Client) GetReadChannel() chan []byte {
	return c.in
}

func (c *Client) GetWriteChannel() chan []byte {
	return c.out
}

func (c *Client) handleRead() {
	for {
		buf := make([]byte, c.clientMaxReceiveSize)
		n, err := c.conn.Read(buf)
		if err != nil {
			// log.Printf("ERROR: %v\n", err)
			close(c.in)
			<-c.quitRead
			return
		}
		select {
		case <-c.quitRead:
			return
		case c.in <- buf[:n]:
		}
	}
}

func (c *Client) handleWrite() {
loop:
	for {
		var valToSend []byte
		select {
		case <-c.quitWrite:
			return
		case valToSend = <-c.out:
		}
		_, err := c.conn.Write(valToSend)
		if err != nil {
			// log.Printf("ERROR: %v\n", err)
			break loop
		}
	}
	<-c.quitWrite
}

type Server struct {
	conn      []net.Conn
	out       map[int]chan []byte
	in        map[int]chan []byte
	quitRead  map[int]chan bool
	quitWrite map[int]chan bool

	serverMaxReceiveSize int
	listener             net.Listener
}

func (server *Server) Init(port string, playerNum int, serverMaxReceiveSize int) error {
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

		buf := make([]byte, len(pingMagic))
		_, err = conn.Read(buf)
		if err != nil {
			return err
		}
		if string(buf) != pingMagic {
			return fmt.Errorf("expected ping pong test\n")
		}
		buf = []byte(pongMagic)
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

func (s *Server) Close() {
	log.Printf("Closing server\n")
	for k := range s.in {
		err := s.conn[k].Close()
		if err != nil {
			log.Fatalf("Failed to close server correctly")
		}
	}
	for k := range s.in {
		s.quitRead[k] <- true
		s.quitWrite[k] <- true
	}
	s.listener.Close()
}

func (s *Server) GetReadChannels() map[int]chan []byte {
	return s.in
}

func (s *Server) GetWriteChannels() map[int]chan []byte {
	return s.out
}

func handleRead(s *Server, connId int) {
	for {
		buf := make([]byte, s.serverMaxReceiveSize, s.serverMaxReceiveSize)
		read, err := s.conn[connId].Read(buf)
		if err != nil {
			close(s.in[connId])
			<-s.quitRead[connId]
			return
		}
		select {
		case <-s.quitRead[connId]:
			return
		case s.in[connId] <- buf[:read]:
		}
	}
}

func handleWrite(s *Server, connId int) {
	var buf []byte
	for {
		select {
		case <-s.quitWrite[connId]:
			return
		case buf = <-s.out[connId]:
		}
		_, err := s.conn[connId].Write(buf)
		if err != nil {
			<-s.quitWrite[connId]
			return
		}
	}
}
