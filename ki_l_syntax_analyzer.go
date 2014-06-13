package lsa

import (
	"errors"
	"github.com/biorhitm/memfs"
	"io"
	"unsafe"
)

type TBigShortArray [0x1FFFFFFFFFFFF]uint16
type PBigShortArray *TBigShortArray

type TLexemType uint

// TLexemType
const (
	ltUnknown = iota
	ltEOF
	ltNumber // 12
	ltString // "test"
	ltChar   // 'a' 'x' '%'
	ltIdent  // имя функции, переменной или типа
	ltEOL
	ltExclamationMark   = '!'
	ltQuote             = '"'
	ltSharp             = '#'
	ltDollar            = '$'
	ltPercent           = '%'
	ltAmpersand         = '&'
	ltSingleQuote       = 39
	ltOpenParenthesis   = '('
	ltCloseParenthesis  = ')'
	ltStar              = '*'
	ltPlus              = '+'
	ltComma             = ','
	ltMinus             = '-'
	ltDot               = '.'
	ltSlash             = '/'
	ltColon             = ':'
	ltSemicolon         = ';'
	ltBelowSign         = '<'
	ltEqualSign         = '='
	ltAboveSign         = '>'
	ltQuestionMark      = '?'
	ltAt                = '@'
	ltOpenBracket       = '['
	ltBackSlash         = 92
	ltCloseBracket      = ']'
	ltinvolution        = '^'
	ltBackQuote         = '`'
	ltOpenShapeBracket  = '{'
	ltVerticalLine      = '|'
	ltCloseShapeBracket = '}'
	ltTilde             = '~'
)

type TLexem struct {
	Next PLexem
	Text memfs.PBigByteArray
	Size int
	Type TLexemType
}

type PLexem *TLexem

type TReader struct {
	Text      memfs.PBigByteArray
	Size      uint64
	Index     uint64
	PrevIndex uint64
}

func (self *TLexem) LexemAsString() string {
	S := ""

	if self.Size > 0 {
		b := make([]uint8, self.Size)
		for i := 0; i < self.Size; i++ {
			b[i] = self.Text[i]
		}
		S = string(b)
	}

	return S
}

func isLetter(C rune) bool {
	return (0x0410 <= C && C <= 0x044F) || C == 0x0401 || C == 0x0451
}

func isIdentLetter(C rune) bool {
	return C == '_' || isLetter(C)
}

func isDigit(C rune) bool {
	return '0' <= C && C <= '9'
}

/*
 Возвращает истину, если аргумент является символом из списка:
   ! " # $ % & ' ( ) * + , - . /  : ; < = > ? @  [ \ ] ^  `    { | } ~
   33...                   ...47, 58...   ...64, 91...94, 96, 123...126
 Тоже самое в числах:

*/
func isSymbol(C rune) bool {
	return (33 <= C && C <= 47) ||
		(58 <= C && C <= 64) ||
		(91 <= C && C <= 94) ||
		C == 96 ||
		(123 <= C && C <= 126)
}

var InvalidRune = errors.New("Invalid utf8 char, support russian only")

func (R *TReader) readRune() (aChar rune, E error) {

	if R.Index >= R.Size {
		return 0, io.EOF
	}

	size := uint(1)
	ok := false

	for !ok {
		B := R.Text[R.Index]

		if B&0x80 != 0 { //first or next byte utf-8 sequence longer than 1 byte
			B <<= 1
			sequenceLen := uint(1)
			for ; B&0x80 != 0; B <<= 1 {
				sequenceLen++
			}

			if sequenceLen < 2 {
				// курсор указывает не на начало последовательности
				R.Index++
				if R.Index >= R.Size {
					return 0, io.EOF
				}
				continue
			}
			if sequenceLen > 4 {
				// курсор указывает на неправильную последовательность
				R.Index++
				if R.Index >= R.Size {
					return 0, io.EOF
				}
				continue
			}
			if R.Index+uint64(sequenceLen)-1 >= R.Size {
				// последовательность не помещается в буфер
				return 0, io.EOF
			}

			aChar = rune(B >> sequenceLen)
			ok = true

			for i := uint(1); i < sequenceLen; i++ {
				B = R.Text[R.Index+uint64(i)]

				// если в старших двух битах B число 2(10xxx xxxx), то это
				// продолжение последовательности, иначе неправильная
				// последовательность
				if B&0xC0 == 0x80 {
					aChar = aChar<<6 | rune(B&0x3F)
				} else {
					R.Index++
					if R.Index >= R.Size {
						return 0, io.EOF
					}
					ok = false
					break
				}
			}
			if ok {
				size = sequenceLen
			}

		} else {
			aChar = rune(B)
			ok = true
		}
	}

	R.PrevIndex = R.Index
	R.Index += uint64(size)
	return aChar, nil
}

