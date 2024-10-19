package pointsalad

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type VegType int

const (
	PEPPER  VegType = iota
	LETTUCE VegType = iota

	CARROT  VegType = iota
	CABBAGE VegType = iota

	ONION  VegType = iota
	TOMATO VegType = iota

	vegetableTypeNum = iota
)

const (
	playPilesNum          = 3
	serverByteReceiveSize = 4
	serverByteSendSize    = 1024
)

type Card struct {
	id      int
	vegType VegType
}

type ActorData struct {
	vegetableNum [vegetableTypeNum]int
	pointPile    []Card
}

type CardSpot struct {
	hasCard bool
	card    Card
}

type JCriteria struct {
	PEPPER  string
	LETTUCE string
	CARROT  string
	CABBAGE string
	ONION   string
	TOMATO  string
}

type JCard struct {
	Id       int
	Criteria JCriteria
}

type JCards struct {
	Cards []JCard
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

	data, err := os.ReadFile("pointsaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonCards := JCards{}

	err = json.Unmarshal(data, &jsonCards)
	if err != nil {
		log.Fatal(err)
	}

	{
		seed := time.Now().Unix() 
		game_state, err := createGameState(&jsonCards, playerNum, botNum, seed)
		if err != nil {
			log.Fatalf("ERROR: Failed to create game state: %s\n", err)
			return
		}
		*state = game_state
	}
}

func (state *GameState) RunHost(in map[int]chan []byte, out map[int]chan []byte) {
	for true {
		flipCardsFromPiles(state)
		is_bot := in[state.activeActor] == nil

		// get decisions from actor
		var market_action ActorAction
		if is_bot {
			market_action = getMarketActionFromBot(state)
		} else {
			out[state.activeActor] <- []byte(getActorCardsString(state, state.activeActor))
			out[state.activeActor] <- []byte(getMarketString(state))
			market_action = getMarketActionFromPlayer(state, in[state.activeActor], out[state.activeActor])
		}
		broadcastToAll(out, getActionString(state, market_action))
		doAction(state, market_action)

		if len(state.actorData[state.activeActor].pointPile) > 0 {
			var swap_action ActorAction
			if is_bot {
				swap_action = getSwapActionFromBot(state)
			} else {
				out[state.activeActor] <- []byte(getActorCardsString(state, state.activeActor))
				swap_action = getSwapActionFromPlayer(state, in[state.activeActor], out[state.activeActor])
			}
			broadcastToAll(out, getActionString(state, swap_action))
			doAction(state, swap_action)
		}
		// show hand to other players
		for i, v := range out {
			if i != state.activeActor {
				v <- []byte(getActorCardsString(state, i))
			}
		}

		if hasWon(state) {
			broadcastToAll(out, getFinalScoresString(state))
			break
		}

		// next actor
		state.activeActor += 1
		state.activeActor %= state.playerNum + state.botNum
	}
}

func (_ *GameState) RunPlayer(in chan []byte, out chan []byte) {
	assert(in != nil)
	assert(out != nil)

	r := bufio.NewReader(os.Stdin)
	for {
		data := <-in
		if expectQuit(data) {
			return
		}
		fmt.Printf("%s", string(data))
		if expectResponse(data) {
			var str string
			{
				s, err := r.ReadString('\n')
				if err != nil {
					log.Fatalf("ERROR: %s\n", err)
				}
				// should work for linux/macos too
				s = strings.TrimSuffix(s, "\n")
				s = strings.TrimSuffix(s, "\r")
				str = s
			}
			out <- []byte(str)
		}
	}
}

func (_ *GameState) GetMaxHostDataSize() int {
	return serverByteReceiveSize
}

func (_ *GameState) GetMaxPlayerDataSize() int {
	return serverByteSendSize
}

func expectQuit(data []byte) bool {
	return len(data) == 0
}

func expectResponse(data []byte) bool {
	return strings.Contains(string(data), "pick")
}

func hasWon(state *GameState) bool {
	// winner if all piles are empty
	for i := range state.piles {
		if len(state.piles[i]) != 0 {
			return false
		}
	}
	return true
}

func getFinalScoresString(state *GameState) string {
	type Score struct {
		score   int
		actorId int
	}
	scores := []Score{}

	for i := range state.playerNum + state.botNum {
		scores = append(scores, Score{score: calculateScore(state, i), actorId: i})
	}

	slices.SortFunc(scores, func(a, b Score) int {
		return b.score - a.score
	})

	highScore := scores[0].score

	builder := strings.Builder{}

	builder.WriteString("---- Final scores ----\n")
	for _, s := range scores {
		builder.WriteString(fmt.Sprintf("Player %d with score %d", s.actorId, s.score))
		if s.score == highScore {
			builder.WriteString(" Winner\n")
		} else {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func assert(c bool) {
	if !c {
		s := fmt.Sprintf("assertion failed %v", c)
		panic(s)
	}
}

func createGameState(jsonCards *JCards, playerNum int, botNum int, seed int64) (GameState, error) {
	actorNum := playerNum + botNum
	assert(actorNum >= 2 && actorNum <= 6)

	s := GameState{}

	rand.Seed(seed)

	for _, card := range jsonCards.Cards {
		s.strCriterias = append(s.strCriterias, card.Criteria.PEPPER)
		s.strCriterias = append(s.strCriterias, card.Criteria.LETTUCE)
		s.strCriterias = append(s.strCriterias, card.Criteria.CARROT)
		s.strCriterias = append(s.strCriterias, card.Criteria.CABBAGE)
		s.strCriterias = append(s.strCriterias, card.Criteria.ONION)
		s.strCriterias = append(s.strCriterias, card.Criteria.TOMATO)
	}

	table, err := createCriteriaTable(jsonCards)
	if err != nil {
		return GameState{}, err
	}
	s.criteriaTable = table


	var ids []int
	for id, _ := range jsonCards.Cards {
		ids = append(ids, id)
	}

	perVegetableNum := actorNum * 3
	var deck []Card

	for i := range vegetableTypeNum {
		rand.Shuffle(len(ids), func(i int, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		for j := 0; j < perVegetableNum; j += 1 {
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

func deepCloneGameState(s *GameState) GameState {
	new := GameState{}

	for i := range s.strCriterias {
		new.strCriterias = append(new.strCriterias, s.strCriterias[i])
	}

	for i := range s.criteriaTable {
		new.criteriaTable = append(new.criteriaTable, s.criteriaTable[i])
	}

	for i := range s.piles {
		new.piles = append(new.piles, []Card{})
		for j := range s.piles[i] {
			new.piles[i] = append(new.piles[i], s.piles[i][j])
		}
	}
	new.market = s.market

	for i := range s.actorData {
		new.actorData = append(new.actorData, ActorData{})
		new.actorData[i].vegetableNum = s.actorData[i].vegetableNum
		for j := range s.actorData[i].pointPile {
			new.actorData[i].pointPile = append(new.actorData[i].pointPile, s.actorData[i].pointPile[j])
		}
	}

	new.activeActor = s.activeActor
	new.playerNum = s.playerNum
	new.botNum = s.botNum

	assert(fmt.Sprintf("%v", new) == fmt.Sprintf("%v", *s))

	return new
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
	var action ActorAction
	for true {
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

func getMarketActionFromPlayer(s *GameState, in chan []byte, out chan []byte) ActorAction {
	assert(in != nil)
	assert(out != nil)

	action := ActorAction{}
	for {
		out <- []byte("pick 1 or 2 vegetables example: AB or\npick 1 point card example: 0\n")
		input := <-in

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

func broadcastToAll(out map[int]chan []byte, str string) {
	fmt.Print(str)
	for _, value := range out {
		value <- []byte(str)
	}
}

func getCriteriaString(s *GameState, veg_type VegType, id int) string {
	return s.strCriterias[int(veg_type)+id*vegetableTypeNum]
}

func getMarketString(s *GameState) string {
	builder := strings.Builder{}
	builder.WriteString("---- MARKET ----\n")
	for i, cardspot := range s.market {
		if cardspot.hasCard {
			card := cardspot.card
			builder.WriteString(fmt.Sprintf("[%c] %v\n", i+'A', card.vegType))
		}
	}
	builder.WriteString("piles:\n")
	for i, pile := range s.piles {
		if len(pile) > 0 {
			top_card := pile[len(pile)-1]
			builder.WriteString(fmt.Sprintf("[%d] %s\n", i, getCriteriaString(s, top_card.vegType, top_card.id)))
		} else {
			builder.WriteString("\n")
		}
	}
	return builder.String()
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
	max := 0
	index := -1

	for i, p := range s.piles {
		if len(p) > max {
			max = len(p)
			index = i
		}
	}

	return index
}

func getActorCardsString(s *GameState, actorId int) string {
	assert(s.activeActor < len(s.actorData))
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("---- Player %d ----\n", s.activeActor))

	builder.WriteString(fmt.Sprintf("%d current score\n", calculateScore(s, s.activeActor)))
	builder.WriteString("--------\n")

	for i, num := range s.actorData[s.activeActor].vegetableNum {
		builder.WriteString(fmt.Sprintf("%d %v\n", num, VegType(i)))
	}

	builder.WriteString("---- point cards ----\n")

	for i, card := range s.actorData[s.activeActor].pointPile {
		builder.WriteString(fmt.Sprintf("%d: %s\n", i, getCriteriaString(s, card.vegType, card.id)))
	}
	return builder.String()
}

func flipCardsFromPiles(s *GameState) {
	for y := range s.piles {
		for x := range 2 {
			market_pos := y + x*playPilesNum
			if !s.market[market_pos].hasCard {
				if len(s.piles[y]) > 0 {
					s.market[market_pos].card = drawFromTop(s, y)
					s.market[market_pos].hasCard = true

				} else {
					index := getMaxPileIndex(s)
					// all piles are empty
					if index == -1 {
						return
					}
					s.market[market_pos].card = drawFromBot(s, index)
					s.market[market_pos].hasCard = true
				}
			}
		}
	}
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

func getCardFromStr(s *GameState, str string) (Card, error) {
	for i, c_str := range s.strCriterias {
		if str == c_str {
			return Card{id: i, vegType: VegType(i % vegetableTypeNum)}, nil
		}
	}
	return Card{}, fmt.Errorf("Failed to find card with criteria %s", str)
}

func calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, point_card := range s.actorData[actorId].pointPile {

		criteria := s.criteriaTable[point_card.id]

		score += criteria.calculateScore(s, actorId)
	}

	return score
}

// should be called before doAction
func getActionString(s *GameState, action ActorAction) string {
	assert(IsActionLegal(s, action) == nil)
	builder := strings.Builder{}

	builder.WriteString("---- Action ----\n")
	switch action.kind {
	case INVALID:
		panic("unreachable")
	case PICK_VEG_FROM_MARKET:
		{
			for i := range action.amount {
				builder.WriteString(fmt.Sprintf("Player %d drew %v from market\n", s.activeActor, s.market[action.ids[i]].card.vegType.String()))
			}
		}
	case PICK_POINT_FROM_MARKET:
		{
			for i := range action.amount {
				pile := s.piles[action.ids[i]]
				card := pile[len(pile)-1]
				criteria := getCriteriaString(s, card.vegType, card.id)
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
					criteria := getCriteriaString(s, card.vegType, card.id)
					builder.WriteString(fmt.Sprintf("Player %d swapped %v to %v\n", s.activeActor, criteria, card.vegType))
				}
			}
		}
	}
	return builder.String()
}
