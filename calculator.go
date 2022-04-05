// calculator
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var (
	errInput             = errors.New("Неизвестная ошибка ввода.")
	errExtClosingBrace   = errors.New("Обнаружена лишняя закрывающая скобка.")
	errExtOpeningBrace   = errors.New("Обнаружена лишняя открывающая скобка.")
	errOperatorIsNotTerm = errors.New("Оператор не является термом.")
	errBraceIsNotTerm    = errors.New("Скобка не является термом.")
)

func main() {
	runes, err := input()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tokens, err := toTokens(runes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//testTokens(tokens)

	groupedTokens, err := GroupTokens(tokens)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//fmt.Printf("%v: %#v\n", groupedTokens, groupedTokens)

	term, err := groupedTokens.ToTerm()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//fmt.Printf("%v: %#v\n", term, term)

	fmt.Println(term.Compute())
}

// Чтение строки из стандартного ввода
func input() ([]rune, error) {
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return []rune{}, errInput
	}
	return []rune(input), nil
}

// Объявление типов и констант, которые понадобятся для токенизации

type Token interface { // Токен может быть числом, операцией или скобкой
	Type() TokenType
	IsOperator() bool
	ToTerm() (Term, error)
}

type TokenType byte
type GroupedTokensType byte
type NumberToken int64
type BraceToken byte
type OperatorToken byte

var (
	NumberTokenType                    TokenType         = 0
	OperatorTokenType                  TokenType         = 1
	BraceTokenType                     TokenType         = 2
	LeftBrace                          BraceToken        = 0
	RightBrace                         BraceToken        = 1
	AdditionOperator                   OperatorToken     = 0
	SubtractionOperator                OperatorToken     = 1
	MultiplicationOperator             OperatorToken     = 2
	DivisionOperator                   OperatorToken     = 3
	TokenGroupedTokensType             GroupedTokensType = 0
	GroupedTokensListGroupedTokensType GroupedTokensType = 2
)

func (n NumberToken) Type() TokenType {
	return NumberTokenType
}
func (o OperatorToken) Type() TokenType {
	return OperatorTokenType
}
func (b BraceToken) Type() TokenType {
	return BraceTokenType
}

func toTokens(runes []rune) ([]Token, error) {
	tokens := []Token{}
	numStr := []byte{} // Здесь будут накапливаться цифры для парсинга числа

	for _, rune := range runes {
		switch rune {
		// если встретилось нечисло, то закончить начатый ввод числа:
		case '(', ')', '+', '-', '*', '/':
			if len(numStr) != 0 {
				num, err := strconv.ParseUint(string(numStr), 0, 32)
				if err != nil {
					return tokens, err
				}
				tokens = append(tokens, NumberToken(num))
				numStr = []byte{}
			}
			switch rune {
			case '(':
				tokens = append(tokens, LeftBrace)
			case ')':
				tokens = append(tokens, RightBrace)
			case '+':
				tokens = append(tokens, AdditionOperator)
			case '-':
				tokens = append(tokens, SubtractionOperator)
			case '*':
				tokens = append(tokens, MultiplicationOperator)
			case '/':
				tokens = append(tokens, DivisionOperator)
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			numStr = append(numStr, byte(rune))
		default:
			return tokens, errors.New("Неизвестный символ: '" + string(rune) + "'")
		}
	}

	// Если строка закончилась, то закончить ввод числа, который
	// мог не закончиться во время цикла
	if len(numStr) != 0 {
		num, err := strconv.ParseUint(string(numStr), 0, 32)
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, NumberToken(num))
	}

	return tokens, nil
}

func testTokens(tokens []Token) { // Напечатать каждый токен на отдельной строке
	for _, t := range tokens {
		switch token := t.(type) {
		case BraceToken:
			if token == LeftBrace {
				fmt.Println("Левая скобка")
			} else {
				fmt.Println("Правая скобка")
			}
		case NumberToken:
			fmt.Printf("Число: %d\n", token)
		case OperatorToken:
			switch token {
			case AdditionOperator:
				fmt.Println("Операция сложение")
			case SubtractionOperator:
				fmt.Println("Операция вычитание")
			case MultiplicationOperator:
				fmt.Println("Операция умножение")
			case DivisionOperator:
				fmt.Println("Операция деление")
			}
		}
	}
}

// Тип обозначает один токен либо список токенов, не содержащий скобок.
// Может содержать элемент своего типа
type GroupedTokens interface {
	IsOperator() bool
	// Такой список сгруппированных токенов можно превратить в терм (см. далее):
	ToTerm() (Term, error)
}

type GroupedTokensList []GroupedTokens

