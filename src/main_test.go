package main
import (
	"testing"
	"os"
	"log"
	"encoding/json"
)

func CorrectVegetableAmount(t *testing.T, json_cards *JCards, actor_num int, expected_num_of_vegetable_per_type int) {
	s := createGameState(json_cards, 0, actor_num, 0)
	
	vegetable_nums := [VEGETABLE_TYPE_NUM]int{}
	
	for i1, pile := range s.piles {
		for j1, card := range pile {
			vegetable_nums[int(card.Vegetable_type)] += 1
			for i2, other_pile := range s.piles {
				for j2, other_card := range other_pile {
					if i1 == i2 && j1 == j2 {
						continue
					}
					if card.Id == other_card.Id && card.Vegetable_type == other_card.Vegetable_type {
						t.Errorf("pile_id1=%d card_id1=%d pile_id2=%d card_id2=%d actor_num=%d %v vegetable with same id %d, ", i1, j1, i2, j2, actor_num, card.Vegetable_type, card.Id)
					}
				}
			}
		}
	}
	
	for i, vegetable_num := range vegetable_nums {
		if vegetable_num != expected_num_of_vegetable_per_type {
			t.Errorf("Expected %d %v got %d", expected_num_of_vegetable_per_type, CardType(i), vegetable_num)
		}
	}
}

inited := false
var json_cards JCards

func initJson() {
	if (inited) {
		return
	}
	data, err := os.ReadFile("../PointSaladManifest.json")
	if err != nil {
		log.Fatal(err)
	}
	
	
	err = json.Unmarshal(data, &json_cards)
	if err != nil {
		log.Fatal(err)
	}
}


// req3
func TestCorrectVegetables(t *testing.T) {
	initJson()
	test_table := []struct{actor_num int, expected_num_of_vegetable_per_type int} {
		{2, 6},
		{3, 9},
		{4, 12},
		{5, 15},
		{6, 18},
	}
	for _, v := range test_table {
		CorrectVegetableAmount(t, &json_cards, v.actor_num, v.expected_num_of_vegetable_per_type)
	}
}

