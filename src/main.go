package main

import (
	"HomeExam/src/network"
	"HomeExam/src/pointsalad"
	"flag"
	"log"
)

type Game interface {
	Init(playerNum int, botNum int)
	RunHost(in map[int]chan []byte, out map[int]chan []byte)
	RunPlayer(in chan []byte, out chan []byte)
	GetMaxHostDataSize() int
	GetMaxPlayerDataSize() int
}

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

func main() {
	var isServer bool
	var hostname string
	var port string
	var playerNum int
	var botNum int

	flag.BoolVar(&isServer, "server", false, "ex. -server")
	flag.StringVar(&hostname, "hostname", "127.0.0.1", "ex. 127.0.0.1")
	flag.StringVar(&port, "port", "8080", "ex. 8080")
	flag.IntVar(&playerNum, "players", 1, "ex. 2")
	flag.IntVar(&botNum, "bots", 1, "ex. 2")
	flag.Parse()

	log.Printf("isServer = %v, hostname = %v port = %v playerNum = %v botNum = %v\n", isServer, hostname, port, playerNum, botNum)
	var game Game
	game = &pointsalad.GameState{}
	if isServer {

		game.Init(playerNum, botNum)

		var server Server
		server = &network.TCPServer{}
		err := server.Init(port, playerNum, game.GetMaxHostDataSize())
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		game.RunHost(server.GetReadChannels(), server.GetWriteChannels())
		server.Close()

	} else {
		var client Client
		client = &network.TCPClient{}

		err := client.Connect(hostname, port, game.GetMaxPlayerDataSize())
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		game.RunPlayer(client.GetReadChannel(), client.GetWriteChannel())
		client.Close()
	}
}
