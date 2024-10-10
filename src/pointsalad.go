package main

import (
	"log"
	"os"
	"encoding/json"
	"slices"
	"fmt"
)

// actors are players and bots
type PointSalad struct {
	str_criterias []string
	criteria_table []Criteria
	// market
	piles [][]Card
	market [6]CardSpot

	//
	actor_data []ActorData
	active_actor int
	player_num int
	bot_num int
}

func (state *PointSalad) init(player_num int, bot_num int) {
	actor_num := player_num + bot_num

	if !(actor_num >= 2 && actor_num <= 6) {
		log.Fatalf("number of players + bots has to be between 2-6\n")
	}

	data, err := os.ReadFile("PointSaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}
	
	json_cards := JCards{}
	
	err = json.Unmarshal(data, &json_cards)
	if err != nil {
		log.Fatal(err)
	}

	{
		game_state, err := createPointSalad(&json_cards, player_num, bot_num, 0)	
		if err != nil {
			log.Fatalf("ERROR: Failed to create game state: %s\n", err)
			return
		}
		*state = game_state
	}
}

func (state *PointSalad) update(in map[int]chan []byte, out map[int]chan []byte) {
	flipCardsFromPiles(state)
	displayActorCards(state, out)
	displayMarket(state, out)
	// get decisions from actor


	is_bot := in[state.active_actor] == nil

	var market_action ActorAction
	if is_bot {
		market_action = getMarketActionFromBot(state)
	} else {
		market_action = getMarketActionFromPlayer(state, in[state.active_actor], out[state.active_actor])
	}
	BroadcastAction(state, market_action, out)
	doAction(state, market_action)

	if (len(state.actor_data[state.active_actor].point_pile) > 0) {
		var swap_action ActorAction
		if is_bot {
			swap_action = getSwapActionFromBot(state)
		} else {
			swap_action = getSwapActionFromPlayer(state, in[state.active_actor], out[state.active_actor])
		}
		BroadcastAction(state, swap_action, out)
		doAction(state, swap_action)
	}


	
	// check win condition
	all_empty := true
	for i := range state.piles {
		if len(state.piles[i]) != 0 {
			all_empty = false
			break
		}
	}
	// print winner if all piles are empty
	if all_empty {
		type Score struct {
			score int 
			actor_id int
		}
		scores := []Score{}
		
		for i := range state.player_num + state.bot_num {
			scores = append(scores, Score{score: calculateScore(state, i), actor_id: i})
		}

		slices.SortFunc(scores, func(a, b Score) int {
			return a.score - b.score
		})
		
		for i, s := range scores {
			broadcast_to_all(out, fmt.Sprintf("%d %d", s.score, s.actor_id))
			if i == 0 {
				broadcast_to_all(out, " Winner\n")
			} else {
				broadcast_to_all(out, "\n")
			}
		}
	}
		
	// next actor
	state.active_actor += 1
	state.active_actor %= state.player_num + state.bot_num
}