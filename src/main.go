package main

import (
	"HomeExam/src/PointSalad"
	"flag"
	"log"
)

type Game interface {
	Init(playerNum int, botNum int)
	RunHost(in map[int]chan []byte, out map[int]chan []byte)
	RunPlayer(in chan []byte, out chan []byte)
}

type Client interface {
	Connect(hostname string, port string) error
	Close()
	GetReadChannel() chan []byte
	GetWriteChannel() chan []byte
}

type Server interface {
	Init(port string, playerNum int) error
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
	game = &PointSalad.GameState{}
	if isServer {

		game.Init(playerNum, botNum)

		var server Server
		server = &TCPServer{}
		server.Init(port, playerNum)
		defer server.Close()

		game.RunHost(server.GetReadChannels(), server.GetWriteChannels())

	} else {
		var client Client
		client = &TCPClient{}

		err := client.Connect(hostname, port)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		defer client.Close()
		game.RunPlayer(client.GetReadChannel(), client.GetWriteChannel())
	}
}
