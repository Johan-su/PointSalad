package game

import (
	"HomeExam/game/pointsalad"
)

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