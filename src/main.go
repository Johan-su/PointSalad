package main

import (
	"bufio"
	"fmt"
	"os"
	"log"
	"strconv"
	"math"
	"math/rand"
	"net"
	"flag"
)

type VegType int
const (
	PEPPER VegType = iota
	LETTUCE VegType = iota

	CARROT VegType = iota
	CABBAGE VegType = iota

	ONION VegType = iota
	TOMATO VegType = iota

	VEGETABLE_TYPE_NUM = iota
)

const (
	PLAY_PILES_NUM = 3
	SERVER_BYTE_RECEIVE_SIZE = 8
	SERVER_BYTE_SEND_SIZE = 512
)


func assert(c bool) {
	if !c {
		s := fmt.Sprintf("assertion failed %v", c)
		panic(s)
	}
}

func assertM(c bool, m string) {
	if !c {
		s := fmt.Sprintf("assertion failed %v %s", c, m)
		panic(s)
	}
}

func todo() {
	s := fmt.Sprintf("TODO")
	panic(s)
}


type JCriteria struct {
	PEPPER string
	LETTUCE string
	CARROT string
	CABBAGE string
	ONION string
	TOMATO string
}

type JCard struct {
	Id int
	Criteria JCriteria
} 

type JCards struct {
	Cards []JCard
}

type CriteriaType int
const (
	MOST CriteriaType = iota
	FEWEST CriteriaType = iota
	EVEN_ODD CriteriaType = iota
	PER CriteriaType = iota
	SUM CriteriaType = iota
	MOST_TOTAL CriteriaType = iota
	FEWEST_TOTAL CriteriaType = iota
	PER_TYPE_GREATER_THAN_EQ CriteriaType = iota
	PER_MISSING_TYPE CriteriaType = iota
	COMPLETE_SET CriteriaType = iota
	CRITERIA_TYPE_NUM = iota
)

type Criteria struct {
	criteria_type CriteriaType
	//
	veg_count [VEGETABLE_TYPE_NUM]int
	// used for single score rules
	single_score int
	greater_than_eq_value int
	// used for even odd rules
	even_score int
	odd_score int
	// used for per rules
	per_scores [VEGETABLE_TYPE_NUM]int
}

type Card struct {
	Id int
	Vegetable_type VegType
}

type ActorData struct {
	vegetable_num [VEGETABLE_TYPE_NUM]int
	point_pile []Card
}

type CardSpot struct {
	has_card bool 
	card Card
}


type TokenType int
const (
	IDENTIFIER TokenType = iota
	EQUAL TokenType = iota
	NUMBER TokenType = iota
	COLON TokenType = iota
	COMMA TokenType = iota
	SLASH TokenType = iota
	PLUS TokenType = iota
	MINUS TokenType = iota
	GREATER TokenType = iota
)
type Token struct {
	token_type TokenType
	s string
}

type Lexer struct {
	tokens []Token
	index int
}

func get_token(lex *Lexer) Token {
	assert(lex.index < len(lex.tokens))
	return lex.tokens[lex.index]
}

func next_token(lex *Lexer) Token {
	lex.index += 1
	assert(lex.index < len(lex.tokens))
	return lex.tokens[lex.index]
}

func expect_token_type(lex *Lexer, token_type TokenType) error {
	token := get_token(lex)
	if token.token_type != token_type {
		return fmt.Errorf("Expected %v got %v\n", token_type, token.token_type) 
	}	
	return nil
}

func expect_next_token_type(lex *Lexer, token_type TokenType) error {
	token := next_token(lex)
	if token.token_type != token_type {
		return fmt.Errorf("Expected %v got %v\n", token_type, token.token_type) 
	}	
	return nil
}

func look_token(lex *Lexer, pos int) (Token, bool) {
	if  lex.index + pos >= len(lex.tokens) {
		return Token{}, false
	}
	return lex.tokens[lex.index + pos], true
}

func parseNumber(lex *Lexer, minus_or_num Token) (int, error) {
	var err error
	num := 0
	if minus_or_num.token_type == MINUS {
		t := next_token(lex)
		if t.token_type != NUMBER {
			todo()
		}

		num, err = strconv.Atoi(t.s)
		if err != nil {
			return 0, err
		}
		num *= -1

	} else if minus_or_num.token_type == NUMBER {
		num, err = strconv.Atoi(minus_or_num.s)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, fmt.Errorf("Expected number or minus got %v\n", minus_or_num)
	}
	return num, nil
}


