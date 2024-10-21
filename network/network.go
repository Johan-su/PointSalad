package network

import (
	"HomeExam/network/tcp"
)

type Client interface {
	Connect(hostname string, port string, clientMaxReceiveSize int) error
	Close()
	GetReadChannel() chan []byte
	GetWriteChannel() chan []byte
}

type Server interface {
	Init(port string, playerNum int, serverMaxReceiveSize int) error
	Close()
	GetReadChannels() map[int]chan []byte
	GetWriteChannels() map[int]chan []byte
}

func CreateTCPServerClient() (Server, Client) {
	return &tcp.Server{}, &tcp.Client{}
}
