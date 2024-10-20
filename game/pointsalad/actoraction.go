package pointsalad

import (
	"math/rand"
	"fmt"
	"strconv"
)

type ActorActionType int

const (
	INVALID                ActorActionType = iota
	PICK_VEG_FROM_MARKET   ActorActionType = iota
	PICK_POINT_FROM_MARKET ActorActionType = iota
	PICK_TO_SWAP           ActorActionType = iota
	QUIT ActorActionType = iota
)

type ActorAction struct {
	kind   ActorActionType
	amount int
	ids    [2]int
}

func getMarketActionFromBot(s *GameState) ActorAction {
	var action ActorAction
	for {
		action = ActorAction{}
		if rand.Intn(2) == 0 {
			action.kind = PICK_VEG_FROM_MARKET
			action.amount = rand.Intn(2) + 1
			for i := range action.amount {
				action.ids[i] = rand.Intn(len(s.market))
			}
		} else {
			action.kind = PICK_POINT_FROM_MARKET
			action.amount = 1
			action.ids[0] = rand.Intn(len(s.piles))
		}
		err := IsActionLegal(s, action)
		if err == nil {
			beforeScore := calculateScore(s, s.activeActor)

			new_s := deepCloneGameState(s)

			doAction(&new_s, action)

			AfterScore := calculateScore(&new_s, new_s.activeActor)

			if AfterScore >= beforeScore {
				break
			}
		}
	}
	return action
}

func getSwapActionFromBot(s *GameState) ActorAction {
	assert(len(s.actorData[s.activeActor].pointPile) > 0)

	action := ActorAction{}
	for true {
		action.kind = PICK_TO_SWAP
		action.amount = rand.Intn(2)

		for i := range action.amount {
			n := len(s.actorData[s.activeActor].pointPile)
			action.ids[i] = rand.Intn(n)
		}

		err := IsActionLegal(s, action)
		if err == nil {
			beforeScore := calculateScore(s, s.activeActor)

			new_s := deepCloneGameState(s)

			doAction(&new_s, action)

			AfterScore := calculateScore(&new_s, new_s.activeActor)

			if AfterScore >= beforeScore {
				break
			}
		}
	}
	return action
}

func isWithinAtoF(a byte) bool {
	return a >= 'A' && a <= 'F'
}

func parseMarketActionFromPlayer(s *GameState, input []byte) (ActorAction, error) {
	action := ActorAction{}

	if len(input) == 1 && input[0] >= '0' && input[0] <= '9' {
		index := int(input[0] - '0')
		action = ActorAction{kind: PICK_POINT_FROM_MARKET, amount: 1, ids: [2]int{index, 0}}

	} else if len(input) == 1 && isWithinAtoF(input[0]) {
		index := int(input[0] - 'A')
		action = ActorAction{kind: PICK_VEG_FROM_MARKET, amount: 1, ids: [2]int{index, 0}}

	} else if len(input) == 2 && isWithinAtoF(input[0]) && isWithinAtoF(input[1]) {
		indicies := [2]int{int(input[0] - 'A'), int(input[1] - 'A')}
		action = ActorAction{kind: PICK_VEG_FROM_MARKET, amount: 2, ids: indicies}

	} else {
		return action, fmt.Errorf("Invalid input")
	}
	err := IsActionLegal(s, action)
	if err != nil {
		return action, err
	}
	return action, nil
}

func parseSwapActionFromPlayer(s *GameState, input []byte) (ActorAction, error) {
	action := ActorAction{}

	if input[0] == 'n' {
		action = ActorAction{kind: PICK_TO_SWAP, amount: 0}
	} else {
		index, err := strconv.Atoi(string(input))
		if err != nil {
			return action, fmt.Errorf("Expected a number or 'n'\n") 
		}
		action = ActorAction{kind: PICK_TO_SWAP, amount: 1, ids: [2]int{index, 0}}
	}
	err := IsActionLegal(s, action)
	if err != nil {
		return action, err
	}
	return action, nil
}

func IsActionLegal(s *GameState, action ActorAction) error {
	switch action.kind {
	case INVALID:
		return fmt.Errorf("Invalid action kind")
	case PICK_VEG_FROM_MARKET:
		{
			if action.amount < 1 || action.amount > 2 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			for i := range action.amount {
				if action.ids[i] < 0 || action.ids[i] >= len(s.market) {
					return fmt.Errorf("Cannot take card outside of market range")
				}
				if !s.market[action.ids[i]].hasCard {
					return fmt.Errorf("Cannot take card from empty market spot")
				}
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			if action.amount != 1 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			if action.ids[0] < 0 || action.ids[0] >= len(s.piles) {
				return fmt.Errorf("Cannot take card outside of pile range")
			}
			if len(s.piles[action.ids[0]]) == 0 {
				return fmt.Errorf("Cannot take card from empty pile")
			}
		}
	case PICK_TO_SWAP:
		{
			if action.amount < 0 || action.amount > 1 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			if action.amount == 1 {
				if action.ids[0] < 0 || action.ids[0] >= len(s.actorData[s.activeActor].pointPile) {
					return fmt.Errorf("Cannot take card outside of pile range")
				}
			}
		}
	}
	return nil
}

func doAction(s *GameState, action ActorAction) {
	assert(IsActionLegal(s, action) == nil)

	switch action.kind {
	case INVALID:
		panic("unreachable")
	case PICK_VEG_FROM_MARKET:
		{
			for i := range action.amount {
				card := s.market[action.ids[i]].card
				s.actorData[s.activeActor].vegetableNum[card.vegType] += 1
				s.market[action.ids[i]].hasCard = false
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			for i := range action.amount {
				card := drawFromTop(s, action.ids[i])
				s.actorData[s.activeActor].pointPile = append(s.actorData[s.activeActor].pointPile, card)
			}
		}
	case PICK_TO_SWAP:
		{
			if action.amount == 1 {
				veg_type := s.actorData[s.activeActor].pointPile[action.ids[0]].vegType
				s.actorData[s.activeActor].vegetableNum[int(veg_type)] += 1

				// remove element
				for i := action.ids[0]; i < len(s.actorData[s.activeActor].pointPile)-1; i += 1 {
					s.actorData[s.activeActor].pointPile[i] = s.actorData[s.activeActor].pointPile[i+1]
				}
				s.actorData[s.activeActor].pointPile = s.actorData[s.activeActor].pointPile[0 : len(s.actorData[s.activeActor].pointPile)-1]
			}
		}
	}
}

