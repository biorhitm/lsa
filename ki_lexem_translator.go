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
	ltitEqual
	ltitAbove
	ltitBelow
	ltitAboveEqual
	ltitBelowEqual
	ltitLeftShift
	ltitRightShift
	ltitFunction
	ltitVarList
	ltitDataType
	ltitBegin
	ltitEnd
	ltitClassMember
	ltitParameters
	ltitPackageName
	ltitIf
	ltitElse
	ltitWhile
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
	Keyword       TKeywordId
}

type TKeywordId uint
type TKeyword struct {
	Id   TKeywordId
	Name string
}

// KeywordsIds
const (
	kwiUnknown = TKeywordId(iota)
	kwiFunction
	kwiVariable
	kwiBegin
	kwiEnd
	kwiIf
	kwiElse
	kwiWhile
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
		TKeyword{kwiIf, "если"},
		TKeyword{kwiIf, "if"},
		TKeyword{kwiElse, "иначе"},
		TKeyword{kwiElse, "else"},
		TKeyword{kwiWhile, "пока"},
		TKeyword{kwiWhile, "while"},
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
	EExpectedCloseOper  = &lsaError{Msg: "Отсутствует 'конец'"}
	EUnExpectedKeyword  = &lsaError{Msg: "Встретилось зарезервированное слово"}
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

// Извлекает из лексем идентификатор, состоящий из нескольких слов
// Возвращает: (1)ошибку не nil, если первая лексема не ltIdent или первое
//  слово — зарезервированное слово; (2)строку — идентификатор, состоящий из
//  нескольких слов, разделённых пробелом(кол-во пробелов между словами не
//  различается; (3)номер зарезервированного слова, если не равен kwiUnknown,
//  значит такое слово встретилось
func (self *TSyntaxDescriptor) ExtractComplexIdent() (error, string,
	TKeywordId) {
	if self.Lexem.Type != ltIdent {
		return self.Lexem.errorAt(ESyntaxError), "", kwiUnknown
	}
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	if kId != kwiUnknown {
		return nil, "", kId
	}
	kId = kwiUnknown
	self.NextLexem()
	res := S
	for self.Lexem.Type == ltIdent {
		S := self.Lexem.LexemAsString()
		kId := toKeywordId(S)
		if kId != kwiUnknown {
			return nil, res, kId
		}
		res += " " + S
		self.NextLexem()
	}
	return nil, res, kwiUnknown
}

func toKeywordId(S string) TKeywordId {
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
		E    error
		kId  TKeywordId = kwiUnknown
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
				E = self.translateComplexIdent()
				if E != nil {
					return self.Lexem.errorAt(&lsaError{
						Msg: E.Error() + ". Отсутствует имя параметра"})
				}
				if self.Lexem.Type != ltComma {
					break
				}
				self.NextLexem()
			}

			if self.Lexem.Type != ltColon {
				return self.Lexem.errorAt(&lsaError{
					Msg: "Не указан тип параметра"})
			}
			self.NextLexem()

			//ТИП = [<ИМЯ ПАКЕТА> '.']<ИДЕНТИФИКАТОР>
			E, name, kId = self.ExtractComplexIdent()
			if E != nil || kId != kwiUnknown {
				//TODO: Тип может быть 'array ...'
				return self.Lexem.errorAt(&lsaError{
					Msg: ". Ожидается тип"})
			}
			self.AppendItem(ltitDataType)
			if self.Lexem.Type == ltDot {
				self.NextLexem()
				self.AppendItem(ltitPackageName)
				self.AppendIdent(name)
				E, name, kId = self.ExtractComplexIdent()
				if E != nil || kId != kwiUnknown {
					return self.Lexem.errorAt(&lsaError{
						Msg: ". Ожидается тип"})
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
		E, name, kId = self.ExtractComplexIdent()
		if E != nil {
			return self.Lexem.errorAt(&lsaError{
				Msg: ". Ожидается тип"})
		}
		self.AppendItem(ltitDataType)
		if self.Lexem.Type == ltDot {
			self.NextLexem()
			self.AppendItem(ltitPackageName)
			self.AppendIdent(name)
			E, name, kId = self.ExtractComplexIdent()
			if E != nil || kId != kwiUnknown {
				if E != nil {
					return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип"})
				}
			}
		}
		self.AppendIdent(name)
	}

	return nil
}

func (self *TSyntaxDescriptor) translateVarList() error {
	var (
		S, name string
		E       error
		keywId  TKeywordId
	)

	S = self.Lexem.LexemAsString()
	keywId = toKeywordId(S)
	if keywId == kwiVariable {
		self.AppendItem(ltitVarList)
		self.NextLexem()

		typeNotPresent := true
		for {
			E, name, keywId = self.ExtractComplexIdent()
			if E != nil || keywId != kwiUnknown {
				if E != nil {
					return self.Lexem.errorAt(&lsaError{Msg: "Ожидается имя переменной"})
				}
			}
			self.AppendIdent(name)

			if self.Lexem.Type == ltColon {
				self.NextLexem()
				E, name, keywId = self.ExtractComplexIdent()
				if E != nil || keywId != kwiUnknown {
					if E != nil {
						return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип переменной"})
					}
				}
				self.AppendItem(ltitDataType)
				if self.Lexem.Type == ltDot {
					self.NextLexem()
					self.AppendItem(ltitPackageName)
					self.AppendIdent(name)
					E, name, keywId = self.ExtractComplexIdent()
					if E != nil || keywId != kwiUnknown {
						if E != nil {
							return self.Lexem.errorAt(&lsaError{Msg: "Ожидается тип переменной"})
						}
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
		E       error
		keywId  TKeywordId
	)

	S = self.Lexem.LexemAsString()
	keywId = toKeywordId(S)
	if keywId != kwiFunction {
		return self.Lexem.errorAt(&lsaError{Msg: "Can't translateFunctionDeclaration, type not function."})
	}

	self.AppendItem(ltitFunction)
	self.NextLexem()

	// [<ИМЯ КЛАССА> '.']<ИДЕНТИФИКАТОР> '('
	E, name, keywId = self.ExtractComplexIdent()
	if E != nil || keywId != kwiUnknown {
		if E != nil {
			return self.Lexem.errorAt(&lsaError{Msg: "Ожидается идентификатор"})
		}
	}

	if self.Lexem.Type == ltDot { //функция является членом класса
		self.NextLexem()
		self.AppendItem(ltitClassMember)
		self.AppendIdent(name)
		E, name, keywId = self.ExtractComplexIdent()
		if E != nil || keywId != kwiUnknown {
			return self.Lexem.errorAt(&lsaError{Msg: "Ожидается идентификатор"})
		}
	}
	self.AppendIdent(name)

	if E = self.translateFunctionPrototype(); E != nil {
		return E
	}

	if self.Lexem.Type == ltSemicolon {
		self.NextLexem()
	}

	// читаю список локальных переменных
	if E := self.translateVarList(); E != nil {
		return E
	}
	if E := self.translateGroupOfStatements(); E != nil {
		return E
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
	self.AppendNumber(S)
	self.NextLexem()
	return nil
}

func (self *TSyntaxDescriptor) translateComplexIdent() error {
	self.Keyword = kwiUnknown
	if self.Lexem.Type != ltIdent {
		return self.Lexem.errorAt(&lsaError{Msg: "Can't translateComplexIdent, type not ltIdent."})
	}

	S := self.Lexem.LexemAsString()
	K := toKeywordId(S)
	if K != kwiUnknown {
		self.Keyword = K
		return self.Lexem.errorAt(&lsaError{Msg: "Can't translateComplexIdent, keyword."})
	}
	self.NextLexem()

	ident := S
	for self.Lexem.Type == ltIdent {
		S = self.Lexem.LexemAsString()
		K = toKeywordId(S)
		if K != kwiUnknown {
			self.Keyword = K
			break
		}

		ident += " " + S
		self.NextLexem()
	}

	self.AppendIdent(ident)
	return nil
}

func (self *TSyntaxDescriptor) translateString() error {
	if self.Lexem.Type != ltString {
		return self.Lexem.errorAt(&lsaError{
			Msg: "Can't translateString, type not ltString."})
	}

	S := self.Lexem.LexemAsString()
	self.NextLexem()

	self.AppendString(S)

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
func (Self *TSyntaxDescriptor) translateArgument() (E error) {
	E = nil
	var (
		S string
	)

	// пропускаю необязательные открывающие скобки
	for Self.Lexem.Type == ltOpenParenthesis {
		Self.NextLexem()
		Self.AppendItem(ltitOpenParenthesis)
		Self.Parenthesis++
	}

	switch Self.Lexem.Type {
	case ltNumber:
		S = Self.Lexem.LexemAsString()
		Self.AppendNumber(S)
		Self.NextLexem()

	case ltIdent:
		E = Self.translateComplexIdent()

	case ltString:
		E = Self.translateString()

	default:
		return Self.Lexem.errorAt(EExpectedArgument)
	}

	// пропускаю необязательные закрывающие скобки
	for Self.Lexem.Type == ltCloseParenthesis {
		Self.NextLexem()
		Self.AppendItem(ltitCloseParenthesis)
		Self.Parenthesis--
	}
	return
}

//TODO: добавить проверку остальных операций: ! ~ & | and or xor not shr shl
func (self *TSyntaxDescriptor) translateOperation() error {
	var curT, nextT TLexemType
	var lit TLanguageItemType

	curT = self.Lexem.Type
	nextT = ltUnknown
	if self.Lexem.Next != nil {
		nextT = self.Lexem.Next.Type
	}

	switch curT {
	case ltStar:
		lit = ltitMathMul
	case ltPlus:
		lit = ltitMathAdd
	case ltMinus:
		lit = ltitMathSub
	case ltSlash:
		lit = ltitMathDiv
	case ltEqualSign:
		lit = ltitEqual
	case ltAboveSign:
		lit = ltitAbove
		if nextT == ltEqualSign {
			self.NextLexem()
			lit = ltitAboveEqual
		} else if nextT == ltAboveSign {
			self.NextLexem()
			lit = ltitRightShift
		}

	case ltBelowSign:
		lit = ltitBelow
		if nextT == ltEqualSign {
			self.NextLexem()
			lit = ltitBelowEqual
		} else if nextT == ltBelowSign {
			self.NextLexem()
			lit = ltitLeftShift
		}

	default:
		return self.Lexem.errorAt(ESyntaxError)
	}

	self.AppendItem(lit)
	self.NextLexem()
	return nil

}

func (self *TSyntaxDescriptor) translateExpression() (E error) {
	self.Parenthesis = 0

	// обрабатываю аргумент
	E = self.translateArgument()

	// [<ОПЕРАЦИЯ><АРГУМЕНТ>]
	for E == nil {
		E = self.translateOperation()
		if E != nil {
			E = nil
			break
		}
		E = self.translateArgument()
	}

	if E == nil {
		if self.Parenthesis < 0 {
			E = self.Lexem.errorAt(ETooMuchCloseRB)
		}
		if self.Parenthesis > 0 {
			E = self.Lexem.errorAt(ETooMuchOpenRB)
		}
	}

	return
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
func (Self *TSyntaxDescriptor) translateAssignment() (E error) {
	E = Self.translateComplexIdent()
	if E != nil {
		return Self.Lexem.errorAt(ESyntaxError)
	}

	if Self.Lexem.Type == ltEqualSign {
		Self.AppendItem(ltitAssignment)
		Self.NextLexem() // пропускаю знак =
	} else {
		return Self.Lexem.errorAt(ESyntaxError)
	}

	if Self.Lexem.Type == ltEOF {
		return Self.Lexem.errorAt(EExpectedExpression)
	}
	Self.Lexem = Self.Lexem.skipEOL()

	E = Self.translateExpression()

	return
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

/*
Анализ группы операторов в программных скобках '{' '}'
*/
func (self *TSyntaxDescriptor) translateGroupOfStatements() (E error) {
	E = nil

	var (
		wasBegin bool
		wasEnd   bool
	)

	wasBegin = self.Lexem.Type == ltLBrace
	if !wasBegin && self.Lexem.Type == ltIdent {
		S := self.Lexem.LexemAsString()
		kId := toKeywordId(S)
		wasBegin = kId == kwiBegin
	}

	if !wasBegin {
		return self.translateLexem()
	}

	self.begin()

Loop:
	for {
		switch self.Lexem.Type {
		case ltRBrace:
			wasEnd = true

		case ltEOF:
			return self.Lexem.errorAt(EExpectedCloseOper)

		case ltIdent:
			S := self.Lexem.LexemAsString()
			kId := toKeywordId(S)
			wasEnd = kId == kwiEnd
		}

		if wasEnd {
			self.end()
			break Loop
		}

		E = self.translateLexem()
	}

	return
}

/*
BNF-определения для оператора 'если'
ОПЕРАТОР ЕСЛИ = <ЕСЛИ> <ВЫРАЖЕНИЕ> <ВЕТКА> [<ИНАЧЕ> <ВЕТКА>]
ЕСЛИ = 'если' | 'if'
ИНАЧЕ = 'иначе' | 'else'
ВЕТКА = <НАЧАЛО> <ОПЕРАТОРЫ> <КОНЕЦ>
*/
func (self *TSyntaxDescriptor) translateIfStatement() (E error) {
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	if kId != kwiIf {
		return self.Lexem.errorAt(ESyntaxError)
	}
	self.NextLexem()
	self.AppendItem(ltitIf)

	if E = self.translateExpression(); E != nil {
		return
	}
	if E = self.translateGroupOfStatements(); E != nil {
		return
	}
	S = self.Lexem.LexemAsString()
	kId = toKeywordId(S)
	if kId == kwiElse {
		self.NextLexem()
		self.AppendItem(ltitElse)
		E = self.translateGroupOfStatements()
	}

	return
}

func (self *TSyntaxDescriptor) translateWhileStatement() (E error) {
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	if kId != kwiWhile {
		return self.Lexem.errorAt(ESyntaxError)
	}
	self.NextLexem()
	self.AppendItem(ltitWhile)

	if E = self.translateExpression(); E != nil {
		return
	}
	if E = self.translateGroupOfStatements(); E != nil {
		return
	}

	return
}

func (self *TSyntaxDescriptor) translateIdent() (E error) {
	E = nil
	S := self.Lexem.LexemAsString()
	kId := toKeywordId(S)
	switch kId {
	case kwiVariable:
		E = self.translateVarList()

	case kwiFunction:
		E = self.translateFunctionDeclaration()

	case kwiBegin:
		self.begin()

	case kwiEnd:
		self.end()

	case kwiIf:
		E = self.translateIfStatement()

	case kwiWhile:
		E = self.translateWhileStatement()

	default:
		self.NextLexem()
	}

	return
}

/*
Анализ лексемы, когда нет активного оператора
Например, после for должна быть инициализация переменной цикла, 'to',
<ВЫРАЖЕНИЕ>, возможно шаг,
возможно 'begin', далее идёт всё что угодно, вот тут translateLexem и нужен
*/
func (self *TSyntaxDescriptor) translateLexem() (E error) {
	E = nil
	switch self.Lexem.Type {
	case ltIdent:
		E = self.translateIdent()

	case ltEqualSign:
		self.Lexem = self.StartLexem
		E = self.translateAssignment()

	case ltEOL:
		self.NextLexem()

	case ltLBrace:
		self.begin()

	case ltRBrace:
		self.end()

	default:
		if self.Lexem.Size > 0 {
			fmt.Printf("Лехема: %d size: %d %s ",
				self.Lexem.Type, self.Lexem.Size, self.Lexem.LexemAsString())
		}
		self.NextLexem()
	}

	return
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

	for sd.Lexem != nil && sd.Lexem.Type != ltEOF {
		if E := sd.translateLexem(); E != nil {
			return TSyntaxDescriptor{}, E
		}
	}

	return sd, nil
}
