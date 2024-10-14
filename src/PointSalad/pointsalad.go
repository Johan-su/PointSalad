package PointSalad

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
)

type VegType int

const (
	PEPPER  VegType = iota
	LETTUCE VegType = iota

	CARROT  VegType = iota
	CABBAGE VegType = iota

	ONION  VegType = iota
	TOMATO VegType = iota

	VEGETABLE_TYPE_NUM = iota
)

const (
	playPilesNum          = 3
	serverByteReceiveSize = 8
	serverByteSendSize    = 512
)

type Card struct {
	id      int
	vegType VegType
}

type ActorData struct {
	vegetableNum [VEGETABLE_TYPE_NUM]int
	pointPile    []Card
}

type CardSpot struct {
	hasCard bool
	card    Card
}

// actors are players and bots
type GameState struct {
	strCriterias  []string
	criteriaTable []Criteria
	// market
	piles  [][]Card
	market [6]CardSpot

	//
	actorData   []ActorData
	activeActor int
	playerNum   int
	botNum      int
}

func (state *GameState) Init(playerNum int, botNum int) {
	actorNum := playerNum + botNum

	if !(actorNum >= 2 && actorNum <= 6) {
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
		game_state, err := createGameState(&json_cards, playerNum, botNum, 0)
		if err != nil {
			log.Fatalf("ERROR: Failed to create game state: %s\n", err)
			return
		}
		*state = game_state
	}
}

func (state *GameState) RunHost(in map[int]chan []byte, out map[int]chan []byte) {
	flipCardsFromPiles(state)
	displayActorCards(state, out[state.activeActor])
	displayMarket(state, out)
	// get decisions from actor

	is_bot := in[state.activeActor] == nil

	var market_action ActorAction
	if is_bot {
		market_action = getMarketActionFromBot(state)
	} else {
		market_action = getMarketActionFromPlayer(state, in[state.activeActor], out[state.activeActor])
	}
	BroadcastAction(state, market_action, out)
	doAction(state, market_action)

	displayActorCards(state, out[state.activeActor])

	if len(state.actorData[state.activeActor].pointPile) > 0 {
		var swap_action ActorAction
		if is_bot {
			swap_action = getSwapActionFromBot(state)
		} else {
			swap_action = getSwapActionFromPlayer(state, in[state.activeActor], out[state.activeActor])
		}
		BroadcastAction(state, swap_action, out)
		doAction(state, swap_action)
	}
	// show hand to players
	for _, v := range out {
		displayActorCards(state, v)
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
			score   int
			actorId int
		}
		scores := []Score{}

		for i := range state.playerNum + state.botNum {
			scores = append(scores, Score{score: calculateScore(state, i), actorId: i})
		}

		slices.SortFunc(scores, func(a, b Score) int {
			return a.score - b.score
		})

		for i, s := range scores {
			broadcast_to_all(out, fmt.Sprintf("%d %d", s.score, s.actorId))
			if i == 0 {
				broadcast_to_all(out, " Winner\n")
			} else {
				broadcast_to_all(out, "\n")
			}
		}
	}

	// next actor
	state.activeActor += 1
	state.activeActor %= state.playerNum + state.botNum
}

func (state *GameState) RunPlayer(in chan []byte, out chan []byte) {
	assert(false)
}

func assert(c bool) {
	if !c {
		s := fmt.Sprintf("assertion failed %v", c)
		panic(s)
	}
}

