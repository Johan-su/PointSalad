package pointsalad

import (
	"fmt"
	"math"
	"strconv"
)

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


type Criteria interface {
	calculateScore(s *GameState, actorId int) int
}

type CriteriaMost struct {
	vegType VegType
	score   int
}

func (c *CriteriaMost) calculateScore(s *GameState, actorId int) int {
	vegType := int(c.vegType)

	max := math.MinInt32
	maxId := -1
	for i, actorData := range s.actorData {
		if actorData.vegetableNum[vegType] > max {
			max = actorData.vegetableNum[vegType]
			maxId = i
		}
	}
	if maxId != actorId {
		return 0
	}
	return c.score
}

func (c *CriteriaMost) String() string {
	return fmt.Sprintf("MOST %v = %v", c.vegType, c.score)
}

type CriteriaFewest struct {
	vegType VegType
	score   int
}

func (c *CriteriaFewest) calculateScore(s *GameState, actorId int) int {
	vegType := int(c.vegType)

	min := math.MaxInt32
	minId := -1
	for i, actorData := range s.actorData {
		if actorData.vegetableNum[vegType] < min {
			min = actorData.vegetableNum[vegType]
			minId = i
		}
	}
	if minId != actorId {
		return 0
	}
	return c.score
}

func (c *CriteriaFewest) String() string {
	return fmt.Sprintf("FEWEST %v = %v", c.vegType, c.score)
}

type CriteriaEvenOdd struct {
	vegType   VegType
	evenScore int
	oddScore  int
}

func (c *CriteriaEvenOdd) calculateScore(s *GameState, actorId int) int {
	if s.actorData[actorId].vegetableNum[c.vegType]%2 == 0 {
		return c.evenScore
	} else {
		return c.oddScore
	}
}

type CriteriaPer struct {
	perScores [vegetableTypeNum]int
}

func (c *CriteriaPer) calculateScore(s *GameState, actorId int) int {
	score := 0
	for j, per_value := range c.perScores {
		score += s.actorData[actorId].vegetableNum[j] * per_value
	}
	return score
}

type CriteriaSum struct {
	vegCount [vegetableTypeNum]int
	score    int
}

func (c *CriteriaSum) calculateScore(s *GameState, actorId int) int {
	min := math.MaxInt32
	for j, count := range c.vegCount {
		if count == 0 {
			continue
		}
		non_repeated_value := s.actorData[actorId].vegetableNum[j] / count
		if non_repeated_value < min {
			min = non_repeated_value
		}
	}
	if min == math.MaxInt32 {
		return 0
	}
	return min * c.score
}

type CriteriaMostTotal struct {
	score int
}

func (c *CriteriaMostTotal) calculateScore(s *GameState, actorId int) int {
	vegCount := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		vegCount += count
	}

	for _, actorData := range s.actorData {
		other_vegCount := 0
		for _, count := range actorData.vegetableNum {
			other_vegCount += count
		}
		if other_vegCount >= vegCount {
			return 0
		}
	}
	return c.score
}

type CriteriaFewestTotal struct {
	score int
}

func (c *CriteriaFewestTotal) calculateScore(s *GameState, actorId int) int {
	vegCount := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		vegCount += count
	}

	for _, actorData := range s.actorData {
		other_vegCount := 0
		for _, count := range actorData.vegetableNum {
			other_vegCount += count
		}
		if other_vegCount <= vegCount {
			return 0
		}
	}
	return c.score
}

type CriteriaPerTypeGreaterThanEq struct {
	greaterThanEq int
	score         int
}

func (c *CriteriaPerTypeGreaterThanEq) calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		if count >= c.greaterThanEq {
			score += c.score
		}
	}
	return score
}

type CriteriaPerMissingType struct {
	score int
}

func (c *CriteriaPerMissingType) calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		if count == 0 {
			score += c.score
		}
	}
	return score
}

type CriteriaCompleteSet struct {
	score int
}

func (c *CriteriaCompleteSet) calculateScore(s *GameState, actorId int) int {
	min := s.actorData[actorId].vegetableNum[0]
	for _, count := range s.actorData[actorId].vegetableNum {
		if count < min {
			min = count
		}
	}
	return c.score * min
}

type TokenType int

const (
	IDENTIFIER TokenType = iota
	EQUAL      TokenType = iota
	NUMBER     TokenType = iota
	COLON      TokenType = iota
	COMMA      TokenType = iota
	SLASH      TokenType = iota
	PLUS       TokenType = iota
	MINUS      TokenType = iota
	GREATER    TokenType = iota
)

