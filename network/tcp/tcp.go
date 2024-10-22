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

// Connect establishes a TCP connection to the specified host and port,
// performs a ping-pong test to verify the connection, and initializes
// channels for reading and writing data. The function also starts two
// goroutines for handling reading and writing concurrently.
//
// Parameters:
// - hostname: The target host to connect to.
// - port: The target port to connect to.
// - clientMaxReceiveSize: The maximum size for receiving data from the server.
//
// Returns:
// - error: An error if the connection or ping-pong test fails, or nil if successful.
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

// Close terminates the client's connection by closing the TCP connection,
// and sends signals to stop the reading and writing goroutines. It logs
// the closure of the client connection.
//
// This function gracefully shuts down the client by ensuring that the
// connection is properly closed and that no further read/write operations
// are attempted.
//
// Returns:
// - None
func (c *Client) Close() {
	log.Printf("Closing client\n")
	c.conn.Close()
	c.quitRead <- true
	c.quitWrite <- true
}

// GetReadChannel returns the channel used for reading data from the client connection.
// This allows other parts of the application to receive data that the client reads
// from the server over the established connection.
//
// Returns:
// - chan []byte: The channel for receiving incoming data.
func (c *Client) GetReadChannel() chan []byte {
	return c.in
}

// GetWriteChannel returns the channel used for writing data to the client connection.
// This allows other parts of the application to send data to the server through
// the client over the established connection.
//
// Returns:
// - chan []byte: The channel for sending outgoing data.
func (c *Client) GetWriteChannel() chan []byte {
	return c.out
}

// handleRead continuously reads data from the client's connection in a loop,
// and sends the data to the read channel (`c.in`). If an error occurs while reading
// or the read operation is interrupted, it closes the read channel and stops the goroutine.
// The function stops reading when the quitRead channel is signaled.
//
// It uses a buffered slice to store the incoming data, and selects between
// quitting or sending the data to the read channel.
//
// Returns:
// - None
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

// handleWrite continuously listens for data on the write channel (`c.out`)
// and writes it to the client's connection. If an error occurs while writing,
// the loop is broken, and the function terminates. It stops writing when the quitWrite
// channel is signaled.
//
// The function uses a loop to wait for data to write and checks for a quit signal
// to stop the writing process. Once the data is successfully written, the function
// proceeds to the next data or terminates if an error occurs.
//
// Returns:
// - None
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

// Listen initializes the server to listen on the specified port and accepts connections
// from a predefined number of players. It sets up channels for communication with each player
// and starts goroutines to handle reading and writing data from/to each connected client.
//
// Parameters:
// - port: The port on which the server listens for incoming connections.
// - playerNum: The number of players (clients) the server expects to connect.
// - serverMaxReceiveSize: The maximum size for receiving data from clients.
//
// Returns:
//   - error: Returns an error if there is an issue during the server setup or client connections,
//     or nil if the server was successfully initialized and is accepting connections.
func (server *Server) Listen(port string, playerNum int, serverMaxReceiveSize int) error {
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

// Close gracefully shuts down the server by closing all client connections,
// stopping the read and write operations, and closing the server listener.
// It ensures all resources are cleaned up and that the server terminates
// without leaving any open connections or goroutines.
//
// Returns:
// - None
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

// GetReadChannels returns a map of channels used for reading data from each connected client.
// This allows other parts of the application to access the channels and read incoming data
// from the clients.
//
// Returns:
//   - map[int]chan []byte: A map of read channels, where the key is the client ID and the value
//     is the read channel for that client.
func (s *Server) GetReadChannels() map[int]chan []byte {
	return s.in
}

// GetWriteChannels returns a map of channels used for writing data to each connected client.
// This allows other parts of the application to send data to the clients via their respective
// write channels.
//
// Returns:
//   - map[int]chan []byte: A map of write channels, where the key is the client ID and the value
//     is the write channel for that client.
func (s *Server) GetWriteChannels() map[int]chan []byte {
	return s.out
}

// handleRead continuously reads data from a client connection and sends the
// data to the associated read channel (`s.in[connId]`). If an error occurs
// while reading or if the quitRead signal is triggered, it gracefully closes
// the read channel and stops the goroutine.
//
// Parameters:
// - s: The server instance, which holds the client connections and associated channels.
// - connId: The ID of the connection (client) being read from.
//
// Returns:
// - None
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

// handleWrite continuously listens for data to write to a client connection
// and writes the data to the client's connection. If an error occurs while
// writing or if the quitWrite signal is triggered, it stops the goroutine and
// gracefully terminates the write operation.
//
// Parameters:
// - s: The server instance, which holds the client connections and associated channels.
// - connId: The ID of the connection (client) being written to.
//
// Returns:
// - None
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
