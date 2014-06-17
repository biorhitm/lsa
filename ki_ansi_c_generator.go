package lsa

import (
	"fmt"
)

type TLanguageItemType uint

//TLanguageItemType типы синтаксичекских элементов
const (
	ltitUnknown = TLanguageItemType(iota)
	ltitEOF
	ltitIdent
	ltitAssignment
	ltitOpenParenthesis
	ltitCloseParenthesis
	ltitNumber
	ltitMathOperation //TODO: для каждой операции свой код
)

type TLanguageItem struct {
	Type TLanguageItemType
	// идентификаторы будут держать здесь номер строки из массива
	// всех идентификаторов
	Index uint
}

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

var strNumbers = make([]string, 0, 1024)
var strIdents = make([]string, 0, 1024)

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

/*
АРГУМЕНТ = {'('} <ПРОСТОЙ АРГУМЕНТ> {')'}
ПРОСТОЙ АРГУМЕНТ = <ЧИСЛО> | <СЛОЖНЫЙ ИДЕНТИФИКАТОР> | <ВЫЗОВ ФУНКЦИИ>
  | <СИМВОЛ> | <СТРОКА>
ВЫЗОВ ФУНКЦИИ = <СЛОЖНЫЙ ИДЕНТИФИКАТОР> '(' [<ПАРАМЕТРЫ>] ')'
ПАРАМЕТРЫ = [<ВЫРАЖЕНИЕ>] {',' [<ВЫРАЖЕНИЕ>]}
*/
func (L *TLexem) translateArgument(AItems *[]TLanguageItem) error {
	var item TLanguageItem

	// пропускаю необязательные открывающие скобки
	for L.Type == ltOpenParenthesis {
		L = L.Next
		item = TLanguageItem{Type: ltitOpenParenthesis}
		*AItems = append(*AItems, item)
		parenthesis++
	}

	switch L.Type {
	case ltNumber:
		{
			item = TLanguageItem{Type: ltitNumber, Index: uint(len(strNumbers))}
			*AItems = append(*AItems, item)
			strNumbers = append(strNumbers, (*L).LexemAsString())
			L = L.Next
		}

	default:
		{
			return L.errorAt(EExpectedArgument)
		}
	}

	// пропускаю необязательные закрывающие скобки
	for L.Type == ltCloseParenthesis {
		L = L.Next
		item = TLanguageItem{Type: ltitCloseParenthesis}
		*AItems = append(*AItems, item)
		parenthesis--
	}
	return nil
}

/*
BNF-определения для присваивания выражения переменной
<СЛОЖНЫЙ ИДЕНТИФИКАТОР> '=' <ВЫРАЖЕНИЕ>
СЛОЖНЫЙ ИДЕНТИФИКАТОР = <ИДЕНТИФИКАТОР> {' ' <ИДЕНТИФИКАТОР>}
ВЫРАЖЕНИЕ = <АРГУМЕНТ> {<ОПЕРАЦИЯ> <АРГУМЕНТ>}
ОПЕРАЦИЯ = '+' | '-' | '*' | '/' | '%' | '^'
*/
func (lexem *TLexem) translateAssignment(AItems *[]TLanguageItem) error {
	var item TLanguageItem
	var E error

	if lexem.Type == ltIdent {
		cnt := uint(len(strIdents))
		strIdents = append(strIdents, lexem.LexemAsString())
		lexem = lexem.Next

		for lexem.Type == ltIdent {
			strIdents[cnt] = strIdents[cnt] + " " + lexem.LexemAsString()
			lexem = lexem.Next
		}
		item = TLanguageItem{Type: ltitIdent, Index: cnt}
		*AItems = append(*AItems, item)
	} else {
		return lexem.errorAt(ESyntaxError)
	}

	if lexem.Type == ltEqualSign {
		item = TLanguageItem{Type: ltitAssignment}
		*AItems = append(*AItems, item)
		lexem = lexem.Next // пропускаю знак =
	} else {
		return lexem.errorAt(ESyntaxError)
	}

	if lexem.Type == ltEOF {
		return lexem.errorAt(EExpectedExpression)
	}
	lexem = lexem.skipEOL()

	parenthesis = 0

	// обрабатываю аргумент
	E = lexem.translateArgument(AItems)
	if E != nil {
		return E
	}

	// далее может серия аргументов через знаки операций
Loop:
	for {
		switch lexem.Type {
		case ltSemicolon:
			{
				break Loop
			}

		//TODO: вынести проверку операции в функцию и добавить проверку
		//      остальных операций: > < >= <= >> << ! ~
		case ltPlus, ltMinus, ltStar, ltSlash, ltPercent:
			{
				item = TLanguageItem{Type: ltitMathOperation}
				*AItems = append(*AItems, item)
				lexem = lexem.Next

				E = lexem.translateArgument(AItems)
				if E != nil {
					return E
				}
			}

		default:
			{
				break Loop
			}
		}
	}

	if parenthesis < 0 {
		return lexem.errorAt(ETooMuchCloseRB)
	}
	if parenthesis > 0 {
		return lexem.errorAt(ETooMuchOpenRB)
	}

	return nil
}

/*
 Переводит текст в лексемах в массив элементов языка
*/
func TranslateCode(ALexem PLexem) ([]TLanguageItem, error) {
	languageItems := make([]TLanguageItem, 5, 4000)
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
				E = (*startLexem).translateAssignment(&languageItems)
				if E != nil {
					return nil, E
				}
				ALexem = nextLexem
			}

		case ltEOL:
			{
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
	return languageItems, nil
}