func createGameState(json_cards *JCards, playerNum int, botNum int, seed int64) (GameState, error) {
	actorNum := playerNum + botNum
	assert(actorNum >= 2 && actorNum <= 6)

	log.Printf("seed = %d\n", seed)
	rand.Seed(seed)

	var ids []int
	for id, _ := range json_cards.Cards {
		ids = append(ids, id)
	}

	per_vegetable_num := actorNum * 3
	var deck []Card

	for i := range VEGETABLE_TYPE_NUM {
		rand.Shuffle(len(ids), func(i int, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		for j := 0; j < per_vegetable_num; j += 1 {
			card := Card{
				id:      ids[j],
				vegType: VegType(i),
			}
			deck = append(deck, card)

		}
	}

	rand.Shuffle(len(deck), func(i int, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	s := GameState{}

	for _, card := range json_cards.Cards {
		s.strCriterias = append(s.strCriterias, card.Criteria.PEPPER)
		s.strCriterias = append(s.strCriterias, card.Criteria.LETTUCE)
		s.strCriterias = append(s.strCriterias, card.Criteria.CARROT)
		s.strCriterias = append(s.strCriterias, card.Criteria.CABBAGE)
		s.strCriterias = append(s.strCriterias, card.Criteria.ONION)
		s.strCriterias = append(s.strCriterias, card.Criteria.TOMATO)
	}

	table, err := createCriteriaTable(json_cards)
	if err != nil {
		return GameState{}, err
	}
	s.criteriaTable = table

	pile_size := len(deck) / playPilesNum
	pile_size_remainder := len(deck) % playPilesNum
	assert(pile_size_remainder == 0)

	for i := range playPilesNum {
		s.piles = append(s.piles, []Card{})
		s.piles[i] = deck[i*pile_size : (i+1)*pile_size]
	}

	for range actorNum {
		s.actorData = append(s.actorData, ActorData{})
	}

	s.activeActor = rand.Intn(actorNum)
	s.playerNum = playerNum
	s.botNum = botNum
	return s, nil
}

type ActorActionType int

const (
	INVALID                ActorActionType = iota
	PICK_VEG_FROM_MARKET   ActorActionType = iota
	PICK_POINT_FROM_MARKET ActorActionType = iota
	PICK_TO_SWAP           ActorActionType = iota
)

type ActorAction struct {
	kind   ActorActionType
	amount int
	ids    [2]int
}

func getMarketActionFromBot(s *GameState) ActorAction {

	action := ActorAction{}
	for true {
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
			break
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
			break
		}
	}
	return action
}

func getMarketActionFromPlayer(s *GameState, in chan []byte, out chan []byte) ActorAction {
	assert(in != nil)
	assert(out != nil)

	action := ActorAction{}
	for true {
		out <- []byte(fmt.Sprintf("pick 1 or 2 vegetables example: AB or\npick 1 point card example: 0\n"))
		input := <-in

		if input[0] >= '0' && input[0] <= '9' && input[1] == 0 {
			index := int(input[0] - '0')
			action = ActorAction{kind: PICK_POINT_FROM_MARKET, amount: 1, ids: [2]int{index, 0}}

		} else if isWithinAtoF(input[0]) && input[1] == 0 {
			index := int(input[0] - 'A')
			action = ActorAction{kind: PICK_VEG_FROM_MARKET, amount: 1, ids: [2]int{index, 0}}

		} else if isWithinAtoF(input[0]) && isWithinAtoF(input[1]) && input[2] == 0 {
			indicies := [2]int{int(input[0] - 'A'), int(input[1] - 'A')}
			action = ActorAction{kind: PICK_VEG_FROM_MARKET, amount: 2, ids: indicies}

		} else {
			continue
		}
		err := IsActionLegal(s, action)
		if err != nil {
			out <- []byte(fmt.Sprintf("%v\n", err))
		} else {
			break
		}
	}
	return action
}

func getSwapActionFromPlayer(s *GameState, in chan []byte, out chan []byte) ActorAction {
	assert(in != nil)
	assert(out != nil)
	assert(len(s.actorData[s.activeActor].pointPile) > 0)

	action := ActorAction{}
	for true {
		out <- []byte(fmt.Sprintf("pick 0-1 point card to flip to vegetable, type n to pick none example: 5\n"))
		input := <-in

		if input[0] == 'n' {
			action = ActorAction{kind: PICK_TO_SWAP, amount: 0}
		} else {
			index, err := strconv.Atoi(string(input))
			if err != nil {
				out <- []byte(fmt.Sprintf("Expected a number or 'n'\n"))
				continue
			}
			action = ActorAction{kind: PICK_TO_SWAP, amount: 1, ids: [2]int{index, 0}}
		}
		err := IsActionLegal(s, action)
		if err != nil {
			out <- []byte(fmt.Sprintf("%v\n", err))
		} else {
			break
		}
	}
	return action
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

func broadcast_to_all(out map[int]chan []byte, str string) {
	fmt.Print(str)
	for _, value := range out {
		value <- []byte(str)
	}
}

func getCriteriaString(s *GameState, veg_type VegType, id int) string {
	return s.strCriterias[int(veg_type)+id*VEGETABLE_TYPE_NUM]
}

func displayMarket(s *GameState, out map[int]chan []byte) {
	broadcast_to_all(out, fmt.Sprintf("---- MARKET ----\n"))
	for i, cardspot := range s.market {
		if cardspot.hasCard {
			card := cardspot.card
			broadcast_to_all(out, fmt.Sprintf("[%c] %v\n", i+'A', card.vegType))
		}
	}
	fmt.Println("piles")
	for i, pile := range s.piles {
		if len(pile) > 0 {
			top_card := pile[len(pile)-1]
			broadcast_to_all(out, fmt.Sprintf("[%d] %s\n", i, getCriteriaString(s, top_card.vegType, top_card.id)))
		} else {
			broadcast_to_all(out, "\n")
		}
	}
}

func drawFromTop(s *GameState, pile_index int) Card {
	assert(len(s.piles[pile_index]) > 0)
	c := s.piles[pile_index][len(s.piles[pile_index])-1]
	s.piles[pile_index] = s.piles[pile_index][0 : len(s.piles[pile_index])-1]
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

func displayActorCards(s *GameState, out chan []byte) {
	assert(s.activeActor < len(s.actorData))
	out <- []byte(fmt.Sprintf("---- Player %d ----\n", s.activeActor))

	out <- []byte(fmt.Sprintf("%d current score\n", calculateScore(s, s.activeActor)))
	out <- []byte("--------\n")
	for i, num := range s.actorData[s.activeActor].vegetableNum {
		out <- []byte(fmt.Sprintf("%d %v\n", num, VegType(i)))

	}

	out <- []byte(fmt.Sprintf("---- point cards ----\n"))

	for i, card := range s.actorData[s.activeActor].pointPile {
		out <- []byte(fmt.Sprintf("%d: %s\n", i, getCriteriaString(s, card.vegType, card.id)))
	}
}

func flipCardsFromPiles(s *GameState) {
	for y := range s.piles {
		for x := range 2 {
			market_pos := y + x*playPilesNum
			if !s.market[market_pos].hasCard {
				if len(s.piles[y]) == 0 {
					s.market[market_pos].card = drawFromTop(s, y)
					s.market[market_pos].hasCard = true

				} else {
					index := getMaxPileIndex(s)
					s.market[market_pos].card = drawFromBot(s, index)
					s.market[market_pos].hasCard = true
				}
			}
		}
	}
}

func isWithinAtoF(a byte) bool {
	return a >= 'A' && a <= 'F'
}

func pickCardToChangeToVeg(s *GameState, in chan []byte, out chan []byte) {
	for true {
		if len(s.actorData[s.activeActor].pointPile) == 0 {
			break
		}
		out <- []byte(fmt.Sprintf("pick 0-1 point card to flip to vegetable, type n to pick none example: 5\n"))
		input := <-in

		if input[0] == 'n' {
			break
		}

		index, err := strconv.Atoi(string(input))
		if err != nil {
			continue
		}

		if index >= 0 && index < len(s.actorData[s.activeActor].pointPile) {

			card := s.actorData[s.activeActor].pointPile[index]

			s.actorData[s.activeActor].vegetableNum[int(card.vegType)] += 1

			// shift slice
			for i := index; i < len(s.actorData[s.activeActor].pointPile)-1; i += 1 {
				s.actorData[s.activeActor].pointPile[i] = s.actorData[s.activeActor].pointPile[i+1]
			}
			// remove last element
			s.actorData[s.activeActor].pointPile = s.actorData[s.activeActor].pointPile[0 : len(s.actorData[s.activeActor].pointPile)-1]
			break
		}
	}
}

func calculateScore(s *GameState, actorId int) int {
	score := 0

	for _, point_card := range s.actorData[actorId].pointPile {
		_ = point_card
		var criteria Criteria

		switch criteria.criteria_type {
		case MOST:
			{
				is_most := true
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					}
					max := s.actorData[actorId].vegetableNum[0]
					max_id := 0
					for _, actorData := range s.actorData {
						if actorData.vegetableNum[j] > max {
							max = actorData.vegetableNum[j]
							max_id = j
						}
					}
					if max_id != actorId {
						is_most = false
						break
					}
				}
				if is_most {
					score += criteria.single_score
				}
			}
		case FEWEST:
			{
				is_fewest := true
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					}
					min := s.actorData[0].vegetableNum[0]
					min_id := 0
					for _, actorData := range s.actorData {
						if actorData.vegetableNum[j] < min {
							min = actorData.vegetableNum[j]
							min_id = j
						}
					}
					if min_id != actorId {
						is_fewest = false
						break
					}
				}
				if is_fewest {
					score += criteria.single_score
				}
			}
		case EVEN_ODD:
			{
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					}
					if s.actorData[actorId].vegetableNum[VegType(j)]%2 == 0 {
						score += criteria.even_score
					} else {
						score += criteria.odd_score
					}
				}
			}
		case PER:
			{
				for j, per_value := range criteria.per_scores {
					score += s.actorData[actorId].vegetableNum[j] * per_value
				}
			}
		case SUM:
			{
				min := math.MaxInt32
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					}
					non_repeated_value := s.actorData[actorId].vegetableNum[j] / count
					if non_repeated_value < min {
						min = non_repeated_value
					}
				}
				score += min * criteria.single_score
			}
		case MOST_TOTAL:
			{
				veg_count := 0
				for _, count := range s.actorData[actorId].vegetableNum {
					veg_count += count
				}

				is_most := true
				for _, actorData := range s.actorData {
					other_veg_count := 0
					for _, count := range actorData.vegetableNum {
						other_veg_count += count
					}
					if other_veg_count >= veg_count {
						is_most = false
						break
					}
				}
				if is_most {
					score += criteria.single_score
				}
			}
		case FEWEST_TOTAL:
			{
				veg_count := 0
				for _, count := range s.actorData[actorId].vegetableNum {
					veg_count += count
				}

				is_fewest := true
				for _, actorData := range s.actorData {
					other_veg_count := 0
					for _, count := range actorData.vegetableNum {
						other_veg_count += count
					}
					if other_veg_count <= veg_count {
						is_fewest = false
						break
					}
				}
				if is_fewest {
					score += criteria.single_score
				}
			}
		case PER_TYPE_GREATER_THAN_EQ:
			{
				for _, count := range s.actorData[actorId].vegetableNum {
					if count >= criteria.greater_than_eq_value {
						score += criteria.single_score
					}
				}
			}
		case PER_MISSING_TYPE:
			{
				for _, count := range s.actorData[actorId].vegetableNum {
					if count == 0 {
						score += criteria.single_score
					}
				}
			}
		case COMPLETE_SET:
			{
				min := s.actorData[actorId].vegetableNum[0]
				for _, count := range s.actorData[actorId].vegetableNum {
					if count < min {
						min = count
					}
				}
				score += criteria.single_score * min
			}
		default:
			{
				assert(false)
			}
		}

	}

	return score
}