// Эта функция будет рекурсивно применятся к каждому выражению внутри скобок
func GroupTokens(tokens []Token) (GroupedTokens, error) {
	groupedTokens := GroupedTokensList{}
	// Сюда будут записываться все токены внутри скобок
	// первого уровня вложенности:
	innerTokens := []Token{}

	braceCount := 0       // Подсчёт количества скобок
	placeInBrace := false // Флаг, записывать внутрь скобок или в основной список

	for _, t := range tokens {
		switch token := t.(type) {
		case NumberToken, OperatorToken:
			if placeInBrace {
				innerTokens = append(innerTokens, token)
			} else {
				groupedTokens = append(groupedTokens, token)
			}
		case BraceToken:
			switch token {
			case LeftBrace:
				braceCount++
				// Если сейчас идёт запись внутрь скобок, то скобка записывается
				// туда же вместе с остальными токенами
				if placeInBrace {
					innerTokens = append(innerTokens, token)
				}
				placeInBrace = true
			case RightBrace:
				braceCount--
				if braceCount < 0 {
					return nil, errExtClosingBrace
				} else if braceCount == 0 {
					// Закрылась скобка первого уровня вложенности.
					// Всё, что внутри, рекурсивно сгруппировать
					innerGroup, err := GroupTokens(innerTokens)
					if err != nil {
						return nil, err
					}
					// И добавить в список
					groupedTokens = append(groupedTokens, innerGroup)
					// Обнулить переменные, в которых в будущем будет
					// содержимое новых скобок
					placeInBrace = false
					innerTokens = []Token{}
				} else {
					innerTokens = append(innerTokens, token)
				}
			}
		}
	}
	if braceCount > 0 {
		return nil, errExtOpeningBrace
	}

	return groupedTokens, nil
}

// Терм — это то, что можно вычислить
type Term interface {
	Compute() int64
}

// Это либо число:
func (num NumberToken) Compute() int64 { return int64(num) }

// Либо выражение, состоящее из операции и двух операндов (термов):
type ExprTerm struct {
	LeftOperand  Term
	RightOperand Term
	Operation    OperatorToken
}

func (ex ExprTerm) Compute() int64 {
	leftValue := ex.LeftOperand.Compute()
	rightValue := ex.RightOperand.Compute()
	switch ex.Operation {
	case AdditionOperator:
		return leftValue + rightValue
	case SubtractionOperator:
		return leftValue - rightValue
	case MultiplicationOperator:
		return leftValue * rightValue
	case DivisionOperator:
		return leftValue / rightValue
	}
	return 0 // без этого не работает
}

// Конвертирование токенов и сгруппированных токенов в терм

func (t NumberToken) IsOperator() bool       { return false }
func (t OperatorToken) IsOperator() bool     { return true }
func (t BraceToken) IsOperator() bool        { return false }
func (t GroupedTokensList) IsOperator() bool { return false }

func (t NumberToken) ToTerm() (Term, error)   { return t, nil }
func (t OperatorToken) ToTerm() (Term, error) { return nil, errOperatorIsNotTerm }
func (t BraceToken) ToTerm() (Term, error)    { return nil, errBraceIsNotTerm }

func (grTokens GroupedTokensList) ToTerm() (Term, error) {
	//fmt.Println(grTokens)
	if len(grTokens) == 1 {
		// Если список с единственным токеном, то превратить его в терм
		term, err := grTokens[0].ToTerm()
		if err != nil {
			return nil, err
		}
		return term, nil
	} else {
		// Если список с несколькими токенами, то для начала найдём
		// позицию первого оператора сложения или вычитания,
		// и позицию последнего оператора умножения или деления.
		var AddOrSubExists bool
		var AddOrSubIndex, MulOrDivIndex int
		for i, el := range grTokens {
			if el.IsOperator() {
				switch el {
				case AdditionOperator, SubtractionOperator:
					AddOrSubExists = true
					AddOrSubIndex = i
				case MultiplicationOperator, DivisionOperator:
					MulOrDivIndex = i
				}
			}
		}
		if AddOrSubExists {
			return GrTokensToExprTerm(grTokens, AddOrSubIndex)
		} else {
			return GrTokensToExprTerm(grTokens, MulOrDivIndex)
		}
	}

	return nil, nil
}

func GrTokensToExprTerm(grTokens GroupedTokensList, operatorIndex int) (ExprTerm, error) {
	// Функция принимает список токенов и индекс, по которому находится оператор
	// Части списка до и после оператора превращает в термы, возвращает терм
	if operatorIndex == 0 {
		return ExprTerm{}, errors.New("Слева от оператора нет операнда")
	} else if operatorIndex == len(grTokens)-1 {
		return ExprTerm{}, errors.New("Справа от оператора нет операнда")
	}
	leftOperand, err := grTokens[:operatorIndex].ToTerm()
	if err != nil {
		return ExprTerm{}, err
	}
	rightOperand, err := grTokens[(operatorIndex + 1):].ToTerm()
	if err != nil {
		return ExprTerm{}, err
	}
	return ExprTerm{
		LeftOperand:  leftOperand,
		Operation:    grTokens[operatorIndex].(OperatorToken),
		RightOperand: rightOperand,
	}, nil
}
