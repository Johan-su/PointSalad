package main

import (
	"HomeExam/game"
	"HomeExam/network"
	"flag"
	"log"
)


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

	server, client := network.CreateTCPServerClient()

	game := game.CreatePointSalad()
	if isServer {

		game.Init(playerNum, botNum)

		err := server.Init(port, playerNum, game.GetMaxHostDataSize())
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		game.RunHost(server.GetReadChannels(), server.GetWriteChannels())
		server.Close()

	} else {

		err := client.Connect(hostname, port, game.GetMaxPlayerDataSize())
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		game.RunPlayer(client.GetReadChannel(), client.GetWriteChannel())
		client.Close()
	}
}