// should be called before doAction
func BroadcastAction(s *GameState, action ActorAction, out map[int]chan []byte) {
	broadcast_to_all(out, "---- Action ----\n")
	switch action.kind {
	case INVALID:
		panic("unreachable")
	case PICK_VEG_FROM_MARKET:
		{
			for i := range action.amount {
				broadcast_to_all(out, fmt.Sprintf("Player %d drew %v from market\n", s.activeActor, s.market[action.ids[i]].card.vegType.String()))
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			for i := range action.amount {
				pile := s.piles[action.ids[i]]
				card := pile[len(pile)-1]
				criteria := getCriteriaString(s, card.vegType, card.id)
				broadcast_to_all(out, fmt.Sprintf("Player %d drew %v from market\n", s.activeActor, criteria))
			}
		}
	case PICK_TO_SWAP:
		{
			if action.amount == 0 {
				broadcast_to_all(out, fmt.Sprintf("Player %d did not swap any card\n", s.activeActor))
			} else {
				for i := range action.amount {
					card := s.actorData[s.activeActor].pointPile[action.ids[i]]
					criteria := getCriteriaString(s, card.vegType, card.id)
					broadcast_to_all(out, fmt.Sprintf("Player %d swapped %v to %v\n", s.activeActor, criteria, card.vegType))
				}
			}
		}
	}
}