func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isVegetable(s string) bool {

	for i := range VEGETABLE_TYPE_NUM {
		if VegType(i).String() == s {
			return true
		} 
	}

	return false
}

func getVegetableType(s string) VegType {
	for i := range VEGETABLE_TYPE_NUM {
		if VegType(i).String() == s {
			return VegType(i)
		} 
	}
	panic("unreachable")
	return -1
}

func parseCriteria(s string) (Criteria, error) {

	lex := Lexer{}


	// tokenize the criterias
	for i := 0; i < len(s);{
		if s[i] == ' ' {
			i += 1
			continue
		} else if isAlpha(s[i]) {
			start := i
			end := i
			for i < len(s) && isAlpha(s[i]) {
				i += 1
				end += 1
			}
			

			token := Token{IDENTIFIER, s[start:end]}
			lex.tokens = append(lex.tokens, token)
		} else if isDigit(s[i])  {
			start := i
			end := i
			for i < len(s) && isDigit(s[i]) {
				i += 1
				end += 1
			}

			token := Token{NUMBER, s[start:end]}
			lex.tokens = append(lex.tokens, token)
		} else if s[i] == '='  {
			lex.tokens = append(lex.tokens, Token{EQUAL, s[i:i + 1]})
			i += 1
		} else if s[i] == ':'  {
			lex.tokens = append(lex.tokens, Token{COLON, s[i:i + 1]})
			i += 1
		} else if s[i] == ','  {
			lex.tokens = append(lex.tokens, Token{COMMA, s[i:i + 1]})
			i += 1
		} else if s[i] == '/'  {
			lex.tokens = append(lex.tokens, Token{SLASH, s[i:i + 1]})
			i += 1
		} else if s[i] == '+'  {
			lex.tokens = append(lex.tokens, Token{PLUS, s[i:i + 1]})
			i += 1
		} else if s[i] == '-'  {
			lex.tokens = append(lex.tokens, Token{MINUS, s[i:i + 1]})
			i += 1
		} else if s[i] == '>' {
			lex.tokens = append(lex.tokens, Token{GREATER, s[i:i + 1]})
			i += 1
		} else {
			return Criteria{}, fmt.Errorf("Unknown character: %c", s[i])
		}
	}

	criteria := Criteria{}

	for lex.index = 0; lex.index < len(lex.tokens); lex.index += 1 {
		first := get_token(&lex) 
		if first.token_type == IDENTIFIER {
			if first.s == "MOST" || first.s == "FEWEST" {
				t := next_token(&lex) 
				if t.s == "TOTAL" {
					if next_token(&lex).s != "VEGETABLE" {
						todo()
					}
					if next_token(&lex).token_type != EQUAL {
						todo()
					}

					num, err := parseNumber(&lex, next_token(&lex))
					if err != nil {
						return criteria, err
					}
					if first.s == "MOST" {
						criteria.criteria_type = MOST_TOTAL
					} else if first.s == "FEWEST" {
						criteria.criteria_type = FEWEST_TOTAL
					} else {
						panic("unreachable")
					}
					criteria.single_score = num

					return criteria, nil

				} else if isVegetable(t.s) {
					v := getVegetableType(t.s)
					err := expect_next_token_type(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err 
					}
					num, err := parseNumber(&lex, next_token(&lex))
					if err != nil {
						return Criteria{}, err
					}

					if first.s == "MOST" {
						criteria.criteria_type = MOST
					} else if first.s == "FEWEST" {
						criteria.criteria_type = FEWEST
					} else {
						panic("unreachable")
					}
					criteria.veg_count[int(v)] += 1
					criteria.single_score = num

					return criteria, nil

				} else {
					todo()
				}
			} else if first.s == "COMPLETE" {
				if next_token(&lex).s != "SET" {
					todo()
				}
				if next_token(&lex).token_type != EQUAL {
					todo()
				}
				num, err := parseNumber(&lex, next_token(&lex))
				if err != nil {
					return criteria, err
				}

				criteria.criteria_type = COMPLETE_SET
				criteria.single_score = num

				return criteria, nil

			} else if isVegetable(first.s) {
				v := getVegetableType(first.s)
				t := next_token(&lex)
				if t.token_type == COLON {
					if next_token(&lex).s != "EVEN" {
						todo()
					}
					if next_token(&lex).token_type != EQUAL {
						todo()
					}
					even, err := parseNumber(&lex, next_token(&lex))
					if err != nil {
						return criteria, err
					}
					if next_token(&lex).token_type != COMMA {
						todo()
					}
					if next_token(&lex).s != "ODD" {
						todo()
					}
					if next_token(&lex).token_type != EQUAL {
						todo()
					}
					odd, err := parseNumber(&lex, next_token(&lex))
					if err != nil {
						return criteria, err
					}

					criteria.criteria_type = EVEN_ODD
					criteria.veg_count[int(v)] += 1
					criteria.even_score = even
					criteria.odd_score = odd

					return criteria, nil

				} else if t.token_type == PLUS {
					criteria.veg_count[int(v)] += 1
					for true {
						t := next_token(&lex)
						if !isVegetable(t.s) {
							todo()
						}
						v := getVegetableType(t.s)
						criteria.veg_count[int(v)] += 1
						if lex.index >= len(lex.tokens) - 1 || next_token(&lex).token_type != PLUS {
							break
						}
					}
					err := expect_token_type(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err
					}
					num, err := parseNumber(&lex, next_token(&lex))
					if err != nil {
						return Criteria{}, err
					}

					criteria.criteria_type = SUM
					criteria.single_score = num

					return criteria, nil

				} else {
					todo()
				}
			} else {
				fmt.Printf("len(%d) %s\n", len(first.s), first.s)
				todo()
			}
		} else if first.token_type == NUMBER {
			num, err := parseNumber(&lex, first)
			if err != nil {
				return Criteria{}, err
			}
			err = expect_next_token_type(&lex, SLASH)
			if err != nil {
				return Criteria{}, err
			}

			t := next_token(&lex)
			if t.s == "VEGETABLE" {
				if next_token(&lex).s != "TYPE" {
					todo()
				}
				if next_token(&lex).token_type != GREATER {
					todo()
				}
				if next_token(&lex).token_type != EQUAL {
					todo()
				}
				num2, err := parseNumber(&lex, next_token(&lex))
				if err != nil {
					return Criteria{}, err
				}

				criteria.criteria_type = PER_TYPE_GREATER_THAN_EQ
				criteria.single_score = num
				criteria.greater_than_eq_value = num2

				return criteria, nil

			} else if t.s == "MISSING" {
				if next_token(&lex).s != "VEGETABLE" {
					todo()
				}
				if next_token(&lex).s != "TYPE" {
					todo()
				}

				criteria.criteria_type = PER_MISSING_TYPE
				criteria.single_score = num

				return criteria, nil
			} else {
				for true {
					if !isVegetable(t.s) {
						todo()
					}
					v := getVegetableType(t.s)
					criteria.per_scores[int(v)] = num

					if lex.index >= len(lex.tokens) - 1 || next_token(&lex).token_type != COMMA {
						break
					}
					num, err = parseNumber(&lex, next_token(&lex))
					if err != nil {
						return Criteria{}, err
					}
					err = expect_next_token_type(&lex, SLASH)
					if err != nil {
						return Criteria{}, err
					}
					t = next_token(&lex)
				}
	
				criteria.criteria_type = PER
	
				return criteria, nil
			} 
		} else {
			return Criteria{}, fmt.Errorf("Expected Identifier or number as first token")
		}
	}
	todo()
	return Criteria{}, nil 
}

