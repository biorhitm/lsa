package lsa

import (
	"errors"
	"fmt"
	"github.com/biorhitm/memfs"
	"io"
	"unsafe"
)

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

const (
	LF = 0xA // LineFeed
)

type TLexem struct {
	Next     PLexem
	Text     memfs.PBigByteArray
	Size     uint
	Type     TLexemType
	LineNo   uint
	ColumnNo uint
}

type PLexem *TLexem

type TReader struct {
	Text      memfs.PBigByteArray
	Size      uint64
	Index     uint64
	NextIndex uint64
	LineNo    uint
	ColumnNo  uint
}

func (self *TLexem) LexemAsString() string {
	S := ""

	if self.Size > 0 && self.Text != nil {
		b := make([]uint8, self.Size)
		for i := uint(0); i < self.Size; i++ {
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

	if R.NextIndex > R.Index {
		R.ColumnNo++
		if R.Text[R.Index] == LF {
			R.ColumnNo = 0
			R.LineNo++
		}
		R.Index = R.NextIndex
	}

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

	R.NextIndex = R.Index + uint64(size)
	return aChar, nil
}

func (R *TReader) unread() {
	R.NextIndex = R.Index
}

func (R *TReader) createNewLexem(parent PLexem, _type TLexemType) (PLexem, error) {
	var startIndex uint64 = 0

	L := new(TLexem)
	L.Type = _type
	L.Next = nil
	L.Size = 0
	L.Text = nil
	L.LineNo = R.LineNo
	L.ColumnNo = R.ColumnNo

	switch _type {
	case ltIdent:
		{
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[R.Index]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						L.Size = uint(R.Index - startIndex)
						break
					}
					return nil, err
				}
				if !isIdentLetter(C) && !isDigit(C) {
					R.unread()
					L.Size = uint(R.Index - startIndex)
					break
				}
			}
		}

	case ltNumber:
		{
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						L.Size = uint(R.Index - startIndex)
						break
					}
					return nil, err
				}
				if !isDigit(C) {
					R.unread()
					L.Size = uint(R.Index - startIndex)
					break
				}
			}
		}

	case ltString:
		{
			startIndex = R.NextIndex
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						err = fmt.Errorf("[%d:%d] Незакрытая строка",
							L.LineNo, L.ColumnNo)
					}
					return nil, err
				}
				if C == '"' {
					L.Size = uint(R.Index - startIndex)
					break
				}
			}
		}

	case ltChar:
		{
			startIndex = R.NextIndex
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[startIndex]))
			//TODO заменить на вызов чтения символа, разрулить спец. послед.
			for {
				C, err := R.readRune()
				if err != nil {
					if err == io.EOF {
						//TODO выдать ошибку unterminated char
						break
					}
					return nil, err
				}
				if C == 0x27 {
					L.Size = uint(R.Index - startIndex)
					break
				}
			}
		}
	}

	if parent != nil {
		parent.Next = L
	}
	return L, nil
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
*/
func (R *TReader) BuildLexems() (PLexem, error) {
	var curLexem, firstLexem PLexem

	curLexem = new(TLexem)
	firstLexem = curLexem

	for {
		C, err := R.readRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch {
		case isIdentLetter(C):
			{
				R.unread()
				curLexem, err = R.createNewLexem(curLexem, ltIdent)
			}

		case isDigit(C):
			{
				R.unread()
				curLexem, err = R.createNewLexem(curLexem, ltNumber)
			}

		case C == '"':
			{
				curLexem, err = R.createNewLexem(curLexem, ltString)
			}

		case C == 0x27: //single quote
			{
				curLexem, err = R.createNewLexem(curLexem, ltChar)
			}

		case isSymbol(C):
			{
				// код символа будет типом лексемы
				curLexem, err = R.createNewLexem(curLexem, TLexemType(C))
			}

		case C == 0x0A:
			{
				curLexem, err = R.createNewLexem(curLexem, ltEOL)
			}

		default:
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	R.createNewLexem(curLexem, ltEOF)

	return firstLexem.Next, nil
}
