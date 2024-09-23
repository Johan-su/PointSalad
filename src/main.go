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
	"slices"
)

type CardType int
const (
	PEPPER CardType = iota
	LETTUCE CardType = iota
	CARROT CardType = iota
	CABBAGE CardType = iota
	ONION CardType = iota
	TOMATO CardType = iota
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

type Criteria struct {
	use_veg [VEGETABLE_TYPE_NUM]bool
}

type Card struct {
	Id int
	Vegetable_type CardType
}
type ActorData struct {
	vegetable_num [VEGETABLE_TYPE_NUM]int
	point_pile []Card
}
// actors are players and bots
type GameState struct {
	piles [][]Card
	market [6]Card
	actor_data []ActorData
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


	for range actor_num {
		s.actor_data = append(s.actor_data, ActorData{})
	}


	s.active_actor = rand.Intn(actor_num)
	s.player_num = player_num
	s.bot_num = bot_num

	return s
}

func isNullCard(c *Card) bool {
	return c.Id == 0
}

func setNullCard(c *Card) {
	c.Id = 0
}

func displayMarket(s *GameState) {
	fmt.Println("---- MARKET ----")
	for i, card := range s.market {
		if !isNullCard(&card) {
			fmt.Printf("[%c] %v\n", i + 'A', card.Vegetable_type)
		} 
	}
	fmt.Println("piles")
	for i := range s.piles {
		if len(s.piles[i]) > 0 {
			fmt.Printf("[%d] ADDCRITERIAHERE\n", i)
		} else {
			fmt.Println("")
		}
	}
}

func drawFromTop(s *GameState, pile_index int) Card {
	assert(len(s.piles[pile_index]) > 0)
	c := s.piles[pile_index][len(s.piles[pile_index]) - 1]
	s.piles[pile_index] = s.piles[pile_index][0:len(s.piles[pile_index]) - 1]
	return c	
}

func drawFromBot(s *GameState, pile_index int) Card {
	assert(len(s.piles[pile_index]) > 0)
	c := s.piles[pile_index][0]
	s.piles[pile_index] = s.piles[pile_index][1:len(s.piles[pile_index])]
	return c
}

func getMaxPileIndex(s *GameState) int {
	max := len(s.piles[0])
	index := 0

	for i, p := range s.piles {
		if len(p) > max {
			max = len(p)
			index = i
		}
	}

	return index
}

func displayActorCards(s *GameState) {
	assert(len(s.actor_data) + 1 > s.active_actor)
	fmt.Printf("---- Actor %d ----\n", s.active_actor)

	fmt.Printf("%d current score\n", calculateScore(s, s.active_actor))
	fmt.Println("--------")
	for i, num := range s.actor_data[s.active_actor].vegetable_num {
		fmt.Printf("%d %v\n", num, CardType(i))
	}

	fmt.Println("---- point cards ----")

	for range s.actor_data[s.active_actor].point_pile {
		fmt.Printf("ADDCRITERIAHERE\n")
	}
}

func flipCardsFromPiles(s *GameState) {
	for y := range s.piles {
		for x := range 2 {
			market_pos := y + x * PLAY_PILES_NUM
			if isNullCard(&s.market[market_pos]) {
				if len(s.piles[y]) == 0 {
					s.market[market_pos] = drawFromTop(s, y)
			
				} else {
					index := getMaxPileIndex(s)
					s.market[market_pos] = drawFromBot(s, index)
				}
			}
		}
	}
}


func isWithinAtoF(a byte) bool {
	return a >= 'A' && a <= 'F'
}

func pickCardsFromMarket(reader *bufio.Reader, s *GameState) {
	out:
	for true {
		fmt.Printf("pick 1 or 2 vegetables example: AB or\npick 1 point card example: 0\n")
		st, err := reader.ReadString('\n')
		str := st[0: len(st) - 2]
		if err != nil {
			log.Fatal(err)
		}
		
		switch len(str) {
			case 1: {
				if str[0] >= '0' && str[0] <= '9' {
					index, err := strconv.Atoi(str)
					if err != nil {
						log.Fatal(err)
					}
					if len(s.piles[index]) > 0 {
						s.actor_data[s.active_actor].point_pile = append(s.actor_data[s.active_actor].point_pile, drawFromTop(s, index))
					} else {
						fmt.Printf("%d pile is empty pick another one\n", index)
						continue out
					}
				} else if isWithinAtoF(str[0]) {
					index := str[0] - 'A'
					c := s.market[index]
					if isNullCard(&c) {
						continue out
					}
					s.actor_data[s.active_actor].vegetable_num[int(c.Vegetable_type)] += 1
					setNullCard(&s.market[index])
				}
				break out
			}
			case 2: {
				if  isWithinAtoF(str[0]) && isWithinAtoF(str[1]) {
					indicies := [2]int{int(str[0]) - 'A', int(str[1]) - 'A'}

					for i, index := range indicies {
						c := s.market[index]
						if isNullCard(&c) {
							fmt.Printf("%c pos is empty pick another one\n", str[i])
							continue out
						}
						s.actor_data[s.active_actor].vegetable_num[int(c.Vegetable_type)] += 1
						setNullCard(&s.market[index])
					}

					break out
				}
			}
		}
	}
}

func pickCardToChangeToVeg(reader *bufio.Reader, s *GameState) {
	for true {
		if len(s.actor_data[s.active_actor].point_pile) == 0 {
			break
		}
		fmt.Printf("pick 0-1 point card, type n to pick none example: 5\n")
		st, err := reader.ReadString('\n')
		str := st[0: len(st) - 2]
		if err != nil {
			log.Fatal(err)
		}
		if len(str) == 0 {
			continue
		}
		if len(str) == 1 && str[0] == 'n' {
			break
		}

		index, err := strconv.Atoi(str)
		if err != nil {
			continue
		}

		if index >= 0 && index < len(s.actor_data[s.active_actor].point_pile) {
			

			card := s.actor_data[s.active_actor].point_pile[index]

			s.actor_data[s.active_actor].vegetable_num[int(card.Vegetable_type)] += 1
			

			// shift slice
			for i := index; i < len(s.actor_data[s.active_actor].point_pile) - 1; i += 1 {
				s.actor_data[s.active_actor].point_pile[i] = s.actor_data[s.active_actor].point_pile[i + 1] 
			}
			// remove last element
			s.actor_data[s.active_actor].point_pile = s.actor_data[s.active_actor].point_pile[0:len(s.actor_data[s.active_actor].point_pile) - 1]
			break
		}
	}
}


func calculateScore(s *GameState, actor_id int) int {
	score := 0

	for i, point_card := range s.actor_data[actor_id].point_pile {
		
		
		if most {



		} else if fewest {

		} else if even_odd {

		} else if per {

		} else if sum {

		} else if most_total {
		} else if fewest_total {

		} else {
			assert(false)
		}
	}



	return score
}

func main() {
	data, err := os.ReadFile("PointSaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}
	
	json_cards := JCards{}
	
	err = json.Unmarshal(data, &json_cards)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(os.Stdin)

	player_num, bot_num, err := getNumPlayerBotConfigInput(reader, "type number of players,bots example 1,1", 2, 6)
	if err != nil {
		log.Fatal(err)
	}

	state := createGameState(&json_cards, player_num, bot_num, 0)	

	for true {
		flipCardsFromPiles(&state)
		displayActorCards(&state)
		displayMarket(&state)
		// get decisions from actor
		pickCardsFromMarket(reader, &state)
		pickCardToChangeToVeg(reader, &state)


		all_empty := true
		for i := range state.piles {
			if len(state.piles[i]) != 0 {
				all_empty = false
				break
			}
		}
		if all_empty {
			type Score struct {
				score int 
				actor_id int
			}
			scores := []Score{}

			for i := range state.player_num + state.bot_num {
				scores = append(scores, Score{score: calculateScore(&state, i), actor_id: i})
			}

			slices.SortFunc(scores, func(a, b Score) int {
				return a.score - b.score
			})

			for i, s := range scores {
				fmt.Printf("%d %d", s.score, s.actor_id)
				if i == 0 {
					fmt.Printf(" Winner\n")
				} else {
					fmt.Printf("\n")
				}
			}
		}

		// next player
		state.active_actor += 1
		state.active_actor %= state.player_num + state.bot_num
	}
}