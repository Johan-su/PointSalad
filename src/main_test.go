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


func TestPointCards(t *testing.T) {
	initJson()
}

