package lsa

import (
	"fmt"
)

func getLexemAfterLexem(ALexem PLexem, _type TLexemType, text string) PLexem {
	for ALexem != nil {
		if ALexem.Type == _type {
			if text != "" && (*ALexem).LexemAsString() == text {
				return ALexem.Next
			}
		}

		ALexem = ALexem.Next
	}
	return ALexem
}

type TKeyword struct {
	Id   int
	Name string
}

// KeywordsIds
const (
	kwiUnknown = iota
	kwiFunction
	kwiProcedure
)

var (
	keywordList = []TKeyword{
		TKeyword{kwiFunction, "функция"},
		TKeyword{kwiProcedure, "процедура"},
		TKeyword{kwiUnknown, ""},
	}
)

// ошибки синтаксиса
var (
	EExpectedExpression = &lsaError{Msg: "Отсутствует выражение после знака ="}
	ESyntaxError        = &lsaError{Msg: "Синтаксическая ошибка"}
	EExpectedArgument   = &lsaError{Msg: "Ожидается операнд"}
	ETooMuchCloseRB     = &lsaError{Msg: "Слишком много )"}
	ETooMuchOpenRB      = &lsaError{Msg: "Слишком много ("}
)

func (self *TLexem) toKeywordId() int {
	S := (*self).LexemAsString()
	for i := 0; i < len(keywordList); i++ {
		if S == keywordList[i].Name {
			return keywordList[i].Id
		}
	}
	return kwiUnknown
}

/*
  'функция' <имя функции> '(' {<имя параметра> ':' <тип>} ')'
  <локальные переменные>
  'начало' <тело функции> 'конец'
*/
func generateFunction(ALexem PLexem) {
	var S string
	var L PLexem
	functionName := ALexem

	parameter := getLexemAfterLexem(ALexem, ltOpenParenthesis, "")
	L = parameter
	for L != nil {
		if L.Type == ltCloseParenthesis {
			L = L.Next
			break
		}
		L = L.Next
	}

	if L != nil && L.Type == ltColon {
		L = L.Next
	}

	// печатаю тип функции
	for L != nil {
		if L.Type == ltEOL {
			L = L.Next
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)

		L = L.Next
	}

	localVars := L

	// печатаю имя функции
	L = functionName
	for L != nil {
		if L.Type == ltOpenParenthesis {
			fmt.Print(" (")
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)
		L = L.Next
	}

	// печатаю параметры функции
	L = parameter
	for L != nil {
		if L.Type == ltCloseParenthesis {
			fmt.Print(" )")
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)
		L = L.Next
	}

	L = getLexemAfterLexem(localVars, ltIdent, "начало")

	fmt.Print(" {\n")

	// печатаю тело функции
	for L != nil {
		if L.Type == ltIdent && (*L).LexemAsString() == "конец" {
			break
		}
		S = (*L).LexemAsString()
		fmt.Printf("\t%s\n", S)
		L = L.Next
	}
	fmt.Print("}\n")
}

func (self *TLexem) skipEOL() PLexem {
	if self.Type == ltEOL {
		return self.Next
	}
	return self
}

var parenthesis int = 0

func (L *TLexem) errorAt(E *lsaError) error {
	E.LineNo = L.LineNo
	E.ColumnNo = L.ColumnNo
	return E
}

func (L *TLexem) translateArgument() (PLexem, error) {
	// пропускаю необязательные открывающие скобки
	for L.Type == ltOpenParenthesis {
		L = L.Next
		if L.Type == ltEOL {
			return nil, L.errorAt(EExpectedArgument)
		}
		fmt.Print("(")
		parenthesis++
	}

	switch L.Type {
	case ltNumber:
		{
			fmt.Printf("%s", (*L).LexemAsString())
			L = L.Next
		}

	default:
		{
			return nil, L.errorAt(EExpectedArgument)
		}
	}

	// пропускаю необязательные закрывающие скобки
	for L.Type == ltCloseParenthesis {
		L = L.Next
		fmt.Print(")")
		parenthesis--
	}

	return L, nil
}

