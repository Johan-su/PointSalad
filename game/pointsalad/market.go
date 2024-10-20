package pointsalad


type Market struct {
	// the amount of piles is the width
	piles  [][]Card
	// the amount of cardSpots has to be a multiple of the amount of piles
	cardSpots []CardSpot
}

func createMarket(width int, height int, deck []Card) Market {
	m := Market{}
	
	pileSize := len(deck) / playPilesNum
	pileSizeRemainder := len(deck) % playPilesNum
	assert(pileSizeRemainder == 0)

	for i := range playPilesNum {
		m.piles = append(m.piles, []Card{})
		m.piles[i] = deck[i*pileSize : (i+1)*pileSize]
	}

	m.cardSpots = make([]CardSpot, width * height)
	assert(len(m.cardSpots) % len(m.piles) == 0)
	return m
}

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

func flipCardsFromPiles(m *Market) {
	for y := range m.piles {
		for x := range 2 {
			market_pos := y + x*playPilesNum
			if !m.cardSpots[market_pos].hasCard {
				if len(m.piles[y]) > 0 {
					m.cardSpots[market_pos].card = drawFromTop(m, y)
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

func getMarketWidth(m *Market) int {
	return len(m.piles)
}

func getMarketHeight(m *Market) int {
	return len(m.cardSpots) / len(m.piles)
}