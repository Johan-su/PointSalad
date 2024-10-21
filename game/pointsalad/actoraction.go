package pointsalad

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type ActorActionType int

const (
	INVALID                ActorActionType = iota
	PICK_VEG_FROM_MARKET   ActorActionType = iota
	PICK_POINT_FROM_MARKET ActorActionType = iota
	PICK_TO_SWAP           ActorActionType = iota
	QUIT                   ActorActionType = iota
)

type ActorAction struct {
	kind   ActorActionType
	amount int
	ids    [2]int
}

// getMarketActionFromBot generates a random market action for a bot player. The bot either chooses 
// to pick vegetables or point cards from the market. It performs this action only if the action is legal
// and results in an equal or better score compared to the current game state. It ensures that the action 
// chosen is beneficial by simulating the effect of the action before finalizing it.
//
// Parameters:
//   - s: The current game state (GameState) to evaluate the action on.
//
// Returns:
//   - An ActorAction representing the bot's decision on the market (either picking vegetables or point cards).
//
// This function ensures that the bot chooses an action that is legal and maximizes or maintains the bot's score.
func getMarketActionFromBot(s *GameState) ActorAction {
	marketWidth := getMarketWidth(&s.market)
	marketHeight := getMarketHeight(&s.market)
	var action ActorAction
	for {
		action = ActorAction{}
		if rand.Intn(2) == 0 {
			action.kind = PICK_VEG_FROM_MARKET
			action.amount = rand.Intn(2) + 1
			for i := range action.amount {
				action.ids[i] = rand.Intn(marketWidth * marketHeight)
			}
		} else {
			action.kind = PICK_POINT_FROM_MARKET
			action.amount = 1
			action.ids[0] = rand.Intn(marketWidth)
		}
		err := isActionLegal(s, action)
		if err == nil {
			beforeScore := calculateScore(s, s.activeActor)

			new_s := deepCloneGameState(s)

			assert(fmt.Sprintf("%v", new_s) == fmt.Sprintf("%v", *s))
			doAction(&new_s, action)
			assert(fmt.Sprintf("%v", new_s) != fmt.Sprintf("%v", *s))

			AfterScore := calculateScore(&new_s, new_s.activeActor)

			if AfterScore >= beforeScore {
				break
			}
		}
	}
	assert(isActionLegal(s, action) == nil)
	return action
}

// getSwapActionFromBot generates a random swap action for a bot player. The bot chooses to swap one or 
// two point cards from their point pile with vegetables from the market. The action is only accepted 
// if it is legal and results in a score that is equal to or greater than the previous score. 
// The bot makes sure the chosen swap is beneficial by simulating the result first.
//
// Parameters:
//   - s: The current game state (GameState) to evaluate the swap action on.
//
// Returns:
//   - An ActorAction representing the bot's decision to swap point cards for vegetables.
//
// This function ensures the bot chooses a swap action that is legal and maximizes or maintains the bot's score.
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

		err := isActionLegal(s, action)
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

// parseMarketActionFromPlayer parses a player's input for a market action. The input specifies
// the player's choice of picking point cards or vegetables from the market. The function interprets
// the input based on the format and returns an ActorAction representing the player's decision.
// If the input is invalid or the action is illegal, an error is returned.
//
// Parameters:
//   - s: The current game state (GameState).
//   - input: A slice of bytes representing the player's input.
//
// Returns:
//   - ActorAction: The action generated based on the player's input.
//   - error: An error if the input is invalid or the action is not legal.
//
// The input can be:
//   - A single digit ('0'-'9') to pick a point card from the market.
//   - A single letter ('A'-'F') to pick a vegetable from the market.
//   - Two letters ('A'-'F') to pick two vegetables from the market.
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
	err := isActionLegal(s, action)
	if err != nil {
		return action, err
	}
	return action, nil
}

// parseSwapActionFromPlayer parses a player's input for a swap action. The player can either 
// choose to skip the swap ('n') or specify an index to swap a point card with a vegetable from the market.
// If the input is invalid or the swap is not legal, an error is returned.
//
// Parameters:
//   - s: The current game state (GameState).
//   - input: A slice of bytes representing the player's input.
//
// Returns:
//   - ActorAction: The action generated based on the player's input (either no swap or a specific swap).
//   - error: An error if the input is invalid or the swap is not legal.
//
// The input can be:
//   - 'n' to indicate no swap.
//   - A number to indicate the index of the point card the player wants to swap.
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
	err := isActionLegal(s, action)
	if err != nil {
		return action, err
	}
	return action, nil
}

