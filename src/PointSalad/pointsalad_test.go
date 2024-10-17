package pointsalad

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

var inited bool = false
var jsonCards JCards

func initJson() {
	if inited {
		return
	}
	data, err := os.ReadFile("../../pointsaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &jsonCards)
	if err != nil {
		log.Fatal(err)
	}
}

// ---- Requirement 1 ----
func correctPlayerAmount(t *testing.T, expected bool, playerNum int, botNum int) {
	// s, err := createGameState(&jsonCards, playerNum, botNum, 0)
	// value := err != nil
	// if expected != value {
	// 	t.Errorf("Expected %v got %v with %v %v\n", expected, value, playerNum, botNum)
	// }
}

func TestPlayerAmount(t *testing.T) {
	initJson()
	test_table := []struct {
		expected   bool
		playerNum int
		botNum    int
	}{
		{false, -1, -1},
		// players
		{false, 0, 0},
		{false, 1, 0},
		{true, 2, 0},
		{true, 3, 0},
		{true, 4, 0},
		{true, 5, 0},
		{true, 6, 0},
		{false, 7, 0},
		{false, 0, 1},
		// bots
		{false, 0, 1},
		{true, 0, 2},
		{true, 0, 3},
		{true, 0, 4},
		{true, 0, 5},
		{true, 0, 6},
		{false, 0, 7},
		{false, 0, 8},
	}
	for _, test := range test_table {
		correctPlayerAmount(t, test.expected, test.playerNum, test.botNum)
	}
}

// ---- Requirement 2 ----
func CorrectParsing(t *testing.T, criteria string, expected Criteria) {
	c, err := parseCriteria(criteria)
	if err != nil {
		t.Errorf("Error parsing %s, %s", criteria, err)
		return
	}
	if c != expected {
		t.Errorf("%s %v not equal to %v", criteria, c, expected)
	}
}

func TestCriteriaParsing(t *testing.T) {
	initJson()

	test_table := []struct {
		criteria_str string
		expected     Criteria
	}{
		{"MOST LETTUCE = 10", Criteria{criteria_type: MOST, veg_count: [6]int{0, 1, 0, 0, 0, 0}, single_score: 10}},
		{"MOST PEPPER = 10", Criteria{criteria_type: MOST, veg_count: [6]int{1, 0, 0, 0, 0, 0}, single_score: 10}},
		{"MOST CABBAGE = 10", Criteria{criteria_type: MOST, veg_count: [6]int{0, 0, 0, 1, 0, 0}, single_score: 10}},
		{"MOST CARROT = 10", Criteria{criteria_type: MOST, veg_count: [6]int{0, 0, 1, 0, 0, 0}, single_score: 10}},
		{"MOST TOMATO = 10", Criteria{criteria_type: MOST, veg_count: [6]int{0, 0, 0, 0, 0, 1}, single_score: 10}},
		{"MOST ONION = 10", Criteria{criteria_type: MOST, veg_count: [6]int{0, 0, 0, 0, 1, 0}, single_score: 10}},
		{"FEWEST LETTUCE = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{0, 1, 0, 0, 0, 0}, single_score: 7}},
		{"FEWEST PEPPER = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{1, 0, 0, 0, 0, 0}, single_score: 7}},
		{"FEWEST CABBAGE = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{0, 0, 0, 1, 0, 0}, single_score: 7}},
		{"FEWEST CARROT = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{0, 0, 1, 0, 0, 0}, single_score: 7}},
		{"FEWEST TOMATO = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{0, 0, 0, 0, 0, 1}, single_score: 7}},
		{"FEWEST ONION = 7", Criteria{criteria_type: FEWEST, veg_count: [6]int{0, 0, 0, 0, 1, 0}, single_score: 7}},
		{"LETTUCE: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{0, 1, 0, 0, 0, 0}, even_score: 7, odd_score: 3}},
		{"PEPPER: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{1, 0, 0, 0, 0, 0}, even_score: 7, odd_score: 3}},
		{"CABBAGE: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{0, 0, 0, 1, 0, 0}, even_score: 7, odd_score: 3}},
		{"CARROT: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{0, 0, 1, 0, 0, 0}, even_score: 7, odd_score: 3}},
		{"TOMATO: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{0, 0, 0, 0, 0, 1}, even_score: 7, odd_score: 3}},
		{"ONION: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: [6]int{0, 0, 0, 0, 1, 0}, even_score: 7, odd_score: 3}},
		{"2 / LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, 2, 0, 0, 0, 0}}},
		{"2 / PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{2, 0, 0, 0, 0, 0}}},
		{"2 / CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 2, 0, 0}}},
		{"2 / CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 2, 0, 0, 0}}},
		{"2 / TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 0, 0, 2}}},
		{"2 / ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 0, 2, 0}}},
		{"LETTUCE + LETTUCE = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 2, 0, 0, 0, 0}, single_score: 5}},
		{"PEPPER + PEPPER = 5", Criteria{criteria_type: SUM, veg_count: [6]int{2, 0, 0, 0, 0, 0}, single_score: 5}},
		{"CABBAGE + CABBAGE = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 2, 0, 0}, single_score: 5}},
		{"CARROT + CARROT = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 2, 0, 0, 0}, single_score: 5}},
		{"TOMATO + TOMATO = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 0, 0, 2}, single_score: 5}},
		{"ONION + ONION = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 0, 2, 0}, single_score: 5}},
		{"CARROT + ONION = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 1, 0, 1, 0}, single_score: 5}},
		{"CABBAGE + ONION = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 1, 1, 0}, single_score: 5}},
		{"TOMATO + LETTUCE = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 0, 0, 0, 1}, single_score: 5}},
		{"LETTUCE + ONION = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 0, 0, 1, 0}, single_score: 5}},
		{"CABBAGE + LETTUCE = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 0, 1, 0, 0}, single_score: 5}},
		{"CARROT + LETTUCE = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 1, 0, 0, 0}, single_score: 5}},
		{"CABBAGE + TOMATO = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 1, 0, 1}, single_score: 5}},
		{"CARROT + TOMATO = 5", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 1, 0, 0, 1}, single_score: 5}},
		{"ONION + PEPPER = 5", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 0, 1, 0}, single_score: 5}},
		{"TOMATO + PEPPER = 5", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 0, 0, 1}, single_score: 5}},
		{"CARROT + PEPPER = 5", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 1, 0, 0, 0}, single_score: 5}},
		{"CABBAGE + PEPPER = 5", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 1, 0, 0}, single_score: 5}},
		{"1 / LETTUCE,  1 / ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 1, 0, 0, 1, 0}}},
		{"1 / PEPPER,  1 / TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{1, 0, 0, 0, 0, 1}}},
		{"1 / CABBAGE,  1 / LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, 1, 0, 1, 0, 0}}},
		{"1 / CARROT,  1 / PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{1, 0, 1, 0, 0, 0}}},
		{"1 / TOMATO,  1 / CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 1, 0, 0, 1}}},
		{"1 / ONION,  1 / CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 1, 1, 0}}},
		{"1 / LETTUCE,  1 / TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, 1, 0, 0, 0, 1}}},
		{"1 / PEPPER,  1 / ONION", Criteria{criteria_type: PER, per_scores: [6]int{1, 0, 0, 0, 1, 0}}},
		{"1 / CABBAGE,  1 / PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{1, 0, 0, 1, 0, 0}}},
		{"1 / CARROT,  1 / LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, 1, 1, 0, 0, 0}}},
		{"1 / TOMATO,  1 / CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 1, 0, 1}}},
		{"1 / ONION,  1 / CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 1, 0, 1, 0}}},
		{"3 / LETTUCE,  -2 / CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 3, -2, 0, 0, 0}}},
		{"3 / PEPPER,  -2 / CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{3, 0, 0, -2, 0, 0}}},
		{"3 / CABBAGE,  -2 / TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 3, 0, -2}}},
		{"3 / CARROT,  -2 / ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 3, 0, -2, 0}}},
		{"3 / TOMATO,  -2 / LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, -2, 0, 0, 0, 3}}},
		{"3 / ONION,  -2 / PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{-2, 0, 0, 0, 3, 0}}},
		{"LETTUCE + LETTUCE + LETTUCE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 3, 0, 0, 0, 0}, single_score: 8}},
		{"PEPPER + PEPPER + PEPPER = 8", Criteria{criteria_type: SUM, veg_count: [6]int{3, 0, 0, 0, 0, 0}, single_score: 8}},
		{"CABBAGE + CABBAGE + CABBAGE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 3, 0, 0}, single_score: 8}},
		{"CARROT + CARROT + CARROT = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 3, 0, 0, 0}, single_score: 8}},
		{"TOMATO + TOMATO + TOMATO = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 0, 0, 3}, single_score: 8}},
		{"ONION + ONION + ONION = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 0, 0, 3, 0}, single_score: 8}},
		{"PEPPER + LETTUCE + CABBAGE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 1, 0, 1, 0, 0}, single_score: 8}},
		{"LETTUCE + PEPPER + CARROT = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 1, 1, 0, 0, 0}, single_score: 8}},
		{"CARROT + CABBAGE + ONION = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 1, 1, 1, 0}, single_score: 8}},
		{"CABBAGE + CARROT + TOMATO = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 0, 1, 1, 0, 1}, single_score: 8}},
		{"ONION + TOMATO + PEPPER = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 0, 1, 1}, single_score: 8}},
		{"TOMATO + ONION + LETTUCE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 0, 0, 1, 1}, single_score: 8}},
		{"TOMATO + LETTUCE + CARROT = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 1, 0, 0, 1}, single_score: 8}},
		{"ONION + PEPPER + CABBAGE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 1, 1, 0}, single_score: 8}},
		{"PEPPER + CABBAGE + TOMATO = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 0, 1, 0, 1}, single_score: 8}},
		{"LETTUCE + CARROT + ONION = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 1, 0, 1, 0}, single_score: 8}},
		{"CABBAGE + TOMATO + LETTUCE = 8", Criteria{criteria_type: SUM, veg_count: [6]int{0, 1, 0, 1, 0, 1}, single_score: 8}},
		{"CARROT + ONION + PEPPER = 8", Criteria{criteria_type: SUM, veg_count: [6]int{1, 0, 1, 0, 1, 0}, single_score: 8}},
		{"2/LETTUCE,  1/ONION,  -2/PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{-2, 2, 0, 0, 1, 0}}},
		{"2/PEPPER,  1/TOMATO,  -2/LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{2, -2, 0, 0, 0, 1}}},
		{"2/CABBAGE,  1/LETTUCE,  -2/CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 1, -2, 2, 0, 0}}},
		{"2/CARROT,  1/PEPPER,  -2/CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{1, 0, 2, -2, 0, 0}}},
		{"2/TOMATO,  1/CARROT,  -2/ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 1, 0, -2, 2}}},
		{"2/ONION,  1/CABBAGE,  -2/TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, 1, 2, -2}}},
		{"2/LETTUCE,  2/CARROT,  -4/ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 2, 2, 0, -4, 0}}},
		{"2/PEPPER,  2/CABBAGE,  -4/TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{2, 0, 0, 2, 0, -4}}},
		{"2/CABBAGE,  2/TOMATO,  -4/LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, -4, 0, 2, 0, 2}}},
		{"2/CARROT,  2/ONION,  -4/PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{-4, 0, 2, 0, 2, 0}}},
		{"2/TOMATO,  2/LETTUCE,  -4/CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, 2, -4, 0, 0, 2}}},
		{"2/ONION,  2/PEPPER,  -4/CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{2, 0, 0, -4, 2, 0}}},
		{"3/LETTUCE,  -1/ONION,  -1/PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{-1, 3, 0, 0, -1, 0}}},
		{"3/PEPPER,  -1/TOMATO,  -1/LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{3, -1, 0, 0, 0, -1}}},
		{"3/CABBAGE,  -1/LETTUCE,  -1/CARROT", Criteria{criteria_type: PER, per_scores: [6]int{0, -1, -1, 3, 0, 0}}},
		{"3/CARROT,  -1/PEPPER,  -1/CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{-1, 0, 3, -1, 0, 0}}},
		{"3/TOMATO,  -1/CARROT,  -1/ONION", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, -1, 0, -1, 3}}},
		{"3/ONION,  -1/CABBAGE,  -1/TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, 0, 0, -1, 3, -1}}},
		{"4/LETTUCE,  -2/TOMATO,  -2/CABBAGE", Criteria{criteria_type: PER, per_scores: [6]int{0, 4, 0, -2, 0, -2}}},
		{"4/PEPPER,  -2/ONION,  -2/CARROT", Criteria{criteria_type: PER, per_scores: [6]int{4, 0, -2, 0, -2, 0}}},
		{"4/CABBAGE,  -2/PEPPER,  -2/ONION", Criteria{criteria_type: PER, per_scores: [6]int{-2, 0, 0, 4, -2, 0}}},
		{"4/CARROT,  -2/LETTUCE,  -2/TOMATO", Criteria{criteria_type: PER, per_scores: [6]int{0, -2, 4, 0, 0, -2}}},
		{"4/TOMATO,  -1/CABBAGE,  -2/PEPPER", Criteria{criteria_type: PER, per_scores: [6]int{-2, 0, 0, -1, 0, 4}}},
		{"4/ONION,  -2/CARROT,  -2/LETTUCE", Criteria{criteria_type: PER, per_scores: [6]int{0, -2, -2, 0, 4, 0}}},
		{"MOST TOTAL VEGETABLE = 10", Criteria{criteria_type: MOST_TOTAL, single_score: 10}},
		{"FEWEST TOTAL VEGETABLE = 7", Criteria{criteria_type: FEWEST_TOTAL, single_score: 7}},
		{"5 / VEGETABLE TYPE >=3", Criteria{criteria_type: PER_TYPE_GREATER_THAN_EQ, single_score: 5, greater_than_eq_value: 3}},
		{"5 / MISSING VEGETABLE TYPE", Criteria{criteria_type: PER_MISSING_TYPE, single_score: 5}},
		{"3 / VEGETABLE TYPE >=2", Criteria{criteria_type: PER_TYPE_GREATER_THAN_EQ, single_score: 3, greater_than_eq_value: 2}},
		{"COMPLETE SET = 12", Criteria{criteria_type: COMPLETE_SET, single_score: 12}},
	}

	for _, test := range test_table {
		CorrectParsing(t, test.criteria_str, test.expected)
	}
}

// ---- Requirement 3 ----

func CorrectVegetableAmount(t *testing.T, actorNum int, expectedNumOfVegetablePerType int) {
	s, err := createGameState(&jsonCards, 0, actorNum, 0)
	if err != nil {
		t.Fatalf("Failed to create GameState")
	}

	vegetableNums := [vegetableTypeNum]int{}

	for i1, pile := range s.piles {
		for j1, card := range pile {
			vegetableNums[int(card.vegType)] += 1
			for i2, other_pile := range s.piles {
				for j2, other_card := range other_pile {
					if i1 == i2 && j1 == j2 {
						continue
					}
					if card.id == other_card.id && card.vegType == other_card.vegType {
						t.Errorf("pile_id1=%d card_id1=%d pile_id2=%d card_id2=%d actorNum=%d %v vegetable with same id %d, ", i1, j1, i2, j2, actorNum, card.vegType, card.id)
					}
				}
			}
		}
	}

	for i, vegetable_num := range vegetableNums {
		if vegetable_num != expectedNumOfVegetablePerType {
			t.Errorf("Expected %d %v got %d", expectedNumOfVegetablePerType, VegType(i), vegetable_num)
		}
	}
}

func TestCorrectVegetables(t *testing.T) {
	initJson()
	test_table := []struct {
		actorNum                          int
		expectedNumOfVegetablePerType int
	}{
		{2, 6},
		{3, 9},
		{4, 12},
		{5, 15},
		{6, 18},
	}
	for _, v := range test_table {
		CorrectVegetableAmount(t, v.actorNum, v.expectedNumOfVegetablePerType)
	}
}

// ---- Requirement 10 ----

func TestSwitchingDrawPile(t *testing.T) {
	initJson()
	// s, err := createGameState(&jsonCards, 0, 2, 0)
	// if err != nil {
		
	// }
}

// ---- Requirement 13 ----

func CorrectCalculateScore(t *testing.T, expected_score int, vegetableNum [vegetableTypeNum]int, card_strs []string) {
	s, err := createGameState(&jsonCards, 0, 2, 0)
	if err != nil {
		t.Fatalf("Failed to create GameState")
	}
	s.actorData[0].vegetableNum = vegetableNum
	for _, str := range card_strs {
		card, err := getCardFromStr(&s, str)
		if err != nil {
			t.Fatalf("Failed to get card %s", str)
		}
		s.actorData[0].pointPile = append(s.actorData[0].pointPile, card)
	}
	score := calculateScore(&s, 0)
	if score != expected_score {
		t.Errorf("Expected %d got %d", expected_score, score)
	}

}

func TestCalculateScore(t *testing.T) {
	initJson()

	test_table := []struct{expected_score int; vegetableNum [vegetableTypeNum]int; card_strs []string} {
		{
			13,
			[vegetableTypeNum]int{6, 5, 4, 6, 2, 0},
			[]string {
				"4/ONION,  -2/CARROT,  -2/LETTUCE",
				"4/LETTUCE,  -2/TOMATO,  -2/CABBAGE",
				"ONION: EVEN=7, ODD=3",
				"4/CABBAGE,  -2/PEPPER,  -2/ONION",
			},
		},
		{
			24,
			[vegetableTypeNum]int{2, 15, 2, 2, 7, 2},
			[]string{"COMPLETE SET = 12"},
		},
	}

	for _, test := range test_table {
		CorrectCalculateScore(t, test.expected_score, test.vegetableNum, test.card_strs)
	}


}


// ---- End ----
