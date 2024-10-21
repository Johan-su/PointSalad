package pointsalad

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

// getJCriteria returns the specific criteria associated with a vegetable type
// for a given card ID in the provided JCards structure. The function accesses
// the appropriate vegetable criteria based on the VegType (such as PEPPER, LETTUCE, etc.)
// and the card ID, then returns the corresponding criteria as a string.
//
// Parameters:
// - jsonCards: A pointer to the JCards structure that contains the card data.
// - vegType: The type of vegetable (e.g., PEPPER, LETTUCE, etc.) for which criteria is needed.
// - id: The card ID within the jsonCards structure to retrieve the criteria for.
//
// Returns:
// - string: The criteria associated with the specified vegetable type and card ID.
//
// Panics:
// - The function panics if the vegetable type provided does not match any known type (unreachable case).
func getJCriteria(jsonCards *JCards, vegType VegType, id int) string {
	Criteria := jsonCards.Cards[id].Criteria
	switch vegType {
	case PEPPER:
		return Criteria.PEPPER
	case LETTUCE:
		return Criteria.LETTUCE
	case CARROT:
		return Criteria.CARROT
	case CABBAGE:
		return Criteria.CABBAGE
	case ONION:
		return Criteria.ONION
	case TOMATO:
		return Criteria.TOMATO
	}
	panic("unreachable")
}

type Criteria interface {
	calculateScore(s *GameState, actorId int) int
	String() string
}

type CriteriaMost struct {
	vegType VegType
	score   int
}

// calculateScore calculates the score based on the actor's vegetable count for a specific vegetable type.
// It checks whether the actor with the given actorId has the highest vegetable count for the specified vegetable type (vegType).
// If the actor does not have the highest count, the function returns 0. Otherwise, it returns the score associated with this criterion.
// 
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor.
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

// String returns a string representation of the CriteriaMost object, indicating the vegetable type and the associated score.
// The format is: "MOST <vegetable type> = <score>"
//
// Returns:
//   - string: The string representation of the criteria.
func (c *CriteriaMost) String() string {
	return fmt.Sprintf("MOST %v = %v", c.vegType, c.score)
}

type CriteriaFewest struct {
	vegType VegType
	score   int
}

// calculateScore calculates the score based on the actor's vegetable count for a specific vegetable type,
// but this time it checks whether the actor with the given actorId has the fewest vegetable count for the specified vegetable type (vegType).
// If the actor does not have the fewest count, the function returns 0. Otherwise, it returns the score associated with this criterion.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor.
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

// String returns a string representation of the CriteriaFewest object, indicating the vegetable type and the associated score.
// The format is: "FEWEST <vegetable type> = <score>"
//
// Returns:
//   - string: The string representation of the criteria.
func (c *CriteriaFewest) String() string {
	return fmt.Sprintf("FEWEST %v = %v", c.vegType, c.score)
}

type CriteriaEvenOdd struct {
	vegType   VegType
	evenScore int
	oddScore  int
}

// calculateScore calculates the score based on the actor's vegetable count for a specific vegetable type,
// determining the score based on whether the count is even or odd.
// If the vegetable count for the specified type (vegType) is even, the function returns the evenScore;
// otherwise, it returns the oddScore.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score based on whether the vegetable count is even or odd.
func (c *CriteriaEvenOdd) calculateScore(s *GameState, actorId int) int {
	if s.actorData[actorId].vegetableNum[c.vegType]%2 == 0 {
		return c.evenScore
	} else {
		return c.oddScore
	}
}

// String returns a string representation of the CriteriaEvenOdd object,
// showing the vegetable type and the associated even and odd scores.
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "<vegetable type>: EVEN=<evenScore>, ODD=<oddScore>"
func (c *CriteriaEvenOdd) String() string {
	return fmt.Sprintf("%v: EVEN=%v, ODD=%v", c.vegType, c.evenScore, c.oddScore)
}

type CriteriaPer struct {
	perScores [vegetableTypeNum]int
}

// calculateScore calculates the total score for an actor based on the vegetable quantities and their associated per-vegetable scores.
// For each vegetable type, the actor's vegetable count is multiplied by the corresponding per-value from perScores, and the results are summed to get the total score.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor, based on the perScores for each vegetable type.
func (c *CriteriaPer) calculateScore(s *GameState, actorId int) int {
	score := 0
	for j, per_value := range c.perScores {
		score += s.actorData[actorId].vegetableNum[j] * per_value
	}
	return score
}

