package pointsalad

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"slices"
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
	marketColumns         = 2
	serverByteReceiveSize = 4
	serverByteSendSize    = 1024
)

type Card struct {
	criteria Criteria
	vegType  VegType
}

type ActorData struct {
	vegetableNum [vegetableTypeNum]int
	pointPile    []Card
}

// actors are players and bots

type GameState struct {
	// market
	market Market

	//
	actorData   []ActorData
	activeActor int
	playerNum   int
	botNum      int
}

// Init initializes the game state for a new game with the specified number of players and bots.
//
// This function sets up the initial game state by:
// 1. Verifying that the total number of players (human + bot) is between 2 and 6.
// 2. Loading the game configuration and card data from the "pointsaladManifest.json" file.
// 3. Creating a new game state based on the provided number of players and bots, and using the current time as a seed for randomization.
//
// Parameters:
//   - playerNum: The number of human players in the game.
//   - botNum: The number of bots in the game. The total number of players and bots must be between 2 and 6 (inclusive).
//
// Side effects:
//   - The function modifies the `GameState` object (`state`) to reflect the initialized game state with players, bots, and cards.
//
// Returns:
//   - None. In case of an error (e.g., invalid number of players/bots or failure to read the game manifest), the function will log the error and terminate the program.
//
// Example usage:
//   - To start a new game with 2 human players and 1 bot:
//     state.Init(2, 1)
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

