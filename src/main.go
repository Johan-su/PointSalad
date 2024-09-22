package main

import (
	"bufio"
	"fmt"
	"os"
	"log"
	"strconv"
	"encoding/json"
	"strings"
	"math/rand"
)

type CardType int
const (
	CARD_PEPPER CardType = iota
	CARD_LETTUCE CardType = iota
	CARD_CARROT CardType = iota
	CARD_CABBAGE CardType = iota
	CARD_ONION CardType = iota
	CARD_TOMATO CardType = iota
	VEGETABLE_TYPE_NUM = iota
)

const (
	PLAY_PILES_NUM = 3
)


func assert(c bool) {
	if !c {
		s := fmt.Sprintf("assertion failed %v", c)
		panic(s)
	}
}

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

type Card struct {
	Id int
	Vegetable_type CardType
	Vegetable_side bool
}
// actors are players and bots
type GameState struct {
	piles [][]Card
	market [6]Card
	actor_piles [][]Card
	active_actor int
	player_num int
	bot_num int
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


func createGameState(json_cards *JCards, player_num int, bot_num int, seed int64) GameState {
	actor_num := player_num + bot_num
	assert(actor_num >= 2 && actor_num <= 6)

	fmt.Printf("seed = %d\n", seed)
	rand.Seed(seed)

	var ids []int
	for _, jcard := range json_cards.Cards {
		ids = append(ids, jcard.Id)
	}

	per_vegetable_num := actor_num * 3
	var deck []Card
	
	for i := range VEGETABLE_TYPE_NUM {
		rand.Shuffle(len(ids), func(i int, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		for j := 0; j < per_vegetable_num; j += 1 {
			card := Card{
				Id: ids[j], 
				Vegetable_type: CardType(i), 
				Vegetable_side: false,
			}
			deck = append(deck, card)

		}
	}

	rand.Shuffle(len(deck), func(i int, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	s := GameState{}
	pile_size := len(deck) / PLAY_PILES_NUM
	pile_size_remainder := len(deck) % PLAY_PILES_NUM
	assert(pile_size_remainder == 0)

	for i := range PLAY_PILES_NUM {
		s.piles = append(s.piles, []Card{})
		s.piles[i] = deck[i * pile_size:(i + 1) * pile_size]
	}



	s.active_actor = rand.Intn(actor_num)
	s.player_num = player_num
	s.bot_num = bot_num

	return s
}

func isNullCard(c *Card) bool {
	return c.Id == 0
}

func displayMarket(s *GameState) {
	for i, card := range s.market {
		assert(isNullCard(&card) || card.Vegetable_side)
		if i == 3 {
			fmt.Printf("\n")
			}
		if !isNullCard(&card) {
			fmt.Printf("[%d] %v ", i, card.Vegetable_type)
		} 
	}

	fmt.Printf("TBD\n")

}

func drawFromTop(p []Card) Card {
	assert(len(p) > 0)
	var c Card
	c := p[len(p) - 1]
	p[0:len(p) - 1]
	return c	
}

func flipCardsFromPiles(s *GameState) {
	assert(len(s.piles) == PLAY_PILES_NUM)


	if len(s.piles[0] == 0) {
		s.market[0] = drawFromTop(&s.piles[0])

	} else {
		
		s.market[0] = drawFromBot(&s.piles[0])
	}
		s.market[1] = drawFromTop(&s.piles[0])
		max_pile_index := 
		s.market[0] = drawFromTop(&s.piles[0])
		s.market[1] = drawFromTop(&s.piles[0])


	s.market[2] = s.piles[1][len(s.piles[1]) - 1]
	s.piles[1] = s.piles[1][0:len(s.piles[1]) - 1]

	s.market[3] = s.piles[1][len(s.piles[1]) - 1]
	s.piles[1] = s.piles[1][0:len(s.piles[1]) - 1]

	s.market[3] = s.piles[2][len(s.piles[2]) - 1]
	s.piles[2] = s.piles[2][0:len(s.piles[2]) - 1]

	s.market[4] = s.piles[2][len(s.piles[2]) - 1]
	s.piles[2] = s.piles[2][0:len(s.piles[2]) - 1]

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

	player_num, bot_num, err := getNumPlayerBotConfigInput(reader, "type number of players,bots example 1,1", 2, 6)
	if err != nil {
		log.Fatal(err)
	}
	
	s := createGameState(&json_cards, player_num, bot_num, 0)	

	for true {
		displayMarket(&s)
		s, err := reader.ReadString('\n')
		str := s[0: len(s) - 2]
		if err != nil {
			log.Fatal(err)
		}
		_ = str

	}

}