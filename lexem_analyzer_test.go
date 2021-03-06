package lsa

import (
	"fmt"
	"github.com/biorhitm/memfs"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"
)

var global_buffer = make([]uint8, 4000)

func stringToUTF8EncodedByteArray(S string) ([]uint8, error) {
	L := len(global_buffer) - 8
	if L < len(S) {
		return nil, fmt.Errorf("need to increase size of 'global_buffer', to %d",
			len(S)+8)
	}
	buf := global_buffer[L:]

	idx := 0
	reader := strings.NewReader(S)

	for {
		if R, _, E := reader.ReadRune(); E == nil {
			size := utf8.EncodeRune(buf, R)
			for i := 0; i < size; i++ {
				global_buffer[idx+i] = buf[i]
			}
			idx += size
		} else {
			break
		}
	}

	return global_buffer[:idx], nil
}

func Test_stringToUTF8EncodedByteArray(t *testing.T) {
	var S string = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ" +
		"абвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	stringUTF8encoding := []uint8{
		0xD0, 0x90, 0xD0, 0x91, 0xD0, 0x92, 0xD0, 0x93, 0xD0, 0x94, 0xD0, 0x95,
		0xD0, 0x81,
		0xD0, 0x96, 0xD0, 0x97, 0xD0, 0x98, 0xD0, 0x99, 0xD0, 0x9A, 0xD0, 0x9B,
		0xD0, 0x9C, 0xD0, 0x9D, 0xD0, 0x9E, 0xD0, 0x9F, 0xD0, 0xA0, 0xD0, 0xA1,
		0xD0, 0xA2, 0xD0, 0xA3, 0xD0, 0xA4, 0xD0, 0xA5, 0xD0, 0xA6, 0xD0, 0xA7,
		0xD0, 0xA8, 0xD0, 0xA9, 0xD0, 0xAA, 0xD0, 0xAB, 0xD0, 0xAC, 0xD0, 0xAD,
		0xD0, 0xAE, 0xD0, 0xAF,

		0xD0, 0xB0, 0xD0, 0xB1, 0xD0, 0xB2, 0xD0, 0xB3, 0xD0, 0xB4, 0xD0, 0xB5,
		0xD1, 0x91,
		0xD0, 0xB6, 0xD0, 0xB7, 0xD0, 0xB8, 0xD0, 0xB9, 0xD0, 0xBA, 0xD0, 0xBB,
		0xD0, 0xBC, 0xD0, 0xBD, 0xD0, 0xBE, 0xD0, 0xBF, 0xD1, 0x80, 0xD1, 0x81,
		0xD1, 0x82, 0xD1, 0x83, 0xD1, 0x84, 0xD1, 0x85, 0xD1, 0x86, 0xD1, 0x87,
		0xD1, 0x88, 0xD1, 0x89, 0xD1, 0x8A, 0xD1, 0x8B, 0xD1, 0x8C, 0xD1, 0x8D,
		0xD1, 0x8E, 0xD1, 0x8F,
	}

	buffer, E := stringToUTF8EncodedByteArray(S)
	if E != nil {
		t.Log(E)
		t.Fatal()
	}

	for i := 0; i < len(buffer); i++ {
		if buffer[i] != stringUTF8encoding[i] {
			t.Fatalf("stringToUTF8EncodedByteArray fail on %d", i)
		}
	}
}

func Test_readRune(t *testing.T) {
	var S string = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ" +
		"абвгдеёжзийклмнопрстуфхцчшщъыьэюя"

	buffer, _ := stringToUTF8EncodedByteArray(S)

	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buffer[0])),
		Size: uint64(len(S)),
	}
	stringReader := strings.NewReader(S)

	var E, E1 error
	var R, R1 rune
	var size int
	var runeNo uint = 0

	for {
		R, E = reader.readRune()
		R1, size, E1 = stringReader.ReadRune()

		if E != E1 {
			t.Fatal()
		}

		if E != nil {
			if E != io.EOF {
				t.Fatal()
			}
			break
		}

		if size != 2 {
			t.Fatal()
		}

		if reader.NextIndex-reader.Index != 2 {
			t.Fatalf("Rune have invalid size, index: %d", runeNo)
		}

		if R != R1 {
			t.Fatalf("Rune have invalid value 0x%X, must have 0x%X, index: %d",
				R, R1, runeNo)
		}

		runeNo++
	}
}

func Test_readRune_negative(t *testing.T) {

	buffer := []uint8{
		0x81, 0x24, 0xF8, 0xD0, 0x90,
	}
	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buffer[0])),
		Size: uint64(len(buffer)),
	}

	R, E := reader.readRune()
	if E != nil || R != 0x24 || (reader.NextIndex-reader.Index) != 1 {
		t.Fatalf("Ошибка пропускания байтов в середине серии: 10xx xxxx\n")
	}

	R, E = reader.readRune()
	if E != nil || R != 0x410 || (reader.NextIndex-reader.Index) != 2 {
		t.Fatalf("Ошибка пропускания серии > 4 байтов: 10xx xxxx\n")
	}
}

func TestIdentifierParser(t *testing.T) {
	var S string = "функция среднее арифметическое"
	buffer, _ := stringToUTF8EncodedByteArray(S)

	R := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buffer[0])),
		Size: uint64(len(S)),
	}

	plexem, E := R.BuildLexems()

	if E != nil {
		t.Fatal(E.Error())
	}

	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "функция" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "среднее" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "арифметическое" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEOF {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
}

func TestNumberParser(t *testing.T) {
	var S string = "\t_счётчик целое = 109\r\n"
	buf, _ := stringToUTF8EncodedByteArray(S)

	R := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf)),
	}

	plexem, E := R.BuildLexems()

	if E != nil {
		t.Fatal(E.Error())
	}

	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "_счётчик" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "целое" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEqualSign {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltNumber {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "109" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}
}

func TestStringParser(t *testing.T) {
	var S string = "\tА = \"Привет мир!\"\r\n"
	buf, _ := stringToUTF8EncodedByteArray(S)

	R := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf)),
	}

	plexem, E := R.BuildLexems()

	if E != nil {
		t.Fatal(E.Error())
	}

	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "А" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEqualSign {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltString {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "Привет мир!" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEOF {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
}

func TestCharParser(t *testing.T) {
	var S string = "\tБ = \n\n\n\n'$'\r\n"
	buf, _ := stringToUTF8EncodedByteArray(S)

	R := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf)),
	}

	plexem, E := R.BuildLexems()

	if E != nil {
		t.Fatal(E.Error())
	}

	if plexem.Type != ltIdent {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "Б" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEqualSign {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEOL {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}

	plexem = plexem.Next
	if plexem.Type != ltChar {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "$" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
	if plexem.Type != ltEOF {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
}

func ExampleUnterminatedStringError() {
	S := "\n\nС = \"test"
	_, E := stringToLexems(S)
	if E != EUnterminatedString {
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("[%v:%v] E != EUnterminatedString\n ", filepath.Base(file), line)
	}
	if E != nil {
		fmt.Print(E.Error())
		return
	}
	//Output: [2:4] Незакрытая строка, ожидается "
}

func ExampleUnterminatedCharError() {
	S := "\n\n\n\n\n\n\nСимвол = '$"
	_, E := stringToLexems(S)
	if E != EUnterminatedChar {
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("[%v:%v] E != EUnterminatedChar\n ", filepath.Base(file), line)
	}
	if E != nil {
		fmt.Print(E.Error())
		return
	}
	//Output: [7:9] Незакрытый символ, ожидается '
}
