package lsa

import (
	"fmt"
)

type TLanguageItemType uint

//TLanguageItemType типы синтаксичекских элементов
const (
	ltitUnknown TLanguageItemType = iota
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
	ltitFunction
	ltitVarList
	ltitDataType
	ltitBegin
	ltitEnd
	ltitClassMember
	ltitParameters
	ltitPackageName
)

type TLanguageItem struct {
	Type TLanguageItemType
	// идентификаторы будут держать здесь номер строки из массива
	// всех идентификаторов
	Index uint
}

type TStringArray []string

type TSyntaxDescriptor struct {
	Lexem *TLexem
	// лексема к которой надо будет вернуться, чтобы продолжить перевод
	// например, если название переменной состоит из нескольких слов, то
	// после нахождения знака =, надо будет вернуться к первому слову
	// переменной, чтобы записать её полное название
	StartLexem    *TLexem
	LanguageItems []TLanguageItem
	Parenthesis   int
	BeginCount    int
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
	Id   uint
	Name string
}

// KeywordsIds
const (
	kwiUnknown = iota
	kwiFunction
	kwiVariable
	kwiBegin
	kwiEnd
)

var (
	keywordList = []TKeyword{
		TKeyword{kwiFunction, "функция"},
		TKeyword{kwiFunction, "процедура"},
		TKeyword{kwiFunction, "function"},
		TKeyword{kwiFunction, "func"},
		TKeyword{kwiFunction, "def"},
		TKeyword{kwiVariable, "переменные"},
		TKeyword{kwiVariable, "var"},
		TKeyword{kwiBegin, "начало"},
		TKeyword{kwiBegin, "begin"},
		TKeyword{kwiEnd, "конец"},
		TKeyword{kwiEnd, "end"},
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

func (self *TSyntaxDescriptor) AppendItem(AType TLanguageItemType) {
	item := TLanguageItem{Type: AType}
	self.LanguageItems = append(self.LanguageItems, item)
}

func (self *TSyntaxDescriptor) AppendIdent(AName string) {
	index := self.StrIdents.addUnique(AName)
	item := TLanguageItem{Type: ltitIdent, Index: index}
	self.LanguageItems = append(self.LanguageItems, item)
}

func (self *TSyntaxDescriptor) AppendNumber(S string) {
	index := self.StrNumbers.addUnique(S)
	item := TLanguageItem{Type: ltitNumber, Index: index}
	self.LanguageItems = append(self.LanguageItems, item)
}

func (self *TSyntaxDescriptor) AppendString(S string) {
	index := self.StrStrings.addUnique(S)
	item := TLanguageItem{Type: ltitString, Index: index}
	self.LanguageItems = append(self.LanguageItems, item)
}

func (self *TSyntaxDescriptor) NextLexem() {
	if self.Lexem != nil && self.Lexem.Type != ltEOF {
		self.Lexem = self.Lexem.Next
	}
}

//TODO: должен возвращать ошибку 'встретилось зарезервированное слово' с
// кодом слова
func (self *TSyntaxDescriptor) ExtractComplexIdent() (string, bool) {
	if self.Lexem.Type != ltIdent {
		return "", false
	}
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	if kId != kwiUnknown {
		return "Встретилось зарезервированное слово", false
	}
	self.NextLexem()
	res := S
	for self.Lexem.Type == ltIdent {
		S := self.Lexem.LexemAsString()
		kId := toKeywordId(S)
		if kId != kwiUnknown {
			return res, true
		}
		res += " " + S
		self.NextLexem()
	}
	return res, true
}

func toKeywordId(S string) uint {
	for i := 0; i < len(keywordList); i++ {
		if S == keywordList[i].Name {
			return keywordList[i].Id
		}
	}
	return kwiUnknown
}

/*
BNF-правила для прототипа функции
ПРОТОТИП ФУНКЦИИ = [<ПАРАМЕТРЫ>] [<РЕЗУЛЬТАТ>]
ПАРАМЕТРЫ = '(' <ТИПИЗИРОВАННЫЕ ПАРАМЕТРЫ> {',' <ТИПИЗИРОВАННЫЕ ПАРАМЕТРЫ>} ')'
ТИПИЗИРОВАННЫЕ ПАРАМЕТРЫ = <СПИСОК ИМЁН> ':' <ТИП>
СПИСОК ИМЁН = <ИМЯ> {',' <ИМЯ>}
РЕЗУЛЬТАТ = ':' [<ИМЯ ПАКЕТА> '.']<ИДЕНТИФИКАТОР>
*/
func (self *TSyntaxDescriptor) translateFunctionPrototype() error {
	var (
		name string
		ok   bool
	)

	//[<ПАРАМЕТРЫ>]
	if self.Lexem.Type == ltOpenParenthesis {
		self.NextLexem()

		if self.Lexem.Type != ltCloseParenthesis {
			self.AppendItem(ltitParameters)
		}

		//<ТИПИЗИРОВАННЫЕ ПАРАМЕТРЫ> {',' <ТИПИЗИРОВАННЫЕ ПАРАМЕТРЫ>} ')'
		for self.Lexem.Type != ltCloseParenthesis {
			//СПИСОК ИМЁН = <ИМЯ> {',' <ИМЯ>}
			for {
				name, ok = self.ExtractComplexIdent()
				if !ok {
					return self.Lexem.errorAt(&lsaError{Msg: "Отсутствует имя параметра"})
				}
				self.AppendIdent(name)
				if self.Lexem.Type != ltComma {
					break
				}
				self.NextLexem()
			}

			if self.Lexem.Type != ltColon {
				return self.Lexem.errorAt(&lsaError{Msg: "Не указан тип параметра"})
			}
			self.NextLexem()

			//ТИП = [<ИМЯ ПАКЕТА> '.']<ИДЕНТИФИКАТОР>
			if name, ok = self.ExtractComplexIdent(); !ok {
				return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип"})
			}
			self.AppendItem(ltitDataType)
			if self.Lexem.Type == ltDot {
				self.NextLexem()
				self.AppendItem(ltitPackageName)
				self.AppendIdent(name)
				if name, ok = self.ExtractComplexIdent(); !ok {
					return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип"})
				}
			}
			self.AppendIdent(name)
			if self.Lexem.Type != ltComma {
				break
			}
			self.NextLexem()
		}

		if self.Lexem.Type != ltCloseParenthesis {
			return self.Lexem.errorAt(&lsaError{Msg: "Ожидается ')'"})
		}
		self.NextLexem()
	}

	//[<РЕЗУЛЬТАТ>]
	//РЕЗУЛЬТАТ = ':' [<ИМЯ ПАКЕТА> '.']<ИМЯ ТИПА>
	if self.Lexem.Type == ltColon {
		self.NextLexem()
		if name, ok = self.ExtractComplexIdent(); !ok {
			return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип"})
		}
		self.AppendItem(ltitDataType)
		if self.Lexem.Type == ltDot {
			self.NextLexem()
			self.AppendItem(ltitPackageName)
			self.AppendIdent(name)
			if name, ok = self.ExtractComplexIdent(); !ok {
				return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип"})
			}
		}
		self.AppendIdent(name)
	}

	return nil
}

/*
BNF-правила для объявления функции
ОБЪЯВЛЕНИЕ ФУНКЦИИ = <ФУНКЦИЯ> <ИМЯ ФУНКЦИИ> [<ПРОТОТИП>]
  [<ЛОКАЛЬНЫЕ ПЕРЕМЕННЫЕ>] <НАЧАЛО> [<ТЕЛО ФУНКЦИИ>] <КОНЕЦ>
ФУНКЦИЯ = 'функция' | 'function' | 'func' | 'def'
НАЧАЛО = 'начало' | 'begin' | '{'
КОНЕЦ = 'конец' | 'end' | '}'
ИМЯ ФУНКЦИИ = [<ИМЯ КЛАССА> '.']<ИДЕНТИФИКАТОР>
ИМЯ КЛАССА = <ИДЕНТИФИКАТОР>
ТИП = [<ИМЯ ПАКЕТА> '.']<ИДЕНТИФИКАТОР>
ИМЯ ПАКЕТА = <ИДЕНТИФИКАТОР>
<ЛОКАЛЬНЫЕ ПЕРЕМЕННЫЕ> = ('переменные' | 'var') <СПИСОК ПЕРЕМЕННЫХ>
СПИСОК ПЕРЕМЕННЫХ = <ПЕРЕМЕННАЯ> {(';' | <LF>) <ПЕРЕМЕННАЯ>}
ПЕРЕМЕННАЯ = <ИМЯ ПЕРЕМЕННОЙ> ':' <ТИП>
ИМЯ ПЕРЕМЕННОЙ = <ИДЕНТИФИКАТОР>
*/
func (self *TSyntaxDescriptor) translateFunctionDeclaration() error {
	var (
		S, name string
		ok      bool
		keywId  uint
	)

	S = self.Lexem.LexemAsString()
	keywId = toKeywordId(S)
	if keywId != kwiFunction {
		return self.Lexem.errorAt(&lsaError{Msg: "Can't translateFunctionDeclaration, type not function."})
	}

	self.AppendItem(ltitFunction)
	self.NextLexem()

	// [<ИМЯ КЛАССА> '.']<ИДЕНТИФИКАТОР> '('
	if name, ok = self.ExtractComplexIdent(); !ok {
		return self.Lexem.errorAt(&lsaError{Msg: "Ожидается идентификатор"})
	}

	if self.Lexem.Type == ltDot { //функция является членом класса
		self.NextLexem()
		self.AppendItem(ltitClassMember)
		self.AppendIdent(name)
		if name, ok = self.ExtractComplexIdent(); !ok {
			return self.Lexem.errorAt(&lsaError{Msg: "Ожидается идентификатор"})
		}
	}
	self.AppendIdent(name)

	E := self.translateFunctionPrototype()
	if E != nil {
		return E
	}

	if self.Lexem.Type == ltSemicolon {
		self.NextLexem()
	}

	// читаю список локальных переменных
	S = self.Lexem.LexemAsString()
	keywId = toKeywordId(S)
	if keywId == kwiVariable {
		self.AppendItem(ltitVarList)
		self.NextLexem()

		typeNotPresent := true
		for {
			if name, ok = self.ExtractComplexIdent(); !ok {
				return self.Lexem.errorAt(&lsaError{Msg: "Ожидается имя переменной"})
			}
			self.AppendIdent(name)

			if self.Lexem.Type == ltColon {
				self.NextLexem()
				if name, ok = self.ExtractComplexIdent(); !ok {
					return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип переменной"})
				}
				self.AppendItem(ltitDataType)
				if self.Lexem.Type == ltDot {
					self.NextLexem()
					self.AppendItem(ltitPackageName)
					self.AppendIdent(name)
					if name, ok = self.ExtractComplexIdent(); !ok {
						return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип переменной"})
					}
					typeNotPresent = false
				}
				self.AppendIdent(name)
				typeNotPresent = false
			}
			//если после параметра нет ',', значит список кончился, жду 'начало'
			if self.Lexem.Type != ltComma {
				break
			}
			self.NextLexem()
			typeNotPresent = false
		}
		if typeNotPresent {
			return self.Lexem.errorAt(&lsaError{Msg: "Не указан тип параметра"})
		}
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

func (self *TSyntaxDescriptor) begin() {
	self.BeginCount++
	self.AppendItem(ltitBegin)
	self.NextLexem()
}

func (self *TSyntaxDescriptor) end() {
	self.BeginCount--
	self.AppendItem(ltitEnd)
	self.NextLexem()
}

func (self *TSyntaxDescriptor) translateIdent() error {
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	switch kId {
	case kwiFunction:
		{
			if E := self.translateFunctionDeclaration(); E != nil {
				return E
			}
		}

	case kwiBegin:
		{
			self.begin()
		}

	case kwiEnd:
		{
			self.end()
		}

	default:
		{
			self.NextLexem()
		}
	}

	return nil
}

/*
 Переводит текст в лексемах в массив элементов языка
*/
func TranslateCode(ALexem PLexem) (TSyntaxDescriptor, error) {
	sd := TSyntaxDescriptor{
		Lexem:         ALexem,
		StartLexem:    ALexem,
		LanguageItems: make([]TLanguageItem, 0, 1000),
		Parenthesis:   0,
		BeginCount:    0,
		StrNumbers:    make([]string, 0, 1024),
		StrIdents:     make([]string, 0, 1024),
		StrStrings:    make([]string, 0, 1024),
	}

	var (
		E error = nil
	)

	for sd.Lexem != nil && sd.Lexem.Type != ltEOF {
		switch sd.Lexem.Type {
		case ltIdent:
			{
				if E = sd.translateIdent(); E != nil {
					return TSyntaxDescriptor{}, E
				}
			}

		case ltEqualSign:
			{
				sd.Lexem = sd.StartLexem
				E = sd.translateAssignment()
				if E != nil {
					return TSyntaxDescriptor{}, E
				}
			}

		case ltEOL:
			{
				sd.NextLexem()
			}

		case ltLBrace:
			{
				sd.begin()
			}

		case ltRBrace:
			{
				sd.end()
			}

		default:
			if sd.Lexem.Size > 0 {
				fmt.Printf("Лехема: %d size: %d %s ",
					sd.Lexem.Type, sd.Lexem.Size, sd.Lexem.LexemAsString())
			}
			sd.NextLexem()
		}
	}

	return sd, nil
}