/*
BNF-определения для присваивания выражения переменной
<СЛОЖНЫЙ ИДЕНТИФИКАТОР> '=' <ВЫРАЖЕНИЕ>
СЛОЖНЫЙ ИДЕНТИФИКАТОР = <ИДЕНТИФИКАТОР> {' ' <ИДЕНТИФИКАТОР>}
ВЫРАЖЕНИЕ = [<LF>] <АРГУМЕНТ> [<LF>] {<ОПЕРАЦИЯ> [<LF>] <АРГУМЕНТ> [<LF>]}
АРГУМЕНТ = {'('} <ПРОСТОЙ АРГУМЕНТ> {')'}
ПРОСТОЙ АРГУМЕНТ = <ЧИСЛО> | <СЛОЖНЫЙ ИДЕНТИФИКАТОР> | <ВЫЗОВ ФУНКЦИИ>
  | <СИМВОЛ> | <СТРОКА>
ВЫЗОВ ФУНКЦИИ = <СЛОЖНЫЙ ИДЕНТИФИКАТОР> '(' [<ПАРАМЕТРЫ>] ')'
ПАРАМЕТРЫ = [<ВЫРАЖЕНИЕ>] {',' [<ВЫРАЖЕНИЕ>]}
ОПЕРАЦИЯ = '+' | '-' | '*' | '/' | '%' | '^'
*/
func (L *TLexem) translateAssignment() (PLexem, error) {
	var E error

	for L.Type == ltIdent {
		fmt.Printf("%s ", L.LexemAsString())
		L = L.Next
	}

	if L.Type != ltEqualSign {
		return nil, L.errorAt(ESyntaxError)
	}

	fmt.Printf("= ")
	L = L.Next // пропускаю знак =

	if L.Type == ltEOF {
		return nil, L.errorAt(EExpectedExpression)
	}
	L = L.skipEOL()

	parenthesis = 0

	// обрабатываю аргумент
	L, E = L.translateArgument()
	if E != nil {
		return nil, E
	}

	// далее может серия аргументов через знаки операций
Loop:
	for {
		switch L.Type {
		case ltSemicolon:
			{
				break Loop
			}

		//TODO: вынести проверку операции в функцию и добавить проверку
		//      остальных операций: > < >= <= >> << ! ~
		case ltPlus, ltMinus, ltStar, ltSlash, ltPercent:
			{
				fmt.Printf(" %c ", L.Type)
				L = L.Next
				L, E = L.translateArgument()
				if E != nil {
					return nil, E
				}
			}

		default:
			{
				break Loop
			}
		}
	}

	fmt.Printf("\n")

	if parenthesis < 0 {
		return nil, L.errorAt(ETooMuchCloseRB)
	}
	if parenthesis > 0 {
		return nil, L.errorAt(ETooMuchOpenRB)
	}

	return L, nil
}

/*
 Переводит текст в лексемах в код на языке С
*/
func TranslateCode(ALexem PLexem) error {
	// лексема к которой надо будет вернуться, чтобы продолжить перевод
	// например, если название переменной состоит из нескольких слов, то
	// после нахождения знака =, надо будет вернуться к первому слову
	// переменной, чтобы записать её полное название
	startLexem := ALexem

	var nextLexem PLexem = nil
	var E error = nil

	for ALexem != nil {
		switch ALexem.Type {
		case ltIdent:
			{
				keywordId := (*ALexem).toKeywordId()
				if keywordId == kwiFunction {
					generateFunction(ALexem.Next)
				} else {
					ALexem = ALexem.Next
				}
			}

		case ltEqualSign:
			{
				nextLexem, E = (*startLexem).translateAssignment()
				if E != nil {
					return E
				}
				ALexem = nextLexem
			}

		case ltEOL:
			{
				fmt.Println()
				ALexem = ALexem.Next
			}

		default:
			if ALexem.Size > 0 {
				fmt.Printf("Лехема: %d size: %d %s ",
					ALexem.Type, ALexem.Size, (*ALexem).LexemAsString())
			}
			ALexem = ALexem.Next
		}
	}

	fmt.Printf("----------EOF-----------\n")
	return nil
}