type Token struct {
	token_type TokenType
	s          string
}

type Lexer struct {
	raw_src string
	tokens  []Token
	index   int
}

func getToken(lex *Lexer) Token {
	assert(lex.index < len(lex.tokens))
	return lex.tokens[lex.index]
}

func nextToken(lex *Lexer) Token {
	lex.index += 1
	assert(lex.index < len(lex.tokens))
	return lex.tokens[lex.index]
}

func expectTokenStr(lex *Lexer, str string) error {
	token := getToken(lex)
	if token.s != str {
		return fmt.Errorf("Expected %v got %v in %s\n", str, token.s, lex.raw_src)
	}
	return nil
}

func expectNextTokenStr(lex *Lexer, str string) error {
	token := nextToken(lex)
	if token.s != str {
		return fmt.Errorf("Expected %v got %v in %s\n", str, token.s, lex.raw_src)
	}
	return nil
}

func expectTokenType(lex *Lexer, token_type TokenType) error {
	token := getToken(lex)
	if token.token_type != token_type {
		return fmt.Errorf("Expected %v got %v in %s\n", token_type, token.token_type, lex.raw_src)
	}
	return nil
}

func expectNextTokenType(lex *Lexer, token_type TokenType) error {
	token := nextToken(lex)
	if token.token_type != token_type {
		return fmt.Errorf("Expected %v got %v in %s\n", token_type, token.token_type, lex.raw_src)
	}
	return nil
}

func lookToken(lex *Lexer, pos int) (Token, bool) {
	if lex.index+pos >= len(lex.tokens) {
		return Token{}, false
	}
	return lex.tokens[lex.index+pos], true
}

