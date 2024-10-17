package pointsalad

import (
	"fmt"
	"strconv"
)

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
	veg_count [vegetableTypeNum]int
	// used for single score rules
	single_score int
	// used for greater than rules
	greater_than_eq_value int
	// used for even odd rules
	even_score int
	odd_score int
	// used for per rules
	per_scores [vegetableTypeNum]int
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
	raw_src string
	tokens []Token
	index int
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
	if  lex.index + pos >= len(lex.tokens) {
		return Token{}, false
	}
	return lex.tokens[lex.index + pos], true
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
		first := getToken(&lex) 
		if first.token_type == IDENTIFIER {
			if first.s == "MOST" || first.s == "FEWEST" {
				t := nextToken(&lex) 
				if t.s == "TOTAL" {

					err := expectNextTokenStr(&lex, "VEGETABLE")
					if err != nil {
						return Criteria{}, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err
					}

					num, err := parseNumber(&lex, nextToken(&lex))
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
					err := expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err 
					}
					num, err := parseNumber(&lex, nextToken(&lex))
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
					return Criteria{}, fmt.Errorf("Expected TOTAL or a vegetable type")
				}
			} else if first.s == "COMPLETE" {
				err := expectNextTokenStr(&lex, "SET")
				if err != nil {
					return Criteria{}, err
				}
				err = expectNextTokenType(&lex, EQUAL)
				if err != nil {
					return Criteria{}, err
				}

				num, err := parseNumber(&lex, nextToken(&lex))
				if err != nil {
					return criteria, err
				}

				criteria.criteria_type = COMPLETE_SET
				criteria.single_score = num

				return criteria, nil

			} else if isVegetable(first.s) {
				v := getVegetableType(first.s)
				t := nextToken(&lex)
				if t.token_type == COLON {
					err := expectNextTokenStr(&lex, "EVEN")
					if err != nil {
						return Criteria{}, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err
					}

					even, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return criteria, err
					}
					err = expectNextTokenType(&lex, COMMA)
					if err != nil {
						return Criteria{}, err
					}
					err = expectNextTokenStr(&lex, "ODD")
					if err != nil {
						return Criteria{}, err
					}
					err = expectNextTokenType(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err
					}

					odd, err := parseNumber(&lex, nextToken(&lex))
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
						t := nextToken(&lex)
						if !isVegetable(t.s) {
							return Criteria{}, fmt.Errorf("Expected vegetable type here")
						}
						v := getVegetableType(t.s)
						criteria.veg_count[int(v)] += 1
						if lex.index >= len(lex.tokens) - 1 || nextToken(&lex).token_type != PLUS {
							break
						}
					}
					err := expectTokenType(&lex, EQUAL)
					if err != nil {
						return Criteria{}, err
					}
					num, err := parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return Criteria{}, err
					}

					criteria.criteria_type = SUM
					criteria.single_score = num

					return criteria, nil

				} else {
					return Criteria{}, fmt.Errorf("Expected COLON or PLUS here")
				}
			} else {
				return Criteria{}, fmt.Errorf("Expected MOST or FEWEST OR COMPLETE or a vegetable type here")
			}
		} else if first.token_type == NUMBER {
			num, err := parseNumber(&lex, first)
			if err != nil {
				return Criteria{}, err
			}
			err = expectNextTokenType(&lex, SLASH)
			if err != nil {
				return Criteria{}, err
			}

			t := nextToken(&lex)
			if t.s == "VEGETABLE" {
				err = expectNextTokenStr(&lex, "TYPE")
				if err != nil {
					return Criteria{}, err
				}
				err = expectNextTokenType(&lex, GREATER)
				if err != nil {
					return Criteria{}, err
				}
				err = expectNextTokenType(&lex, EQUAL)
				if err != nil {
					return Criteria{}, err
				}

				num2, err := parseNumber(&lex, nextToken(&lex))
				if err != nil {
					return Criteria{}, err
				}

				criteria.criteria_type = PER_TYPE_GREATER_THAN_EQ
				criteria.single_score = num
				criteria.greater_than_eq_value = num2

				return criteria, nil

			} else if t.s == "MISSING" {
				err = expectNextTokenStr(&lex, "VEGETABLE")
				if err != nil {
					return Criteria{}, err
				}
				err = expectNextTokenStr(&lex, "TYPE")
				if err != nil {
					return Criteria{}, err
				}

				criteria.criteria_type = PER_MISSING_TYPE
				criteria.single_score = num

				return criteria, nil
			} else {
				for true {
					if !isVegetable(t.s) {
						return Criteria{}, fmt.Errorf("Expected vegetable type here")
					}
					v := getVegetableType(t.s)
					criteria.per_scores[int(v)] = num

					if lex.index >= len(lex.tokens) - 1 || nextToken(&lex).token_type != COMMA {
						break
					}
					num, err = parseNumber(&lex, nextToken(&lex))
					if err != nil {
						return Criteria{}, err
					}
					err = expectNextTokenType(&lex, SLASH)
					if err != nil {
						return Criteria{}, err
					}
					t = nextToken(&lex)
				}
	
				criteria.criteria_type = PER
	
				return criteria, nil
			} 
		} else {
			return Criteria{}, fmt.Errorf("Expected Identifier or number as first token")
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