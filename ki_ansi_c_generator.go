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

type TSyntaxDescriptor struct {
	Lexem         *TLexem
	LanguageItems []TLanguageItem
	Parenthesis   int
	StrNumbers    []string //= make([]string, 0, 1024)
	StrIdents     []string //make([]string, 0, 1024)
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

func (self *TSyntaxDescriptor) Init() {
	self.Lexem = nil
	self.Parenthesis = 0
	self.LanguageItems = make([]TLanguageItem, 0, 0)
	self.StrIdents = make([]string, 0, 0)
	self.StrNumbers = make([]string, 0, 0)
}

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

func (L *TLexem) errorAt(E *lsaError) error {
	E.LineNo = L.LineNo
	E.ColumnNo = L.ColumnNo
	return E
}

func (Self *TSyntaxDescriptor) translateComplexIdent() error {
	if Self.Lexem.Type != ltIdent {
		return Self.Lexem.errorAt(&lsaError{Msg: "Can't translateComplexIdent, type not ltIdent."})
	}

	cnt := uint(len(Self.StrIdents))
	Self.StrIdents = append(Self.StrIdents, Self.Lexem.LexemAsString())
	Self.Lexem = Self.Lexem.Next

	for Self.Lexem.Type == ltIdent {
		Self.StrIdents[cnt] += " " + Self.Lexem.LexemAsString()
		Self.Lexem = Self.Lexem.Next
	}
	item := TLanguageItem{Type: ltitIdent, Index: cnt}
	Self.LanguageItems = append(Self.LanguageItems, item)
	return nil
}

//TODO: распознание символа как аргумента
//TODO: распознание строки как аргумента
//TODO: распознание вызова функции как аргумента
/*
АРГУМЕНТ = {'('} <ПРОСТОЙ АРГУМЕНТ> {')'}
ПРОСТОЙ АРГУМЕНТ = <ЧИСЛО> | <СЛОЖНЫЙ ИДЕНТИФИКАТОР> | <ВЫЗОВ ФУНКЦИИ>
  | <СИМВОЛ> | <СТРОКА>
ВЫЗОВ ФУНКЦИИ = <СЛОЖНЫЙ ИДЕНТИФИКАТОР> '(' [<ПАРАМЕТРЫ>] ')'
ПАРАМЕТРЫ = [<ВЫРАЖЕНИЕ>] {',' [<ВЫРАЖЕНИЕ>]}
*/
func (Self *TSyntaxDescriptor) translateArgument() error {
	var item TLanguageItem
	var S string

	// пропускаю необязательные открывающие скобки
	for Self.Lexem.Type == ltOpenParenthesis {
		Self.Lexem = Self.Lexem.Next
		item = TLanguageItem{Type: ltitOpenParenthesis}
		Self.LanguageItems = append(Self.LanguageItems, item)
		Self.Parenthesis++
	}

	switch Self.Lexem.Type {
	case ltNumber:
		{
			item = TLanguageItem{Type: ltitNumber,
				Index: uint(len(Self.StrNumbers))}
			Self.LanguageItems = append(Self.LanguageItems, item)
			S = Self.Lexem.LexemAsString()
			Self.StrNumbers = append(Self.StrNumbers, S)
			Self.Lexem = Self.Lexem.Next
		}

	case ltIdent:
		{
			E := Self.translateComplexIdent()
			if E != nil {
				return Self.Lexem.errorAt(ESyntaxError)
			}
		}

	default:
		{
			return Self.Lexem.errorAt(EExpectedArgument)
		}
	}

	// пропускаю необязательные закрывающие скобки
	for Self.Lexem.Type == ltCloseParenthesis {
		Self.Lexem = Self.Lexem.Next
		item = TLanguageItem{Type: ltitCloseParenthesis}
		Self.LanguageItems = append(Self.LanguageItems, item)
		Self.Parenthesis--
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
func (Self *TSyntaxDescriptor) translateAssignment() error {
	var item TLanguageItem
	var E error

	E = Self.translateComplexIdent()
	if E != nil {
		return Self.Lexem.errorAt(ESyntaxError)
	}

	if Self.Lexem.Type == ltEqualSign {
		item = TLanguageItem{Type: ltitAssignment}
		Self.LanguageItems = append(Self.LanguageItems, item)
		Self.Lexem = Self.Lexem.Next // пропускаю знак =
	} else {
		return Self.Lexem.errorAt(ESyntaxError)
	}

	if Self.Lexem.Type == ltEOF {
		return Self.Lexem.errorAt(EExpectedExpression)
	}
	Self.Lexem = Self.Lexem.skipEOL()

	Self.Parenthesis = 0

	// обрабатываю аргумент
	E = Self.translateArgument()
	if E != nil {
		return E
	}

	// далее может серия аргументов через знаки операций
Loop:
	for {
		switch Self.Lexem.Type {
		case ltSemicolon:
			{
				break Loop
			}

		//TODO: вынести проверку операции в функцию и добавить проверку
		//      остальных операций: > < >= <= >> << ! ~
		case ltPlus, ltMinus, ltStar, ltSlash, ltPercent:
			{
				item = TLanguageItem{Type: ltitMathOperation}
				Self.LanguageItems = append(Self.LanguageItems, item)
				Self.Lexem = Self.Lexem.Next

				E = Self.translateArgument()
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

	if Self.Parenthesis < 0 {
		return Self.Lexem.errorAt(ETooMuchCloseRB)
	}
	if Self.Parenthesis > 0 {
		return Self.Lexem.errorAt(ETooMuchOpenRB)
	}

	return nil
}

/*
 Переводит текст в лексемах в массив элементов языка
*/
func TranslateCode(ALexem PLexem) (TSyntaxDescriptor, error) {
	syntaxDescriptor := TSyntaxDescriptor{
		Lexem:         ALexem,
		LanguageItems: make([]TLanguageItem, 0, 1000),
		Parenthesis:   0,
		StrNumbers:    make([]string, 0, 1024),
		StrIdents:     make([]string, 0, 1024),
	}

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
				syntaxDescriptor.Lexem = startLexem
				E = syntaxDescriptor.translateAssignment()
				if E != nil {
					return TSyntaxDescriptor{}, E
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
	return syntaxDescriptor, nil
}