func TestCriteriaParsing(t *testing.T) {
	initJson()

	test_table := []struct {s string, c Criteria} {
		{"MOST LETTUCE = 10", Criteria{criteria_type: MOST, veg_count: {0, 1, 0, 0, 0, 0}, single_score: 10}}
		{"MOST PEPPER = 10", Criteria{criteria_type: MOST, veg_count: {1, 0, 0, 0, 0, 0}, single_score: 10}}
		{"MOST CABBAGE = 10", Criteria{criteria_type: MOST, veg_count: {0, 0, 0, 1, 0, 0}, single_score: 10}}
		{"MOST CARROT = 10", Criteria{criteria_type: MOST, veg_count: {0, 0, 1, 0, 0, 0}, single_score: 10}}
		{"MOST TOMATO = 10", Criteria{criteria_type: MOST, veg_count: {0, 0, 0, 0, 0, 1}, single_score: 10}}
		{"MOST ONION = 10", Criteria{criteria_type: MOST, veg_count: {0, 0, 0, 0, 1, 0}, single_score: 10}}
		{"FEWEST LETTUCE = 7", Criteria{criteria_type: FEWEST, veg_count: {0, 1, 0, 0, 0, 0}, single_score: 7}}
		{"FEWEST PEPPER = 7", Criteria{criteria_type: FEWEST, veg_count: {1, 0, 0, 0, 0, 0}, single_score: 7}}
		{"FEWEST CABBAGE = 7", Criteria{criteria_type: FEWEST, veg_count: {0, 0, 0, 1, 0, 0}, single_score: 7}}
		{"FEWEST CARROT = 7", Criteria{criteria_type: FEWEST, veg_count: {0, 0, 1, 0, 0, 0}, single_score: 7}}
		{"FEWEST TOMATO = 7", Criteria{criteria_type: FEWEST, veg_count: {0, 0, 0, 0, 0, 1}, single_score: 7}}
		{"FEWEST ONION = 7", Criteria{criteria_type: FEWEST, veg_count: {0, 0, 0, 0, 1, 0}, single_score: 7}}
		{"LETTUCE: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {0, 1, 0, 0, 0, 0}, even_score: 7, odd_score: 3}}
		{"PEPPER: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {1, 0, 0, 0, 0, 0}, even_score: 7, odd_score: 3}}
		{"CABBAGE: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {0, 0, 0, 1, 0, 0}, even_score: 7, odd_score: 3}}
		{"CARROT: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {0, 0, 1, 0, 0, 0}, even_score: 7, odd_score: 3}}
		{"TOMATO: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {0, 0, 0, 0, 0, 1}, even_score: 7, odd_score: 3}}
		{"ONION: EVEN=7, ODD=3", Criteria{criteria_type: EVEN_ODD, veg_count: {0, 0, 0, 0, 1, 0}, even_score: 7, odd_score: 3}}
		{"2 / LETTUCE", }
		{"2 / PEPPER", }
		{"2 / CABBAGE", }
		{"2 / CARROT", }
		{"2 / TOMATO", }
		{"2 / ONION", }
		{"LETTUCE + LETTUCE = 5", }
		{"PEPPER + PEPPER = 5", }
		{"CABBAGE + CABBAGE = 5", }
		{"CARROT + CARROT = 5", }
		{"TOMATO + TOMATO = 5", }
		{"ONION + ONION = 5", }
		{"CARROT + ONION = 5", }
		{"CABBAGE + ONION = 5", }
		{"TOMATO + LETTUCE = 5", }
		{"LETTUCE + ONION = 5", }
		{"CABBAGE + LETTUCE = 5", }
		{"CARROT + LETTUCE = 5", }
		{"CABBAGE + TOMATO = 5", }
		{"CARROT + TOMATO = 5", }
		{"ONION + PEPPER = 5", }
		{"TOMATO + PEPPER = 5", }
		{"CARROT + PEPPER = 5", }
		{"CABBAGE + PEPPER = 5", }
		{"1 / LETTUCE,  1 / ONION", }
		{"1 / PEPPER,  1 / TOMATO", }
		{"1 / CABBAGE,  1 / LETTUCE", }
		{"1 / CARROT,  1 / PEPPER", }
		{"1 / TOMATO,  1 / CARROT", }
		{"1 / ONION,  1 / CABBAGE", }
		{"1 / LETTUCE,  1 / TOMATO", }
		{"1 / PEPPER,  1 / ONION", }
		{"1 / CABBAGE,  1 / PEPPER", }
		{"1 / CARROT,  1 / LETTUCE", }
		{"1 / TOMATO,  1 / CABBAGE", }
		{"1 / ONION,  1 / CARROT", }
		{"3 / LETTUCE,  -2 / CARROT", }
		{"3 / PEPPER,  -2 / CABBAGE", }
		{"3 / CABBAGE,  -2 / TOMATO", }
		{"3 / CARROT,  -2 / ONION", }
		{"3 / TOMATO,  -2 / LETTUCE", }
		{"3 / ONION,  -2 / PEPPER", }
		{"LETTUCE + LETTUCE + LETTUCE = 8", }
		{"PEPPER + PEPPER + PEPPER = 8", }
		{"CABBAGE + CABBAGE + CABBAGE = 8", }
		{"CARROT + CARROT + CARROT = 8", }
		{"TOMATO + TOMATO + TOMATO = 8", }
		{"ONION + ONION + ONION = 8", }
		{"PEPPER + LETTUCE + CABBAGE = 8", }
		{"LETTUCE + PEPPER + CARROT = 8", }
		{"CARROT + CABBAGE + ONION = 8", }
		{"CABBAGE + CARROT + TOMATO = 8", }
		{"ONION + TOMATO + PEPPER = 8", }
		{"TOMATO + ONION + LETTUCE = 8", }
		{"TOMATO + LETTUCE + CARROT = 8", }
		{"ONION + PEPPER + CABBAGE = 8", }
		{"PEPPER + CABBAGE + TOMATO = 8", }
		{"LETTUCE + CARROT + ONION = 8", }
		{"CABBAGE + TOMATO + LETTUCE = 8", }
		{"CARROT + ONION + PEPPER = 8", }
		{"2/LETTUCE,  1/ONION,  -2/PEPPER", }
		{"2/PEPPER,  1/TOMATO,  -2/LETTUCE", }
		{"2/CABBAGE,  1/LETTUCE,  -2/CARROT", }
		{"2/CARROT,  1/PEPPER,  -2/CABBAGE", }
		{"2/TOMATO,  1/CARROT,  -2/ONION", }
		{"2/ONION,  1/CABBAGE,  -2/TOMATO", }
		{"2/LETTUCE,  2/CARROT,  -4/ONION", }
		{"2/PEPPER,  2/CABBAGE,  -4/TOMATO", }
		{"2/CABBAGE,  2/TOMATO,  -4/LETTUCE", }
		{"2/CARROT,  2/ONION,  -4/PEPPER", }
		{"2/TOMATO,  2/LETTUCE,  -4/CARROT", }
		{"2/ONION,  2/PEPPER,  -4/CABBAGE", }
		{"3/LETTUCE,  -1/ONION,  -1/PEPPER", }
		{"3/PEPPER,  -1/TOMATO,  -1/LETTUCE", }
		{"3/CABBAGE,  -1/LETTUCE,  -1/CARROT", }
		{"3/CARROT,  -1/PEPPER,  -1/CABBAGE", }
		{"3/TOMATO,  -1/CARROT,  -1/ONION", }
		{"3/ONION,  -1/CABBAGE,  -1/TOMATO", }
		{"4/LETTUCE,  -2/TOMATO,  -2/CABBAGE", }
		{"4/PEPPER,  -2/ONION,  -2/CARROT", }
		{"4/CABBAGE,  -2/PEPPER,  -2/ONION", }
		{"4/CARROT,  -2/LETTUCE,  -2/TOMATO", }
		{"4/TOMATO,  -1/CABBAGE,  -2/PEPPER", }
		{"4/ONION,  -2/CARROT,  -2/LETTUCE", }
		{"MOST TOTAL VEGETABLE = 10", }
		{"FEWEST TOTAL VEGETABLE = 7", }
		{"5 / VEGETABLE TYPE >=3", }
		{"5 / MISSING VEGETABLE TYPE", }
		{"3 / VEGETABLE TYPE >=2", }
		{"COMPLETE SET = 12" }
	}
}

