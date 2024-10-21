package game

import (
	"HomeExam/game/pointsalad"
)

// Game defines the interface for a game that can be initialized, run in a host or player mode, and provides information about
// the maximum data size allowed for host and player communication.
//
// Methods:
//   - Init(playerNum int, botNum int): Initializes the game with a specified number of players and bots.
//   - RunHost(in map[int]chan []byte, out map[int]chan []byte): Starts the game in host mode, managing communication between players and bots.
//   - RunPlayer(in chan []byte, out chan []byte): Starts the game in player mode, allowing a human player to interact with the game.
//   - GetMaxHostDataSize(): Returns the maximum data size that can be received by the host (server).
//   - GetMaxPlayerDataSize(): Returns the maximum data size that can be sent by the player (client).
type Game interface {
	Init(playerNum int, botNum int)
	RunHost(in map[int]chan []byte, out map[int]chan []byte)
	RunPlayer(in chan []byte, out chan []byte)
	GetMaxHostDataSize() int
	GetMaxPlayerDataSize() int
}

func CreatePointSalad() Game {
	return &pointsalad.GameState{}
}