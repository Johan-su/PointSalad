package main

import (
	"log"
	"net"
	"fmt"
)

type TCPClient struct {
	conn              net.Conn
	in                chan []byte
	out               chan []byte
	quitRead          chan bool
	quitWrite         chan bool
	clientMaxReceiveSize int
}

func (c *TCPClient) Connect(hostname string, port string, clientMaxReceiveSize int) error {
	c.clientMaxReceiveSize = clientMaxReceiveSize
	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		return err
	}

	buf := make([]byte, 4, 4)
	buf = []byte("ping")
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}
	buf = make([]byte, 4, 4)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}
	if string(buf) != "pong" {
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

func (c *TCPClient) Close() {
	log.Printf("Closing client\n")
	c.conn.Close()
	close(c.out)
	c.quitRead <- true
	c.quitWrite <- true
}

func (c *TCPClient) GetReadChannel() chan []byte {
	return c.in
}

func (c *TCPClient) GetWriteChannel() chan []byte {
	return c.out
}

func (c *TCPClient) handleRead() {
loop:
	for true {
		buf := make([]byte, c.clientMaxReceiveSize, c.clientMaxReceiveSize)
		n, err := c.conn.Read(buf)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			break loop
		}
		select {
		case <-c.quitRead: break loop
		case c.in <- buf[:n]:
		}
	}
}

func (c *TCPClient) handleWrite() {
loop:
	for true {
		var valToSend []byte
		select {
		case <-c.quitWrite:
			{
				break loop
			}
		case valToSend = <-c.out:
		}
		_, err := c.conn.Write(valToSend)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			break loop
		}
	}
}