// RunHost runs the main game loop for the host, managing the game flow for both players and bots.
//
// This function orchestrates the core gameplay loop for the host by doing the following:
// 1. **Market Actions**: Alternates between getting actions from either human players or bots. It provides game information to players (or gets automated actions from bots), and processes their market decisions (e.g., choosing vegetables or point cards).
// 2. **Action Execution**: After getting the market action from the active player, the action is broadcast to all players, and the state is updated accordingly.
// 3. **Swap Phase**: If the active player has point cards, it enters a swap phase where players/bots can choose to flip a point card into a vegetable card. The process is similar to the market action, where players can either make a decision or let the bot automatically choose.
// 4. **Hand Sharing**: After each action, the host broadcasts the current state of the active player's hand to all other players. This ensures that each player is aware of others' progress.
// 5. **Game End and Winner Announcement**: The game checks if a player has won, and if so, the host broadcasts the final scores and ends the game.
// 6. **Actor Switching**: After every turn, the host moves to the next active player, cycling through all players and bots, until a winner is found.
//
// Parameters:
//   - in: A map where the keys are actor IDs (player/bot), and the values are channels from which the host can receive input (commands) from the respective actors.
//   - out: A map where the keys are actor IDs, and the values are channels to which the host can send output (game state information) to the respective actors.
//
// Side effects:
//   - This function modifies the state of the game as actions are taken by players or bots, and broadcasts game state updates to all participants.
//   - The game ends when a player wins, and the final scores are broadcast to all players/bots.
//
// Returns:
//   - None. The game loop will continue until a winner is found or a player exits (e.g., by sending 'Q' to quit).
//
// Example usage:
//   - To start the host game loop with two human players and one bot:
//     state.RunHost(playerInputChannels, botInputChannels)
func (state *GameState) RunHost(in map[int]chan []byte, out map[int]chan []byte) {
	var err error
	for _, v := range in {
		assert(v != nil)
	}
	for _, v := range out {
		assert(v != nil)
	}
	for {
		flipCardsFromPiles(&state.market)
		is_bot := in[state.activeActor] == nil

		// get decisions from actor
		var market_action ActorAction
		if is_bot {
			market_action = getMarketActionFromBot(state)
		} else {
			s := getActorCardsString(state, state.activeActor) + getMarketString(state)
			out[state.activeActor] <- []byte(s)
			for {
				out[state.activeActor] <- []byte("pick 1 or 2 vegetables example: AB or\npick 1 point card example: 0\n")
				input := <-in[state.activeActor]
				if len(input) == 0 || (len(input) == 1 && input[0] == 'Q') {
					return
				}
				market_action, err = parseMarketActionFromPlayer(state, input)
				if err != nil {
					out[state.activeActor] <- []byte(fmt.Sprintf("%v\n", err))
				} else {
					break
				}
			}
		}
		broadcastToAll(out, getActionString(state, market_action))
		doAction(state, market_action)

		if len(state.actorData[state.activeActor].pointPile) > 0 {
			var swap_action ActorAction
			if is_bot {
				swap_action = getSwapActionFromBot(state)
			} else {
				out[state.activeActor] <- []byte(getActorCardsString(state, state.activeActor))
				for {
					out[state.activeActor] <- []byte(fmt.Sprintf("pick 0-1 point card to flip to vegetable, type n to pick none example: 5\n"))
					input := <-in[state.activeActor]
					if len(input) == 0 || (len(input) == 1 && input[0] == 'Q') {
						return
					}
					swap_action, err = parseSwapActionFromPlayer(state, input)
					if err != nil {
						out[state.activeActor] <- []byte(fmt.Sprintf("%v\n", err))
					} else {
						break
					}
				}
			}
			broadcastToAll(out, getActionString(state, swap_action))
			doAction(state, swap_action)
		}
		// show hand to all other players
		for k, o := range out {
			if k == state.activeActor {
				continue
			}
			o <- []byte(getActorCardsString(state, state.activeActor))
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

// RunPlayer starts the player game loop for human players, reading and writing data from/to the player's input and output channels.
//
// This function serves as an entry point for running a player in the game. It uses `runPlayerWithReader` to handle player interaction with the game through standard input and output channels.
//
// Parameters:
//   - in: A channel from which the function receives game data to present to the player.
//   - out: A channel to which the function sends player input back to the game (e.g., decisions or actions).
//
// Side effects:
//   - This function expects the player to provide inputs via standard input. Once the player inputs data, it sends the response back to the game through the `out` channel.
//
// Returns:
//   - None. The function loops indefinitely until the player quits (by sending a "quit" command).
func (_ *GameState) RunPlayer(in chan []byte, out chan []byte) {
	runPlayerWithReader(in, out, bufio.NewReader(os.Stdin))
}

// GetMaxHostDataSize returns the maximum size (in bytes) that the server (host) can receive from clients.
//
// This function is used to define the maximum allowed size of incoming data packets for the server, which helps prevent overloads or malicious data injections.
//
// Parameters:
//   - None
//
// Returns:
//   - An integer representing the maximum size of data packets the host can handle.
func (_ *GameState) GetMaxHostDataSize() int {
	return serverByteReceiveSize
}

// GetMaxPlayerDataSize returns the maximum size (in bytes) that the player can send to the server (host).
//
// This function is used to define the maximum size of outgoing data packets from a player to the host, helping to ensure that data sent by the player doesn't exceed acceptable limits.
//
// Parameters:
//   - None
//
// Returns:
//   - An integer representing the maximum size of data packets a player can send.
func (_ *GameState) GetMaxPlayerDataSize() int {
	return serverByteSendSize
}

// runPlayerWithReader handles player interaction with the game using a specified input reader (e.g., stdin).
// It reads input from the player, processes it, and sends back the response through the output channel.
//
// This function continuously listens for incoming game data (in the form of byte slices) and responds with player input. It uses a scanner to read lines of text input from the player. The function expects certain prompts to trigger player responses (such as "pick"), and it terminates when the player provides input that matches the quit condition (e.g., sending an empty byte slice).
//
// Parameters:
//   - in: A channel from which the function receives game data to present to the player (e.g., prompts or information).
//   - out: A channel to which the function sends player input back to the game (e.g., decisions or actions).
//   - r: An `io.Reader` used for reading player input (typically `os.Stdin` for human players).
//
// Side effects:
//   - The function expects the player to provide responses via standard input. Once the player inputs data, it is sent back to the game through the `out` channel.
//   - The function ends when the player gets a signal to quit
//
// Returns:
//   - None.
func runPlayerWithReader(in chan []byte, out chan []byte, r io.Reader) {
	assert(in != nil)
	assert(out != nil)
	assert(r != nil)

	scan := bufio.NewScanner(r)
	for {
		data := <-in
		if expectQuit(data) {
			return
		}
		fmt.Printf("%s", string(data))
		if expectResponse(data) {
			var str string
			{
				if !scan.Scan() {
					err := scan.Err()
					if err != nil {
						log.Fatalf("ERROR: %s\n", err)
					}
					return
				}
				s := scan.Text()
				// should work for linux/macos too
				s = strings.TrimSuffix(s, "\n")
				s = strings.TrimSuffix(s, "\r")
				str = s
			}
			out <- []byte(str)
		}
	}
}

func expectQuit(data []byte) bool {
	return len(data) == 0
}

func expectResponse(data []byte) bool {
	return strings.Contains(string(data), "pick")
}

func hasWon(state *GameState) bool {
	// winner if all piles are empty
	for i := range state.market.piles {
		if len(state.market.piles[i]) != 0 {
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
		s := fmt.Sprintf("assertion failed [%v]", c)
		panic(s)
	}
}

// createDeck generates a deck of cards based on the provided JSON card data and the number of cards per vegetable type.
// It shuffles the card IDs and creates cards using the criteria for each vegetable type. Each card will have an associated vegetable type and its criteria.
//
// Parameters:
//   - jsonCards: A pointer to a `JCards` structure containing the JSON data for the available cards.
//   - perVegetableNum: The number of cards to generate for each vegetable type.
//
// Returns:
//   - A slice of `Card` structures representing the deck of cards.
func createDeck(jsonCards *JCards, perVegetableNum int) []Card {
	var deck []Card
	var ids []int
	for id, _ := range jsonCards.Cards {
		ids = append(ids, id)
	}
	for i := range vegetableTypeNum {
		rand.Shuffle(len(ids), func(i int, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})

		for j := 0; j < perVegetableNum; j += 1 {
			criteria, err := parseCriteria(getJCriteria(jsonCards, VegType(i), ids[j]))
			if err != nil {
				log.Fatalf("ERROR: while creating deck: %v\n", err)
			}
			card := Card{
				criteria: criteria,
				vegType:  VegType(i),
			}
			deck = append(deck, card)

		}
	}
	return deck
}

// createGameState initializes a new game state with a shuffled deck and a random seed for actor turns.
// It creates the game market, assigns actors (players and bots), and sets up the initial conditions for the game based on the provided parameters.
//
// Parameters:
//   - jsonCards: A pointer to a `JCards` structure containing the JSON data for the cards.
//   - playerNum: The number of players in the game.
//   - botNum: The number of bots in the game.
//   - seed: A seed for the random number generator, used to ensure consistent randomization across game sessions.
//
// Returns:
//   - A `GameState` structure representing the initialized game state.
//   - An error if the number of players + bots is out of the expected range (between 2 and 6).
func createGameState(jsonCards *JCards, playerNum int, botNum int, seed int64) (GameState, error) {
	actorNum := playerNum + botNum
	if !(actorNum >= 2 && actorNum <= 6) {
		return GameState{}, fmt.Errorf("Number of players + bots have to be between 2-6")
	}
	rand.Seed(seed)

	s := GameState{}

	deck := createDeck(jsonCards, 3*actorNum)
	rand.Shuffle(len(deck), func(i int, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	s.market = createMarket(playPilesNum, marketColumns, deck)

	for range actorNum {
		s.actorData = append(s.actorData, ActorData{})
	}

	s.activeActor = rand.Intn(actorNum)
	s.playerNum = playerNum
	s.botNum = botNum
	return s, nil
}

// deepCloneGameState creates a deep copy of the provided game state, including all market piles, card spots, and actor data (vegetables and point piles).
// The cloned game state will be an exact copy of the original, allowing for parallel game simulations or backups.
//
// Parameters:
//   - s: The original `GameState` structure to be cloned.
//
// Returns:
//   - A new `GameState` structure that is a deep clone of the original game state.
func deepCloneGameState(s *GameState) GameState {
	new := GameState{}

	for i := range s.market.piles {
		new.market.piles = append(new.market.piles, []Card{})
		for j := range s.market.piles[i] {
			new.market.piles[i] = append(new.market.piles[i], s.market.piles[i][j])
		}
	}

	for i := range s.market.cardSpots {
		new.market.cardSpots = append(new.market.cardSpots, s.market.cardSpots[i])
	}

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

	return new
}

func broadcastToAll(out map[int]chan []byte, str string) {
	fmt.Print(str)
	for _, value := range out {
		value <- []byte(str)
	}
}

func getMarketString(s *GameState) string {
	builder := strings.Builder{}
	builder.WriteString("---- MARKET ----\n")
	for i := range s.market.cardSpots {
		if hasCard(&s.market, i) {
			card := getCardFromMarket(&s.market, i)
			builder.WriteString(fmt.Sprintf("[%c] %v\n", i+'A', card.vegType))
		}
	}
	builder.WriteString("piles:\n")
	for i, pile := range s.market.piles {
		if len(pile) > 0 {
			topCard := pile[len(pile)-1]
			builder.WriteString(fmt.Sprintf("[%d] %s\n", i, topCard.criteria.String()))
		} else {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func getActorCardsString(s *GameState, actorId int) string {
	assert(actorId < len(s.actorData))
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("---- Player %d ----\n", actorId))

	builder.WriteString(fmt.Sprintf("%d current score\n", calculateScore(s, actorId)))
	builder.WriteString("--------\n")

	for i, num := range s.actorData[actorId].vegetableNum {
		builder.WriteString(fmt.Sprintf("%d %v\n", num, VegType(i)))
	}

	builder.WriteString("---- point cards ----\n")

	for i, card := range s.actorData[actorId].pointPile {
		builder.WriteString(fmt.Sprintf("%d: %s\n", i, card.criteria.String()))
	}
	return builder.String()
}


func calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, pointCard := range s.actorData[actorId].pointPile {

		score += pointCard.criteria.calculateScore(s, actorId)
	}

	return score
}