func (R *TReader) Unread() {
	R.Index = R.PrevIndex
}

func (R *TReader) createNewLexem(parent PLexem, _type TLexemType) PLexem {
	var startIndex uint64 = 0

	L := new(TLexem)
	L.Type = _type
	L.Next = nil
	L.Size = 0
	L.Text = nil

	switch _type {
	case ltIdent:
		{
			R.Unread()
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[R.Index]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						L.Size = int(R.Index - startIndex)
						break
					}
					return nil //TODO выдать ошибку
				}
				if !isIdentLetter(C) && !isDigit(C) {
					R.Unread()
					L.Size = int(R.Index - startIndex)
					break
				}
			}
		}

	case ltNumber:
		{
			startIndex = R.PrevIndex
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						L.Size = int(R.Index - startIndex)
						break
					}
					return nil //TODO выдать ошибку
				}
				if !isDigit(C) {
					R.Unread()
					L.Size = int(R.Index - startIndex)
					break
				}
			}
		}

	case ltString:
		{
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						//TODO выдать ошибку unterminated string
						break
					}
					return nil //TODO выдать ошибку
				}
				if C == '"' {
					L.Size = int(R.PrevIndex - startIndex)
					break
				}
			}
		}

	case ltChar:
		{
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			//TODO заменить на вызов чтения символа, разрулить спец. послед.
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						//TODO выдать ошибку unterminated char
						break
					}
					return nil //TODO выдать ошибку
				}
				if C == 0x27 {
					L.Size = int(R.PrevIndex - startIndex)
					break
				}
			}
		}
	}

	if parent != nil {
		parent.Next = L
	}
	return L
}

// Error codes
const (
	errNoSuccess = iota
	errNoUnterminatedString
	errNoUnterminatedChar
)

/* TODO: 1. идущие подряд символы переводы строк, интерпретировать как один
         если следующая строка состоит только из пробельных символов, то
		её тоже не включать в список лексем
		 2. Для символьных лексем сделать отдельные типы: assignment, dollar и т.д.
*/
func (R *TReader) BuildLexems() (lexem PLexem, errorCode uint, errorIndex uint64) {
	var curLexem, firstLexem PLexem

	curLexem = new(TLexem)
	firstLexem = curLexem

	for {
		C, err := R.readRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, 1, R.Index
		}

		switch {
		case isIdentLetter(C):
			{
				curLexem = R.createNewLexem(curLexem, ltIdent)
			}

		case isDigit(C):
			{
				curLexem = R.createNewLexem(curLexem, ltNumber)
			}

		case C == '"':
			{
				curLexem = R.createNewLexem(curLexem, ltString)
			}

		case C == 0x27: //single quote
			{
				curLexem = R.createNewLexem(curLexem, ltChar)
			}

		case isSymbol(C):
			{
				// код символа будет типом лексемы
				curLexem = R.createNewLexem(curLexem, TLexemType(C))
			}

		case C == 0x0A:
			{
				curLexem = R.createNewLexem(curLexem, ltEOL)
			}

		default:
		}
	}

	R.createNewLexem(curLexem, ltEOF)

	return firstLexem, errNoSuccess, 0
}
