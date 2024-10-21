package pointsalad

import (
	"strings"
	"fmt"
)

type CardSpot struct {
	hasCard bool
	card    Card
}

type Market struct {
	// the amount of piles is the width
	piles [][]Card
	// the amount of cardSpots has to be a multiple of the amount of piles
	cardSpots []CardSpot
}

// createMarket initializes a new Market with card piles and card spots based on the provided deck, width, and height.
// The deck is evenly split into `playPilesNum` piles. Each pile gets an equal number of cards, and the remaining cards (if any) are asserted to be zero.
// The function also sets up a grid of card spots with the specified width and height.
//
// Parameters:
// - width: the number of columns for the grid of card spots.
// - height: the number of rows for the grid of card spots.
// - deck: a slice of `Card` that will be divided into piles.
//
// Returns:
// - A `Market` object with the following fields initialized:
//   - `piles`: A slice of piles, each containing an equal number of cards from the deck.
//   - `cardSpots`: A slice representing a grid of card spots with size `width * height`.
//

func createMarket(width int, height int, deck []Card) Market {
	m := Market{}

	pileSize := len(deck) / playPilesNum
	pileSizeRemainder := len(deck) % playPilesNum
	assert(pileSizeRemainder == 0)

	for i := range playPilesNum {
		m.piles = append(m.piles, []Card{})
		m.piles[i] = deck[i*pileSize : (i+1)*pileSize]
	}

	m.cardSpots = make([]CardSpot, width*height)
	assert(len(m.cardSpots)%len(m.piles) == 0)
	return m
}

// getMaxPileIndex returns the index of the pile with the largest number of elements in the Market's piles.
// If the Market has no piles (i.e., an empty slice), it returns -1.
//
// Parameters:
// - m: A pointer to the Market object that contains the piles (a slice of piles).
//
// Returns:
// - An integer representing the index of the pile with the maximum size.
//   If the Market is empty, it returns -1.

func getMaxPileIndex(m *Market) int {
	max := 0
	index := -1

	for i, p := range m.piles {
		if len(p) > max {
			max = len(p)
			index = i
		}
	}

	return index
}

// flipCardsFromPiles iterates through the market grid and places cards onto the market
// spots from the piles. It checks if the market position already has a card, and if not,
// it draws cards either from the specified pile or from the bottom of the pile with the
// most cards. If all piles are empty, the function returns early.
//
// Parameters:
//   m - A pointer to the Market object, which contains the piles, card spots, and other related data.
//
// This function ensures that the market grid is populated with cards from the piles, prioritizing
// the available piles and using the largest pile when necessary. If no cards are available to draw,
// the function terminates early.

func flipCardsFromPiles(m *Market) {
	for x := range getMarketWidth(m) {
		for y := range getMarketHeight(m) {
			market_pos := x + y*playPilesNum
			if !m.cardSpots[market_pos].hasCard {
				if len(m.piles[x]) > 0 {
					m.cardSpots[market_pos].card = drawFromTop(m, x)
					m.cardSpots[market_pos].hasCard = true

				} else {
					index := getMaxPileIndex(m)
					// all piles are empty
					if index == -1 {
						return
					}
					m.cardSpots[market_pos].card = drawFromBot(m, index)
					m.cardSpots[market_pos].hasCard = true
				}
			}
		}
	}
}

// getMarketWidth returns the width of the market, which is determined by the number
// of piles in the Market. It provides the number of columns of available piles.
//
// Parameters:
//   m - A pointer to the Market object.
//
// Returns:
//   An integer representing the number of piles (or columns) in the market.

func getMarketWidth(m *Market) int {
	return len(m.piles)
}

// getMarketHeight returns the height of the market
// It provides the number of rows in the market grid.
//
// Parameters:
//   m - A pointer to the Market object.
//
// Returns:
//   An integer representing the number of rows of available card spots in the market.

func getMarketHeight(m *Market) int {
	return len(m.cardSpots) / len(m.piles)
}

func drawFromTop(m *Market, pile_index int) Card {
	assert(len(m.piles[pile_index]) > 0)
	c := m.piles[pile_index][len(m.piles[pile_index])-1]
	m.piles[pile_index] = m.piles[pile_index][0 : len(m.piles[pile_index])-1]
	return c
}

func drawFromBot(m *Market, pile_index int) Card {
	assert(len(m.piles[pile_index]) > 0)
	c := m.piles[pile_index][0]
	m.piles[pile_index] = m.piles[pile_index][1:len(m.piles[pile_index])]
	return c
}

func hasCard(m *Market, id int) bool {
	return m.cardSpots[id].hasCard
}

func getCardFromMarket(m *Market, id int) Card {
	assert(hasCard(m, id))
	return m.cardSpots[id].card
}

func getMarketString(m *Market) string {
	builder := strings.Builder{}
	builder.WriteString("---- MARKET ----\n")
	for i := range m.cardSpots {
		if hasCard(m, i) {
			card := getCardFromMarket(m, i)
			builder.WriteString(fmt.Sprintf("[%c] %v\n", i+'A', card.vegType))
		}
	}
	builder.WriteString("piles:\n")
	for i, pile := range m.piles {
		if len(pile) > 0 {
			topCard := pile[len(pile)-1]
			builder.WriteString(fmt.Sprintf("[%d] %s\n", i, topCard.criteria.String()))
		} else {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}