// isActionLegal validates whether the provided action is legal within the current game state.
// It checks that the action's parameters (e.g., amount, ids) are within valid ranges and that the action
// can be performed given the current state of the market and the player's resources.
// If any part of the action is illegal, an error is returned. Otherwise, the action is deemed legal.
//
// Parameters:
//   - s: The current game state (GameState).
//   - action: The action to be validated (ActorAction).
//
// Returns:
//   - error: Returns an error if the action is illegal; otherwise, returns nil if the action is legal.
func isActionLegal(s *GameState, action ActorAction) error {
	marketWidth := getMarketWidth(&s.market)
	marketSize := marketWidth * getMarketHeight(&s.market)
	switch action.kind {
	case INVALID:
		return fmt.Errorf("Invalid action kind")
	case PICK_VEG_FROM_MARKET:
		{
			if action.amount < 1 || action.amount > 2 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			for i := range action.amount {
				if action.ids[i] < 0 || action.ids[i] >= marketSize {
					return fmt.Errorf("Cannot take card outside of market range")
				}
				if !hasCard(&s.market, action.ids[i]) {
					return fmt.Errorf("Cannot take card from empty market spot")
				}

				for j := range action.amount {
					if i == j {
						continue
					}
					if action.ids[i] == action.ids[j] {
						return fmt.Errorf("Cannot take from the same spot multiple times")
					}
				}

			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			if action.amount != 1 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			if action.ids[0] < 0 || action.ids[0] >= marketWidth {
				return fmt.Errorf("Cannot take card outside of pile range")
			}
			if len(s.market.piles[action.ids[0]]) == 0 {
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

// doAction performs the specified action on the current game state. It mutates the game state based on
// the kind of action (e.g., picking vegetables, picking point cards, or swapping). The action must be validated
// before this function is called (i.e., `isActionLegal` must return nil).
//
// Parameters:
//   - s: The current game state (GameState).
//   - action: The action to be performed (ActorAction).
//
// This function does not return any value. It modifies the game state directly based on the action kind.
func doAction(s *GameState, action ActorAction) {
	assert(isActionLegal(s, action) == nil)
	switch action.kind {
	case INVALID:
		panic("unreachable")
	case PICK_VEG_FROM_MARKET:
		{
			for i := range action.amount {
				card := getCardFromMarket(&s.market, action.ids[i])
				s.actorData[s.activeActor].vegetableNum[card.vegType] += 1
				s.market.cardSpots[action.ids[i]].hasCard = false
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			for i := range action.amount {
				card := drawFromTop(&s.market, action.ids[i])
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

func getActionString(s *GameState, action ActorAction) string {
	assert(isActionLegal(s, action) == nil)
	builder := strings.Builder{}

	builder.WriteString("---- Action ----\n")
	switch action.kind {
	case INVALID:
		panic("unreachable")
	case PICK_VEG_FROM_MARKET:
		{
			for i := range action.amount {
				builder.WriteString(fmt.Sprintf("Player %d drew %v from market\n", s.activeActor, getCardFromMarket(&s.market, action.ids[i]).vegType.String()))
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			for i := range action.amount {
				pile := s.market.piles[action.ids[i]]
				card := pile[len(pile)-1]
				criteria := card.criteria.String()
				builder.WriteString(fmt.Sprintf("Player %d drew %v from market\n", s.activeActor, criteria))
			}
		}
	case PICK_TO_SWAP:
		{
			if action.amount == 0 {
				builder.WriteString(fmt.Sprintf("Player %d did not swap any card\n", s.activeActor))
			} else {
				for i := range action.amount {
					card := s.actorData[s.activeActor].pointPile[action.ids[i]]
					criteria := card.criteria.String()
					builder.WriteString(fmt.Sprintf("Player %d swapped %v to %v\n", s.activeActor, criteria, card.vegType))
				}
			}
		}
	}
	return builder.String()
}
