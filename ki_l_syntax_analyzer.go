package lsa

import (
	"github.com/biorhitm/memfs"
	"io"
	"syscall"
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
	ltSymbol // ! @ # $ % ^ & * () - + = [] {} и т.д.
	ltIdent  // имя функции, переменной или типа
	ltEOL
	ltReservedWord //функция, конец, если и т.д.
)

type TLexem struct {
	Next PLexem
	Text PBigShortArray
	Size int
	Type TLexemType
}

type PLexem *TLexem

type TReader struct {
	Text  memfs.PBigByteArray
	Size  uint64
	Index uint64
}

func (self *TLexem) LexemAsString() string {
	S := ""

	if self.Size > 0 {
		b := make([]uint16, self.Size)
		for i := 0; i < self.Size; i++ {
			b[i] = self.Text[i]
		}
		S = syscall.UTF16ToString(b)
	}

	return S
}

func isLetter(C uint16) bool {
	return (0x0410 <= C && C <= 0x042F) ||
		(0x0430 <= C && C <= 0x044F) ||
		C == 0x0401 || C == 0x0451
}

func isIdentLetter(C uint16) bool {
	return C == '_' || isLetter(C)
}

func isDigit(C uint16) bool {
	return '0' <= C && C <= '9'
}

func isSymbol(C uint16) bool {
	return ('!' <= C && C <= '/') ||
		(':' <= C && C <= '@') ||
		('[' <= C && C <= '`') ||
		('{' <= C && C <= '~')
}

func createNewLexem(parent PLexem, text uint64, _type TLexemType) PLexem {
	L := new(TLexem)
	L.Text = PBigShortArray(unsafe.Pointer(uintptr(text)))
	L.Type = _type
	L.Size = 0
	L.Next = nil
	if parent != nil {
		parent.Next = L
	}
	return L
}

func (R *TReader) ReadRune()(aChar rune, aSize int, E error) {
	if R.Index >= R.Size {
		return 0, io.EOF
	}

	B := R.Text[R.Index]
	if B == 0xD0 {

		B1 := R.Text[R.Index+1]
		switch B1 {
			case 0x81: {
				return //Ё
			}

			case 0x90 <= B1 && B1 <= 0xAF: {
				return //A..Я
			}

			case 0xB0 <= B1 && B1 <= 0xBF: {
				return // а..п
			}
		}

	} else if B == 0xD1 {

		B1 := R.Text[R.Index+1]
		switch B1 {
			case 0x91: {
				return //ё
			}

			case 0x80 <= B1 && B1 <= 0x8F: {
				return //р..я
			}
		}
	}
}

// Error codes
const (
	errNoSuccess = iota
	errNoUnterminatedString
	errNoUnterminatedChar
)

func BuildLexems(aReader TReader) (lexem PLexem, errorCode uint, errorIndex uint64) {
	var idx uint64 = 1
	var C uint16
	var curLexem, firstLexem PLexem
	var startIdx uint64

	var T PBigShortArray = nil

	if text[0] == 0xFF && text[1] == 0xFE {
		// типа Unicode
		T = PBigShortArray(unsafe.Pointer(text))
		size = (size - 2) / 2
	}

	addrOfText := uint64(uintptr(unsafe.Pointer(T)))

	curLexem = new(TLexem)
	firstLexem = curLexem

	for idx <= size {
		C = T[idx]
		switch {
		case isIdentLetter(C):
			{
				startIdx = idx
				curLexem = createNewLexem(curLexem, addrOfText+idx*2, ltIdent)
				idx++
				for idx < size && (isIdentLetter(T[idx]) || isDigit(T[idx])) {
					idx++
				}
				curLexem.Size = int(idx - startIdx)
			}

		case isDigit(C):
			{
				startIdx = idx
				curLexem = createNewLexem(curLexem, addrOfText+idx*2, ltNumber)
				idx++
				for idx < size && isDigit(T[idx]) {
					idx++
				}
				curLexem.Size = int(idx - startIdx)
			}

		case C == '"':
			{
				startIdx = idx
				idx++
				if idx > size {
					return firstLexem, errNoUnterminatedString, startIdx
				}
				for idx <= size && T[idx] != '"' {
					idx++
				}
				if idx > size {
					return firstLexem, errNoUnterminatedString, startIdx
				}

				startIdx++
				curLexem = createNewLexem(curLexem, addrOfText+startIdx*2, ltString)
				curLexem.Size = int(idx - startIdx)
				idx++
			}

		case C == 0x27: //single quote
			{
				if (idx+2 > size) || (T[idx+2] != 0x27) {
					return firstLexem, errNoUnterminatedChar, idx
				}
				curLexem = createNewLexem(curLexem, addrOfText+(idx+1)*2, ltChar)
				curLexem.Size = 1
				idx += 3
			}

		case isSymbol(C):
			{
				curLexem = createNewLexem(curLexem, addrOfText+idx*2, ltSymbol)
				curLexem.Size = 1
				idx++
			}

		case C == 0x0A:
			{
				curLexem = createNewLexem(curLexem, addrOfText+idx*2, ltEOL)
				idx++
			}

		default:
			idx++
		}
	}

	createNewLexem(curLexem, 0, ltEOF)

	return firstLexem, errNoSuccess, 0
}
