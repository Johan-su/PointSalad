package main

import (
	"bufio"
	"fmt"
	"os"
	"log"
	"strconv"
	"encoding/json"
	"strings"
)

const (
	CARD_PEPPER = iota
	CARD_LETTUCE = iota
	CARD_CARROT = iota
	CARD_CABBAGE = iota
	CARD_ONION = iota
	CARD_TOMATO = iota
)


type JCriteria struct {
	PEPPER string
	LETTUCE string
	CARROT string
	CABBAGE string
	ONION string
	TOMATO string
}

type JCard struct {
	Id int
	Criteria JCriteria

} 

type JCards struct {
	Cards []JCard
}

func getNumPlayerBotConfigInput(reader *bufio.Reader, prompt string, min int, max int) (int, int, error) {
	var player_num int
	var bot_num int
	var err error
	fmt.Print(prompt)
	fmt.Printf(" [%d-%d]\n", min, max)
	for true {
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		str := s[0: len(s) - 2]
		strs := strings.Split(str, ",")

		player_num, err = strconv.Atoi(strs[0])
		if err != nil {
			return 0, 0, err
		}
		bot_num, err = strconv.Atoi(strs[1])
		if err != nil {
			return 0, 0, err
		}

		actor_num := player_num + bot_num
		if actor_num >= min && actor_num <= max {
			break
		} else {
			fmt.Printf("sum of %d + %d not within bounds [%d-%d]\n", player_num, bot_num, min, max)
		}
	}

	return player_num, bot_num, err
}

func main() {
	data, err := os.ReadFile("PointSaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}
	
	var json_cards JCards
	
	
	err = json.Unmarshal(data, &json_cards)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(os.Stdin)

	player_num, bot_num, err := getNumPlayerBotConfigInput(reader, "type number of players,bots example 0,1", 2, 6)
	if err != nil {
		log.Fatal(err)
	}
	actor_num := player_num + bot_num
	
	

	for true {
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(str)

	}

}