func parseNumber(lex *Lexer, minus_or_num Token) (int, error) {
	var err error
	num := 0
	if minus_or_num.token_type == MINUS {
		t := nextToken(lex)
		expectTokenType(lex, NUMBER)

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

	for i := range vegetableTypeNum {
		if VegType(i).String() == s {
			return true
		}
	}

	return false
}

func getVegetableType(s string) VegType {
	for i := range vegetableTypeNum {
		if VegType(i).String() == s {
			return VegType(i)
		}
	}
	panic("unreachable")
	return -1
}

func parseCriteria(s string) (Criteria, error) {

	lex := Lexer{}
	lex.raw_src = s

	// tokenize the criterias
	for i := 0; i < len(s); {
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
		} else if isDigit(s[i]) {
			start := i
			end := i
			for i < len(s) && isDigit(s[i]) {
				i += 1
				end += 1
			}

			token := Token{NUMBER, s[start:end]}
			lex.tokens = append(lex.tokens, token)
		} else if s[i] == '=' {
			lex.tokens = append(lex.tokens, Token{EQUAL, s[i : i+1]})
			i += 1
		} else if s[i] == ':' {
			lex.tokens = append(lex.tokens, Token{COLON, s[i : i+1]})
			i += 1
		} else if s[i] == ',' {
			lex.tokens = append(lex.tokens, Token{COMMA, s[i : i+1]})
			i += 1
		} else if s[i] == '/' {
			lex.tokens = append(lex.tokens, Token{SLASH, s[i : i+1]})
			i += 1
		} else if s[i] == '+' {
			lex.tokens = append(lex.tokens, Token{PLUS, s[i : i+1]})
			i += 1
		} else if s[i] == '-' {
			lex.tokens = append(lex.tokens, Token{MINUS, s[i : i+1]})
			i += 1
		} else if s[i] == '>' {
			lex.tokens = append(lex.tokens, Token{GREATER, s[i : i+1]})
			i += 1
		} else {
			return nil, fmt.Errorf("Unknown character: %c", s[i])
		}
	}

	for lex.index = 0; lex.index < len(lex.tokens); lex.index += 1 {
		first := getToken(&lex)
		if first.token_type == IDENTIFIER {
			if first.s == "MOST" || first.s == "FEWEST" {
				t := nextToken(&lex)
				if t.s == "TOTAL" {

					err := expectNextTokenStr(&lex, "VEGETABLE")
					if err != nil {
						return nil, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return nil, err
					}

					num, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}

					if first.s == "MOST" {
						return &CriteriaMostTotal{score: num}, nil
					} else if first.s == "FEWEST" {
						return &CriteriaFewestTotal{score: num}, nil
					} else {
						panic("unreachable")
					}

				} else if isVegetable(t.s) {
					v := getVegetableType(t.s)
					err := expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return nil, err
					}
					num, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}

					if first.s == "MOST" {
						return &CriteriaMost{vegType: v, score: num}, nil
					} else if first.s == "FEWEST" {
						return &CriteriaFewest{vegType: v, score: num}, nil
					} else {
						panic("unreachable")
					}
				} else {
					return nil, fmt.Errorf("Expected TOTAL or a vegetable type")
				}
			} else if first.s == "COMPLETE" {
				err := expectNextTokenStr(&lex, "SET")
				if err != nil {
					return nil, err
				}
				err = expectNextTokenType(&lex, EQUAL)
				if err != nil {
					return nil, err
				}

				num, err := parseNumber(&lex, nextToken(&lex))
				if err != nil {
					return nil, err
				}

				return &CriteriaCompleteSet{score: num}, nil

			} else if isVegetable(first.s) {
				v := getVegetableType(first.s)
				t := nextToken(&lex)
				if t.token_type == COLON {
					err := expectNextTokenStr(&lex, "EVEN")
					if err != nil {
						return nil, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return nil, err
					}

					even, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}
					err = expectNextTokenType(&lex, COMMA)
					if err != nil {
						return nil, err
					}
					err = expectNextTokenStr(&lex, "ODD")
					if err != nil {
						return nil, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return nil, err
					}

					odd, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}

					return &CriteriaEvenOdd{vegType: v, evenScore: even, oddScore: odd}, nil

				} else if t.token_type == PLUS {
					var vegCount [vegetableTypeNum]int
					vegCount[int(v)] += 1
					for true {
						t := nextToken(&lex)
						if !isVegetable(t.s) {
							return nil, fmt.Errorf("Expected vegetable type")
						}
						v := getVegetableType(t.s)
						vegCount[int(v)] += 1
						if lex.index >= len(lex.tokens)-1 || nextToken(&lex).token_type != PLUS {
							break
						}
					}
					err := expectTokenType(&lex, EQUAL)
					if err != nil {
						return nil, err
					}
					num, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}

					return &CriteriaSum{vegCount: vegCount, score: num}, nil

				} else {
					return nil, fmt.Errorf("Expected COLON or PLUS here")
				}
			} else {
				return nil, fmt.Errorf("Expected MOST or FEWEST OR COMPLETE or a vegetable type here")
			}
		} else if first.token_type == NUMBER {
			num, err := parseNumber(&lex, first)
			if err != nil {
				return nil, err
			}
			err = expectNextTokenType(&lex, SLASH)
			if err != nil {
				return nil, err
			}

			t := nextToken(&lex)
			if t.s == "VEGETABLE" {
				err = expectNextTokenStr(&lex, "TYPE")
				if err != nil {
					return nil, err
				}
				err = expectNextTokenType(&lex, GREATER)
				if err != nil {
					return nil, err
				}
				err = expectNextTokenType(&lex, EQUAL)
				if err != nil {
					return nil, err
				}

				num2, err := parseNumber(&lex, nextToken(&lex))
				if err != nil {
					return nil, err
				}

				return &CriteriaPerTypeGreaterThanEq{greaterThanEq: num2, score: num}, nil

			} else if t.s == "MISSING" {
				err = expectNextTokenStr(&lex, "VEGETABLE")
				if err != nil {
					return nil, err
				}
				err = expectNextTokenStr(&lex, "TYPE")
				if err != nil {
					return nil, err
				}

				return &CriteriaPerMissingType{score: num}, nil

			} else {
				var perScores [vegetableTypeNum]int
				for true {
					if !isVegetable(t.s) {
						return nil, fmt.Errorf("Expected vegetable type here")
					}
					v := getVegetableType(t.s)
					perScores[int(v)] = num

					if lex.index >= len(lex.tokens)-1 || nextToken(&lex).token_type != COMMA {
						break
					}
					num, err = parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return nil, err
					}
					err = expectNextTokenType(&lex, SLASH)
					if err != nil {
						return nil, err
					}
					t = nextToken(&lex)
				}

				return &CriteriaPer{perScores: perScores}, nil
			}
		} else {
			return nil, fmt.Errorf("Expected Identifier or number as first token")
		}
	}
	panic("unreachable")
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
