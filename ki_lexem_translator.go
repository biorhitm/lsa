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
	ltitString
	ltitChar
	ltitMathAdd
	ltitMathSub
	ltitMathMul
	ltitMathDiv
	ltitModulo
	ltitInvolution
	ltitOR
	ltitAND
	ltitXOR
	ltitNOT
	ltitFunctionDeclaration
	ltitColon
)

type TLanguageItem struct {
	Type TLanguageItemType
	// идентификаторы будут держать здесь номер строки из массива
	// всех идентификаторов
	Index uint
}

type TStringArray []string

type TSyntaxDescriptor struct {
	Lexem         *TLexem
	LanguageItems []TLanguageItem
	Parenthesis   int
	StrNumbers    TStringArray
	StrIdents     TStringArray
	StrStrings    TStringArray
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
	self.StrStrings = make([]string, 0, 0)
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
BNF-правила для объявления функции
ОБЪЯВЛЕНИЕ ФУНКЦИИ = <ФУНКЦИЯ> <ИМЯ ФУНКЦИИ> '(' [<ПАРАМЕТРЫ>] ')' [<РЕЗУЛЬТАТ>]
  [<ЛОКАЛЬНЫЕ ПЕРЕМЕННЫЕ>] <НАЧАЛО> [<ТЕЛО ФУНКЦИИ>] <КОНЕЦ>
ФУНКЦИЯ = 'функция' | 'function' | 'func' | 'def'
НАЧАЛО = 'начало' | 'begin' | '{'
КОНЕЦ = 'конец' | 'end' | '}'
ИМЯ ФУНКЦИИ = [<ИМЯ КЛАССА> '.']<ИДЕНТИФИКАТОР>
ИМЯ КЛАССА = <ИДЕНТИФИКАТОР>
ПАРАМЕТРЫ = <ПАРАМЕТР> {',' <ПАРАМЕТР>}
ПАРАМЕТР = <ИМЯ ПАРАМЕТРА> [':' <ТИП>]
РЕЗУЛЬТАТ = ':' <ТИП>
ТИП = [<ИМЯ ПАКЕТА> '.']<ИДЕНТИФИКАТОР>
ИМЯ ПАКЕТА = <ИДЕНТИФИКАТОР>
<ЛОКАЛЬНЫЕ ПЕРЕМЕННЫЕ> = 'переменные' | 'var'
*/
func (self *TSyntaxDescriptor) translateFunctionDeclaration() error {
	var (
		E    error
		S    string
		item TLanguageItem
	)

	S = self.Lexem.LexemAsString()
	//TODO: преобразовать лексему в ключевое слово
	if S == "функция" || S == "function" || S == "func" || S == "def" {
		item = TLanguageItem{Type: ltitFunctionDeclaration}
		self.LanguageItems = append(self.LanguageItems, item)

		self.Lexem = self.Lexem.Next
		E = self.translateComplexIdent()
		if E != nil {
			return self.Lexem.errorAt(&lsaError{Msg: E.Error()})
		}

		if self.Lexem.Type != ltOpenParenthesis {
			return self.Lexem.errorAt(&lsaError{Msg: "Отсутствует '('"})
		}
		item = TLanguageItem{Type: ltitOpenParenthesis}
		self.LanguageItems = append(self.LanguageItems, item)
		self.Lexem = self.Lexem.Next

		if self.Lexem.Type != ltCloseParenthesis {
			for {
				E = self.translateComplexIdent()
				if E != nil {
					//TODO: Улучшить сообщение об ошибке
					return self.Lexem.errorAt(&lsaError{
						Msg: "Ожидается имя параметра. " + E.Error()})
				}

				if self.Lexem.Type == ltCloseParenthesis {
					item = TLanguageItem{Type: ltitCloseParenthesis}
					self.LanguageItems = append(self.LanguageItems, item)
					self.Lexem = self.Lexem.Next
					break
				}
			}
		}

		// читаю тип возвращаемого значения
		if self.Lexem.Type == ltColon {
			item = TLanguageItem{Type: ltitColon}
			self.LanguageItems = append(self.LanguageItems, item)
			self.Lexem = self.Lexem.Next

			E = self.translateComplexIdent()
			if E != nil {
				//TODO: Улучшить сообщение об ошибке
				return self.Lexem.errorAt(&lsaError{
					Msg: "Ожидается тип возвращаемого значения. " + E.Error()})
			}
		}

	} else {
		return self.Lexem.errorAt(&lsaError{Msg: "Can't translateFunctionDeclaration, type not function."})
	}

	return nil
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

func (list *TStringArray) addUnique(S string) uint {
	for i, v := range *list {
		if v == S {
			return uint(i)
		}
	}
	(*list) = append((*list), S)
	return uint(len(*list) - 1)
}

func (self *TSyntaxDescriptor) translateNumber() error {
	S := self.Lexem.LexemAsString()
	index := self.StrNumbers.addUnique(S)

	item := TLanguageItem{Type: ltitNumber, Index: index}
	self.LanguageItems = append(self.LanguageItems, item)

	self.Lexem = self.Lexem.Next
	return nil
}

func (Self *TSyntaxDescriptor) translateComplexIdent() error {
	if Self.Lexem.Type != ltIdent {
		return Self.Lexem.errorAt(&lsaError{Msg: "Can't translateComplexIdent, type not ltIdent."})
	}

	S := Self.Lexem.LexemAsString()
	Self.Lexem = Self.Lexem.Next
	for Self.Lexem.Type == ltIdent {
		S += " " + Self.Lexem.LexemAsString()
		Self.Lexem = Self.Lexem.Next
	}

	index := Self.StrIdents.addUnique(S)
	item := TLanguageItem{Type: ltitIdent, Index: index}
	Self.LanguageItems = append(Self.LanguageItems, item)
	return nil
}

func (self *TSyntaxDescriptor) translateString() error {
	if self.Lexem.Type != ltString {
		return self.Lexem.errorAt(&lsaError{
			Msg: "Can't translateString, type not ltString."})
	}

	S := self.Lexem.LexemAsString()
	self.Lexem = self.Lexem.Next

	index := self.StrStrings.addUnique(S)
	item := TLanguageItem{Type: ltitString, Index: index}
	self.LanguageItems = append(self.LanguageItems, item)

	return nil
}

//TODO: распознание символа как аргумента
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
			if E := Self.translateComplexIdent(); E != nil {
				return Self.Lexem.errorAt(ESyntaxError)
			}
		}

	case ltString:
		{
			if E := Self.translateString(); E != nil {
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

func (self *TSyntaxDescriptor) translateOperation() error {
	var lit TLanguageItemType
	switch self.Lexem.Type {
	case ltStar:
		lit = ltitMathMul

	case ltPlus:
		lit = ltitMathAdd

	case ltMinus:
		lit = ltitMathSub

	case ltSlash:
		lit = ltitMathDiv
	}

	item := TLanguageItem{Type: lit}
	self.LanguageItems = append(self.LanguageItems, item)
	self.Lexem = self.Lexem.Next
	return nil

}

/*
BNF-определения для присваивания выражения переменной
<СЛОЖНЫЙ ИДЕНТИФИКАТОР> '=' <ВЫРАЖЕНИЕ>
СЛОЖНЫЙ ИДЕНТИФИКАТОР = <ИДЕНТИФИКАТОР> {' ' <ИДЕНТИФИКАТОР>}
ВЫРАЖЕНИЕ = <АРГУМЕНТ> {<ОПЕРАЦИЯ> <АРГУМЕНТ>}
СЛОЖНЫЙ АРГУМЕНТ = [<УНАРНАЯ ОПЕРАЦИЯ>] <АРГУМЕНТ>
ОПЕРАЦИЯ = '+' | '-' | '*' | '/' | '%' | '^'
УНАРНАЯ ОПЕРАЦИЯ = '!'
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
				E = Self.translateOperation()
				if E != nil {
					return E
				}

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
					syntaxDescriptor.translateFunctionDeclaration()
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
