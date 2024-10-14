package main

import (
	"log"
	"net"
)

type TCPClient struct {
	conn              net.Conn
	in                chan []byte
	out               chan []byte
	quitRead          chan bool
	quitWrite         chan bool
	clientReceiveSize int
}

func (c *TCPClient) Connect(hostname string, port string) error {
	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		return err
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
	c.conn.Close()
	c.quitRead <- true
	c.quitWrite <- true
	close(c.in)
	// closing out is probably not necessary
	// close(c.out)
}

func (c *TCPClient) GetReadChannel() chan []byte {
	return c.in
}

func (c *TCPClient) GetWriteChannel() chan []byte {
	return c.out
}

func (c *TCPClient) handleRead() {
	running := true
loop:
	for running {
		buf := make([]byte, c.clientReceiveSize, c.clientReceiveSize)
		_, err := c.conn.Read(buf)
		if err != nil {
			running = false
			log.Printf("ERROR: %v\n", err)
		}
		select {
		case <-c.quitRead:
			{
				break loop
			}
		case c.in <- buf:
		}
	}
}

func (c *TCPClient) handleWrite() {
	running := true
loop:
	for running {
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
			running = false
			log.Printf("ERROR: %v\n", err)
		}
	}
}