func createCriteriaTable(json_cards *JCards) ([]Criteria, error) {

	criteria_table := []Criteria{}

	for _, jcard := range json_cards.Cards {
		criteria_PEPPER, err := parseCriteria(jcard.Criteria.PEPPER)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_PEPPER)

		criteria_LETTUCE, err := parseCriteria(jcard.Criteria.LETTUCE)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_LETTUCE)

		criteria_CARROT, err := parseCriteria(jcard.Criteria.CARROT)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_CARROT)
		
		criteria_CABBAGE, err := parseCriteria(jcard.Criteria.CABBAGE)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_CABBAGE)
		
		criteria_ONION, err := parseCriteria(jcard.Criteria.ONION)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_ONION)
		
		criteria_TOMATO, err := parseCriteria(jcard.Criteria.TOMATO)
		if err != nil {
			return criteria_table, err
		}
		criteria_table = append(criteria_table, criteria_TOMATO)
	}
	return criteria_table, nil
}

func createPointSalad(json_cards *JCards, player_num int, bot_num int, seed int64) (PointSalad, error) {
	actor_num := player_num + bot_num
	assert(actor_num >= 2 && actor_num <= 6)

	log.Printf("seed = %d\n", seed)
	rand.Seed(seed)

	
	var ids []int
	for id, _ := range json_cards.Cards {
		ids = append(ids, id)
	}

	per_vegetable_num := actor_num * 3
	var deck []Card
	
	for i := range VEGETABLE_TYPE_NUM {
		rand.Shuffle(len(ids), func(i int, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})
		
		for j := 0; j < per_vegetable_num; j += 1 {
			card := Card{
				Id: ids[j], 
				Vegetable_type: VegType(i), 
			}
			deck = append(deck, card)
			
		}
	}

	rand.Shuffle(len(deck), func(i int, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	
	s := PointSalad{}

	for _, card := range json_cards.Cards {
		s.str_criterias = append(s.str_criterias, card.Criteria.PEPPER)
		s.str_criterias = append(s.str_criterias, card.Criteria.LETTUCE)
		s.str_criterias = append(s.str_criterias, card.Criteria.CARROT)
		s.str_criterias = append(s.str_criterias, card.Criteria.CABBAGE)
		s.str_criterias = append(s.str_criterias, card.Criteria.ONION)
		s.str_criterias = append(s.str_criterias, card.Criteria.TOMATO)
	}

	table, err := createCriteriaTable(json_cards)
	if err != nil {
		return PointSalad{}, err
	}
	s.criteria_table = table

	pile_size := len(deck) / PLAY_PILES_NUM
	pile_size_remainder := len(deck) % PLAY_PILES_NUM
	assert(pile_size_remainder == 0)

	for i := range PLAY_PILES_NUM {
		s.piles = append(s.piles, []Card{})
		s.piles[i] = deck[i * pile_size:(i + 1) * pile_size]
	}


	for range actor_num {
		s.actor_data = append(s.actor_data, ActorData{})
	}
	
	
	s.active_actor = rand.Intn(actor_num)
	s.player_num = player_num
	s.bot_num = bot_num
	return s, nil
}



func getCriteriaString(s *PointSalad, veg_type VegType, id int) string {
	return s.str_criterias[int(veg_type) + id * VEGETABLE_TYPE_NUM]
}

func broadcast_to_all(out map[int]chan []byte, str string) {
	fmt.Print(str)
	for _, value  := range out {
		value <- []byte(str)
	}
}

func displayMarket(s *PointSalad, out map[int]chan []byte) {
	broadcast_to_all(out, fmt.Sprintf("---- MARKET ----\n"))
	for i, cardspot := range s.market {
		if cardspot.has_card {
			card := cardspot.card
			broadcast_to_all(out, fmt.Sprintf("[%c] %v\n", i + 'A', card.Vegetable_type))
		}
	}
	fmt.Println("piles")
	for i, pile := range s.piles {
		if len(pile) > 0 {
			top_card := pile[len(pile) - 1]
			broadcast_to_all(out, fmt.Sprintf("[%d] %s\n", i, getCriteriaString(s, top_card.Vegetable_type, top_card.Id)))
		} else {
			broadcast_to_all(out, "\n")
		}
	}
}

func drawFromTop(s *PointSalad, pile_index int) Card {
	assert(len(s.piles[pile_index]) > 0)
	c := s.piles[pile_index][len(s.piles[pile_index]) - 1]
	s.piles[pile_index] = s.piles[pile_index][0:len(s.piles[pile_index]) - 1]
	return c	
}

func drawFromBot(s *PointSalad, pile_index int) Card {
	assert(len(s.piles[pile_index]) > 0)
	c := s.piles[pile_index][0]
	s.piles[pile_index] = s.piles[pile_index][1:len(s.piles[pile_index])]
	return c
}

func getMaxPileIndex(s *PointSalad) int {
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

func displayActorCards(s *PointSalad, out map[int]chan []byte) {
	assert(s.active_actor < len(s.actor_data))
	broadcast_to_all(out, fmt.Sprintf("---- Player %d ----\n", s.active_actor))

	broadcast_to_all(out, fmt.Sprintf("%d current score\n", calculateScore(s, s.active_actor)))
	broadcast_to_all(out, "--------\n")
	for i, num := range s.actor_data[s.active_actor].vegetable_num {
		broadcast_to_all(out, fmt.Sprintf("%d %v\n", num, VegType(i)))
		
	}

	broadcast_to_all(out, fmt.Sprintf("---- point cards ----\n"))

	for i, card := range s.actor_data[s.active_actor].point_pile {
		broadcast_to_all(out, fmt.Sprintf("%d: %s\n", i, getCriteriaString(s, card.Vegetable_type, card.Id)))
	}
}

func flipCardsFromPiles(s *PointSalad) {
	for y := range s.piles {
		for x := range 2 {
			market_pos := y + x * PLAY_PILES_NUM
			if !s.market[market_pos].has_card {
				if len(s.piles[y]) == 0 {
					s.market[market_pos].card = drawFromTop(s, y)
					s.market[market_pos].has_card = true
			
				} else {
					index := getMaxPileIndex(s)
					s.market[market_pos].card = drawFromBot(s, index)
					s.market[market_pos].has_card = true
				}
			}
		}
	}
}


func isWithinAtoF(a byte) bool {
	return a >= 'A' && a <= 'F'
}


func pickCardToChangeToVeg(s *PointSalad, in chan []byte, out chan []byte) {
	for true {
		if len(s.actor_data[s.active_actor].point_pile) == 0 {
			break
		}
		out <- []byte(fmt.Sprintf("pick 0-1 point card to flip to vegetable, type n to pick none example: 5\n"))
		input := <- in

		if input[0] == 'n' {
			break
		}

		index, err := strconv.Atoi(string(input))
		if err != nil {
			continue
		}

		if index >= 0 && index < len(s.actor_data[s.active_actor].point_pile) {
			

			card := s.actor_data[s.active_actor].point_pile[index]

			s.actor_data[s.active_actor].vegetable_num[int(card.Vegetable_type)] += 1
			

			// shift slice
			for i := index; i < len(s.actor_data[s.active_actor].point_pile) - 1; i += 1 {
				s.actor_data[s.active_actor].point_pile[i] = s.actor_data[s.active_actor].point_pile[i + 1] 
			}
			// remove last element
			s.actor_data[s.active_actor].point_pile = s.actor_data[s.active_actor].point_pile[0:len(s.actor_data[s.active_actor].point_pile) - 1]
			break
		}
	}
}


func calculateScore(s *PointSalad, actor_id int) int {
	score := 0

	for _, point_card := range s.actor_data[actor_id].point_pile {
		_ = point_card	
		var criteria Criteria
		
		switch criteria.criteria_type {
			case MOST: {
				is_most := true
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					} 
					max := s.actor_data[actor_id].vegetable_num[0]
					max_id := 0
					for _, actor_data := range s.actor_data {
						if actor_data.vegetable_num[j] > max {
							max = actor_data.vegetable_num[j]
							max_id = j
						}
					}
					if max_id != actor_id {
						is_most = false
						break 
					}
				}
				if is_most {
					score += criteria.single_score
				}
			} 
			case FEWEST: {
				is_fewest := true
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					} 
					min := s.actor_data[0].vegetable_num[0]
					min_id := 0
					for _, actor_data := range s.actor_data {
						if actor_data.vegetable_num[j] < min {
							min = actor_data.vegetable_num[j]
							min_id = j
						}
					}
					if min_id != actor_id {
						is_fewest = false
						break 
					}
				}
				if is_fewest {
					score += criteria.single_score
				}
			}
			case EVEN_ODD: {
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					}
					if s.actor_data[actor_id].vegetable_num[VegType(j)] % 2 == 0 {
						score += criteria.even_score
					} else {
						score += criteria.odd_score
					}
				}
			}
			case PER: {
				for j, per_value := range criteria.per_scores {
					score += s.actor_data[actor_id].vegetable_num[j] * per_value
				}
			}
			case SUM: {
				min := math.MaxInt32
				for j, count := range criteria.veg_count {
					if count == 0 {
						continue
					} 
					non_repeated_value := s.actor_data[actor_id].vegetable_num[j] / count
					if non_repeated_value < min {
						min = non_repeated_value
					}
				}
				score += min * criteria.single_score
			}
			case MOST_TOTAL: {
				veg_count := 0
				for _, count := range s.actor_data[actor_id].vegetable_num {
					veg_count += count
				}

				is_most  := true
				for _, actor_data := range s.actor_data {
					other_veg_count := 0
					for  _, count := range actor_data.vegetable_num {
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
			case FEWEST_TOTAL: {
				veg_count := 0
				for _, count := range s.actor_data[actor_id].vegetable_num {
					veg_count += count
				}

				is_fewest := true
				for _, actor_data := range s.actor_data {
					other_veg_count := 0
					for  _, count := range actor_data.vegetable_num {
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
			case PER_TYPE_GREATER_THAN_EQ: {
				for _, count := range s.actor_data[actor_id].vegetable_num {
					if count >= criteria.greater_than_eq_value {
						score += criteria.single_score
					}
				}
			}
			case PER_MISSING_TYPE: {
				for _, count := range s.actor_data[actor_id].vegetable_num {
					if count == 0 {
						score += criteria.single_score
					}
				}
			}
			case COMPLETE_SET: {
				min := s.actor_data[actor_id].vegetable_num[0]
				for _, count := range s.actor_data[actor_id].vegetable_num {
					if count < min {
						min = count
					}
				}
				score += criteria.single_score * min
			} 
			default: {
				assert(false)
			}
		}

	}

	return score
}



type Game interface {
	init(player_num int, bot_num int)
	update(in map[int]chan []byte, out map[int]chan []byte)
}


type ActorActionType int
const (
	INVALID ActorActionType = iota
	PICK_VEG_FROM_MARKET ActorActionType = iota
	PICK_POINT_FROM_MARKET ActorActionType = iota
	PICK_TO_SWAP ActorActionType = iota
)
type ActorAction struct {
	kind ActorActionType
	amount int
	ids [2]int
}

func getMarketActionFromBot(s *PointSalad) ActorAction {

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

func getSwapActionFromBot(s *PointSalad) ActorAction {
	assert(len(s.actor_data[s.active_actor].point_pile) > 0)

	action := ActorAction{}
	for true {
		action.kind = PICK_TO_SWAP
		action.amount = rand.Intn(2)

		for i := range action.amount {
			n := len(s.actor_data[s.active_actor].point_pile)
			action.ids[i] = rand.Intn(n)
		}

		err := IsActionLegal(s, action)
		if err == nil {
			break
		}
	}
	return action
}



func getMarketActionFromPlayer(s *PointSalad, in chan []byte, out chan []byte) ActorAction {
	assert(in != nil)
	assert(out != nil)

	action := ActorAction{}
	for true {
		out <- []byte(fmt.Sprintf("pick 1 or 2 vegetables example: AB or\npick 1 point card example: 0\n"))
		input := <- in

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

func getSwapActionFromPlayer(s *PointSalad, in chan []byte, out chan []byte) ActorAction {
	assert(in != nil)
	assert(out != nil)
	assert(len(s.actor_data[s.active_actor].point_pile) > 0)

	action := ActorAction{}
	for true {
		out <- []byte(fmt.Sprintf("pick 0-1 point card to flip to vegetable, type n to pick none example: 5\n"))
		input := <- in

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

func IsActionLegal(s *PointSalad, action ActorAction) (error) {
	switch (action.kind) {
		case INVALID: return fmt.Errorf("Invalid action kind")
		case PICK_VEG_FROM_MARKET: {
			if action.amount < 1 || action.amount > 2 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			for i := range action.amount {
				if action.ids[i] < 0 || action.ids[i] >= len(s.market) {
					return fmt.Errorf("Cannot take card outside of market range")
				}
				if !s.market[action.ids[i]].has_card {
					return fmt.Errorf("Cannot take card from empty market spot")
				}
			}
		}
		case PICK_POINT_FROM_MARKET: {
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
		case PICK_TO_SWAP: {
			if action.amount < 0 || action.amount > 1 {
				return fmt.Errorf("Amount of actions outside of range: %d", action.amount)
			}
			if action.amount == 1 {
				if action.ids[0] < 0 || action.ids[0] >= len(s.actor_data[s.active_actor].point_pile) {
					return fmt.Errorf("Cannot take card outside of pile range")
				}
			}
		}
	}
	return nil
}

func doAction(s *PointSalad, action ActorAction) {
	assert(IsActionLegal(s, action) == nil)

	switch (action.kind) {
		case INVALID: panic("unreachable")
		case PICK_VEG_FROM_MARKET: {
			for i := range action.amount {
				card := s.market[action.ids[i]].card
				s.actor_data[s.active_actor].vegetable_num[card.Vegetable_type] += 1
				s.market[action.ids[i]].has_card = false
			}
		} 
		case PICK_POINT_FROM_MARKET: {
			for i := range action.amount {
				card := drawFromTop(s, action.ids[i])
				s.actor_data[s.active_actor].point_pile = append(s.actor_data[s.active_actor].point_pile, card)
			}
		}
		case PICK_TO_SWAP: {
			if action.amount == 1 {
				veg_type := s.actor_data[s.active_actor].point_pile[action.ids[0]].Vegetable_type
				s.actor_data[s.active_actor].vegetable_num[int(veg_type)] += 1

				// remove element
				for i := action.ids[0]; i < len(s.actor_data[s.active_actor].point_pile) - 1; i += 1 {
					s.actor_data[s.active_actor].point_pile[i] = s.actor_data[s.active_actor].point_pile[i + 1]
				}
				s.actor_data[s.active_actor].point_pile = s.actor_data[s.active_actor].point_pile[0:len(s.actor_data[s.active_actor].point_pile) - 1]
			}
		}
	}
}

// should be called before doAction
func BroadcastAction(s *PointSalad, action ActorAction, out map[int]chan []byte) {
	broadcast_to_all(out, "---- Action ----\n")
	switch (action.kind) {
		case INVALID: panic("unreachable")
		case PICK_VEG_FROM_MARKET: {
			for i := range action.amount {
				broadcast_to_all(out, fmt.Sprintf("Player %d drew %v from market\n", s.active_actor, s.market[action.ids[i]].card.Vegetable_type.String()))
			}
		}
		case PICK_POINT_FROM_MARKET: {
			for i := range action.amount {
				pile := s.piles[action.ids[i]]
				card := pile[len(pile) - 1]
				criteria := getCriteriaString(s, card.Vegetable_type, card.Id)
				broadcast_to_all(out, fmt.Sprintf("Player %d drew %v from market\n", s.active_actor, criteria))
			}
		}
		case PICK_TO_SWAP: {
			if action.amount == 0 {
				broadcast_to_all(out, fmt.Sprintf("Player %d did not swap any card\n", s.active_actor))
			} else {
				for i := range action.amount {
					card := s.actor_data[s.active_actor].point_pile[action.ids[i]]
					criteria := getCriteriaString(s, card.Vegetable_type, card.Id)
					broadcast_to_all(out, fmt.Sprintf("Player %d swapped %v to %v\n", s.active_actor, criteria, card.Vegetable_type))
				}
			}
		}
	}
}



func client_read(conn net.Conn) {
	for true {
		buf := make([]byte, SERVER_BYTE_SEND_SIZE, SERVER_BYTE_SEND_SIZE)
		_, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		length := 0
		for i := range buf {
			if buf[i] == 0 {
				break
			}
			length += 1
		}
		fmt.Print(string(buf[0:length]))
	}
}

func client_send(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for true {
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		str := s[0: len(s) - 2]
	
		if len(str) > 0 && str[0] == 'q' {
			conn.Close()
			break
		}

		conn.Write([]byte(str))
	}
}

func main() {
	var is_server bool
	var hostname string
	var port string
	var player_num int
	var bot_num int


	flag.BoolVar(&is_server, "server", false, "ex. -server")
	flag.StringVar(&hostname, "hostname", "127.0.0.1", "ex. 127.0.0.1")
	flag.StringVar(&port, "port", "8080", "ex. 8080")
	flag.IntVar(&player_num, "players", 1, "ex. 2")
	flag.IntVar(&bot_num, "bots", 1, "ex. 2")
	flag.Parse()


	log.Printf("is_server = %v, hostname = %v port = %v player_num = %v bot_num = %v\n", is_server, hostname, port, player_num, bot_num)
	if is_server {
		var game Game
		game = &PointSalad{}
		game.init(player_num, bot_num)

		server := Server{}
		server_init(&server, port, player_num)
	
	
		for true {
			game.update(server.connections.in, server.connections.out)
		}
		
		server_close(&server)

	} else {
		conn, err := net.Dial("tcp", hostname + ":" + port)
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		go client_read(conn)
		client_send(conn)
	}
}