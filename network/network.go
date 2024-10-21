package network

import (
	"HomeExam/network/tcp"
)

// Client defines the interface for a network client in the game.
// The client is responsible for connecting to a server, sending and receiving data over the network, 
// and properly closing the connection when done.
//
// Methods:
//   - Connect(hostname string, port string, clientMaxReceiveSize int): Establishes a connection to the server
//     on the given hostname and port, with a maximum data receive size for incoming messages.
//   - Close(): Closes the connection to the server.
//   - GetReadChannel(): Returns the channel for receiving data from the server (as a byte slice).
//   - GetWriteChannel(): Returns the channel for sending data to the server (as a byte slice).
type Client interface {
	Connect(hostname string, port string, clientMaxReceiveSize int) error
	Close()
	GetReadChannel() chan []byte
	GetWriteChannel() chan []byte
}

// Server defines the interface for a network server in the game.
// The server is responsible for accepting client connections, handling communication with clients, 
// and managing multiple connections at once.
//
// Methods:
//   - Init(port string, playerNum int, serverMaxReceiveSize int): Initializes the server to listen on the given
//     port with the specified number of players and maximum receive size for incoming data.
//   - Close(): Closes the server, stopping all communication and accepting no further connections.
//   - GetReadChannels(): Returns a map of channels used for receiving data from each connected client, 
//     keyed by the client ID.
//   - GetWriteChannels(): Returns a map of channels used for sending data to each connected client, 
//     keyed by the client ID.
type Server interface {
	Init(port string, playerNum int, serverMaxReceiveSize int) error
	Close()
	GetReadChannels() map[int]chan []byte
	GetWriteChannels() map[int]chan []byte
}

// CreateTCPServerClient creates and returns both a TCP server and a TCP client, 
// allowing the game to run in a networked environment with server-client communication.
//
// Returns:
//   - A Server object (TCP server) to handle incoming client connections and communication.
//   - A Client object (TCP client) to connect to the server and send/receive data.
func CreateTCPServerClient() (Server, Client) {
	return &tcp.Server{}, &tcp.Client{}
}