// String returns a string representation of the CriteriaPer object, showing the vegetable types and their associated per-scores.
// The format is: "<vegetable type> / <score>" for each vegetable type where the score is non-zero. The result is a comma-separated list.
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "<vegetable type> / <score>, ..."
func (c *CriteriaPer) String() string {
	builder := strings.Builder{}

	first := true
	for i, score := range c.perScores {
		if score != 0 {
			if first {
				builder.WriteString(fmt.Sprintf("%v / %v", VegType(i), score))
				first = false
			} else {
				builder.WriteString(fmt.Sprintf(", %v / %v", VegType(i), score))
			}
		}
	}
	return builder.String()
}

type CriteriaSum struct {
	vegCount [vegetableTypeNum]int
	score    int
}

// calculateScore calculates the score for an actor based on the vegetable counts and their associated "vegCount" values.
// For each vegetable type, it checks how many times the actor has that vegetable type divided by the required count (from vegCount).
// The minimum value across all valid (non-zero) vegetable counts is taken, and the score is multiplied by this minimum value.
// If no valid vegetable types are found, the score is 0.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor, which is the minimum of the non-repeated vegetable counts, multiplied by the given score.
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

// String returns a string representation of the CriteriaSum object, showing the vegetable types used in the calculation
// and the associated score. The format is: "<vegetable type> + <vegetable type> = <score>", where the vegetable types
// are listed based on the vegCount array.
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "<vegetable type> + <vegetable type>, ... = <score>"
func (c *CriteriaSum) String() string {
	builder := strings.Builder{}

	first := true
	for i, count := range c.vegCount {
		for range count {
			if first {
				builder.WriteString(fmt.Sprintf("%v", VegType(i)))
				first = false
			} else {
				builder.WriteString(fmt.Sprintf(" + %v", VegType(i)))
			}
		}
	}
	builder.WriteString(fmt.Sprintf(" = %v", c.score))
	return builder.String()
}

type CriteriaMostTotal struct {
	score int
}

// calculateScore calculates the score for an actor based on their total number of vegetables compared to all other actors.
// The actor's total vegetable count is compared to the total vegetable count of each other actor. If any other actor has
// an equal or greater vegetable count, the score is set to 0. If the actor has more vegetables than all other actors, the
// score is returned.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor, which is 0 if any other actor has an equal or greater total vegetable count,
//     otherwise it returns the score associated with this criterion.
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

// String returns a string representation of the CriteriaMostTotal object, indicating that the criterion is based on the
// actor having the most total vegetables. The format is: "MOST TOTAL VEGETABLE = <score>"
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "MOST TOTAL VEGETABLE = <score>"
func (c *CriteriaMostTotal) String() string {
	return fmt.Sprintf("MOST TOTAL VEGETABLE = %v", c.score)
}

type CriteriaFewestTotal struct {
	score int
}

// calculateScore calculates the score for an actor based on their total number of vegetables compared to all other actors.
// The actor's total vegetable count is compared to the total vegetable count of each other actor. If any other actor has
// an equal or fewer total vegetable count, the score is set to 0. If the actor has the fewest vegetables compared to all
// other actors, the score is returned.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score for the actor, which is 0 if any other actor has an equal or fewer total vegetable count,
//     otherwise it returns the score associated with this criterion.
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

// String returns a string representation of the CriteriaFewestTotal object, indicating that the criterion is based on the
// actor having the fewest total vegetables. The format is: "FEWEST TOTAL VEGETABLE = <score>"
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "FEWEST TOTAL VEGETABLE = <score>"
func (c *CriteriaFewestTotal) String() string {
	return fmt.Sprintf("FEWEST TOTAL VEGETABLE = %v", c.score)
}

type CriteriaPerTypeGreaterThanEq struct {
	greaterThanEq int
	score         int
}

// calculateScore calculates the score for an actor based on the number of vegetables they have for each vegetable type.
// If the actor's count for a vegetable type is greater than or equal to the specified threshold (greaterThanEq),
// the score is increased by the given score value (c.score) for each such vegetable type.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score, which is the sum of the score values for each vegetable type where the actor's count
//     is greater than or equal to the threshold (c.greaterThanEq).
func (c *CriteriaPerTypeGreaterThanEq) calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		if count >= c.greaterThanEq {
			score += c.score
		}
	}
	return score
}

