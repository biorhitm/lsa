package lsa

import (
	"errors"
	"github.com/biorhitm/memfs"
	"io"
	"unsafe"
	"fmt"
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
	ltSymbol // ! @ # $ % ^ & * () - + = [] {} и т.д.
	ltIdent  // имя функции, переменной или типа
	ltEOL
	ltReservedWord //функция, конец, если и т.д.
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
	return (0x0410 <= C && C <= 0x042F) ||
		(0x0430 <= C && C <= 0x044F) ||
		C == 0x0401 || C == 0x0451
}

func isIdentLetter(C rune) bool {
	return C == '_' || isLetter(C)
}

func isDigit(C rune) bool {
	return '0' <= C && C <= '9'
}

func isSymbol(C rune) bool {
	return ('!' <= C && C <= '/') ||
		(':' <= C && C <= '@') ||
		('[' <= C && C <= '`') ||
		('{' <= C && C <= '~')
}

var InvalidRune = errors.New("Invalid utf8 char, support russian only")

func (R *TReader) readRune() (aChar rune, E error) {
	var size int = 1

	if R.Index >= R.Size {
		return 0, io.EOF
	}

	B := R.Text[R.Index]
	if B == 0xD0 {

		B1 := R.Text[R.Index+1]
		switch {
		case B1 == 0x81:
			{
				size = 2
				aChar = 'Ё'
			}

		case 0x90 <= B1 && B1 <= 0xAF:
			{
				size = 2
				S := "АБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"
				aChar = rune(S[B1-0x90])
			}

		case 0xB0 <= B1 && B1 <= 0xBF:
			{
				size = 2
				//return // а..п
			}

		default:
			return 0, InvalidRune
		}

	} else if B == 0xD1 {

		B1 := R.Text[R.Index+1]
		switch {
		case B1 == 0x91:
			{
				size = 2
				//return //ё
			}

		case 0x80 <= B1 && B1 <= 0x8F:
			{
				size = 2
				//return //р..я
			}

		default:
			return 0, InvalidRune
		}
	} else {
		aChar = rune(B)
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

	switch _type {
	case ltIdent:
		{
			R.Unread()
			startIndex = R.Index
			L.Text = memfs.PBigByteArray(unsafe.Pointer(&R.Text[R.Index]))
			for C, err := R.readRune(); err == nil; {
				if !isIdentLetter(C) && !isDigit(C) {
					R.Unread()
					break
				}
			}
		}
	}

	L.Size = int(R.Index - startIndex)

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

func (R *TReader) BuildLexems() (lexem PLexem, errorCode uint, errorIndex uint64) {
	var curLexem, firstLexem PLexem

	curLexem = new(TLexem)
	firstLexem = curLexem

	for {
		C, err := R.readRune();
		fmt.Printf("0x%02X ", C)
		if err != nil {
			return nil, 1, R.Index
		}

		switch {
		case isIdentLetter(C):
			{
				//curLexem = R.createNewLexem(curLexem, ltIdent)
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
				curLexem = R.createNewLexem(curLexem, ltSymbol)
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