// String returns a string representation of the CriteriaPerTypeGreaterThanEq object, indicating the threshold for each
// vegetable type and the associated score. The format is: "<vegetable type> / VEGETABLE TYPE >=<threshold>"
// 
// Returns:
//   - string: The string representation of the criteria in the format:
//     "<vegetable type> / VEGETABLE TYPE >=<threshold>"
func (c *CriteriaPerTypeGreaterThanEq) String() string {
	return fmt.Sprintf("%v / VEGETABLE TYPE >=%v", c.greaterThanEq, c.score)
}

type CriteriaPerMissingType struct {
	score int
}

// calculateScore calculates the score for an actor based on the number of vegetable types they are missing (i.e., vegetable count is 0).
// For each vegetable type, if the actor's count is 0 (indicating they are missing that vegetable type), the score is increased
// by the specified score value (c.score) for each missing vegetable type.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score, which is the sum of the score values for each vegetable type the actor is missing (count == 0).
func (c *CriteriaPerMissingType) calculateScore(s *GameState, actorId int) int {
	score := 0
	for _, count := range s.actorData[actorId].vegetableNum {
		if count == 0 {
			score += c.score
		}
	}
	return score
}

// String returns a string representation of the CriteriaPerMissingType object, indicating the associated score for each missing vegetable type.
// The format is: "<score> / MISSING VEGETABLE TYPE"
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "<score> / MISSING VEGETABLE TYPE"
func (c *CriteriaPerMissingType) String() string {
	return fmt.Sprintf("%v / MISSING VEGETABLE TYPE", c.score)
}

type CriteriaCompleteSet struct {
	score int
}

// calculateScore calculates the score for an actor based on the minimum vegetable count across all vegetable types they have.
// The actor must have a complete set of vegetables, and the score is determined by multiplying the minimum count of any vegetable type
// by the given score value (c.score). If the actor has a low count in one type, that number limits the total score.
//
// Parameters:
//   - s *GameState: The current state of the game containing actor data and vegetable counts.
//   - actorId int: The ID of the actor whose score is being calculated.
//
// Returns:
//   - int: The calculated score, which is the product of the score value and the minimum vegetable count the actor has
//     for any vegetable type. If the actor has fewer of any vegetable type, that count limits the score.
func (c *CriteriaCompleteSet) calculateScore(s *GameState, actorId int) int {
	min := s.actorData[actorId].vegetableNum[0]
	for _, count := range s.actorData[actorId].vegetableNum {
		if count < min {
			min = count
		}
	}
	return c.score * min
}

// String returns a string representation of the CriteriaCompleteSet object, indicating that the criterion is based on having
// a complete set of vegetables, with the associated score. The format is: "COMPLETE SET = <score>"
//
// Returns:
//   - string: The string representation of the criteria in the format:
//     "COMPLETE SET = <score>"
func (c *CriteriaCompleteSet) String() string {
	return fmt.Sprintf("COMPLETE SET = %v", c.score)
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

// parseCriteria parses a string representing a criteria into a Criteria object.
// The function takes a string `s` and tokenizes it into a series of tokens using a lexer. It then identifies the specific criteria
// based on keywords and patterns found in the string and constructs the corresponding Criteria object, which can be of various types
// such as CriteriaMost, CriteriaFewest, CriteriaCompleteSet, CriteriaPer, etc. The function handles different formats of criteria,
// including operations on vegetable types, and supports complex expressions like "MOST TOTAL VEGETABLE" or "MOST / VEGETABLE TYPE >= N".
//
// Parameters:
//   - s string: The string representing the criteria to be parsed.
//
// Returns:
//   - Criteria: A Criteria object representing the parsed criteria. This can be of types such as CriteriaMost, CriteriaFewest,
//     CriteriaCompleteSet, CriteriaPer, etc.
//   - error: An error if the parsing fails, such as invalid syntax, unexpected tokens, or unsupported formats.
//
// Example usage:
//   - "MOST TOTAL VEGETABLE = 5" will return a CriteriaMostTotal object with score 5.
//   - "FEWEST / VEGETABLE TYPE >= 2" will return a CriteriaPerTypeGreaterThanEq object with greaterThanEq set to 2 and score 5.
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