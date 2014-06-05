package lsa

import (
	"fmt"
	"github.com/biorhitm/memfs"
	"testing"
	"unsafe"
)

func TestIdentifierParser(t *testing.T) {
	var S string = "функция среднее арифметическое"

	buf := []uint8(S)
	/*buf := []uint8{
	0x44, 0x04, 0x43, 0x04, 0x3D, 0x04, 0x3A, 0x04, 0x46, 0x04, 0x38, 0x04, 0x4F, 0x04, 0x20, 0,
	0x41, 0x04, 0x40, 0x04, 0x35, 0x04, 0x34, 0x04, 0x3D, 0x04, 0x35, 0x04, 0x35, 0x04, 0x20, 0,
	0x30, 0x04, 0x40, 0x04, 0x38, 0x04, 0x44, 0x04, 0x3C, 0x04, 0x35, 0x04, 0x42, 0x04, 0x38, 0x04, 0x47, 0x04, 0x35, 0x04, 0x41, 0x04, 0x3A, 0x04, 0x3E, 0x04, 0x35, 0x04, 0x28, 0x00, 0x30, 0x04}
	*/
	R := TReader{
		Text:      memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size:      uint64(len(buf)),
		Index:     0,
		PrevIndex: 0}

	plexem, errorCode, _ := R.BuildLexems()

	if errorCode != 0 {
		t.Fatalf("errorCode: %d", errorCode)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
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
}

func TestNumberParser(t *testing.T) {
t.Skip()
	var S string
	buf := []uint8{
		0x09, 0x00, 0x5F, 0x00, 0x41, 0x04, 0x47, 0x04, 0x51, 0x04, 0x42, 0x04, 0x47, 0x04, 0x38, 0x04, 0x3A, 0x04, 0x20, 0,
		0x46, 0x04, 0x35, 0x04, 0x3B, 0x04, 0x3E, 0x04, 0x35, 0x04, 0x20, 0,
		0x3D, 0x00, 0x20, 0,
		0x31, 0x00, 0x30, 0x00, 0x39, 0x00, 0x0D, 0x00, 0x0A, 0}

	R := TReader{
		Text:      memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size:      uint64(len(buf)),
		Index:     0,
		PrevIndex: 0}

	plexem, errorCode, _ := R.BuildLexems()

	if errorCode != 0 {
		t.Fatalf("errorCode: %d", errorCode)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
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
	if plexem.Type != ltSymbol {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "=" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
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
	fmt.Printf("%s", buf)
}

func TestStringParser(t *testing.T) {
t.Skip()
	var S string
	buf := []uint8{
		0x09, 0x00, 0x10, 0x04, 0x20, 0,
		0x3D, 0x00, 0x20, 0,
		0x22, 0x00, 0x1F, 0x04, 0x40, 0x04, 0x38, 0x04, 0x32, 0x04, 0x35, 0x04, 0x42, 0x04, 0x20, 0,
		0x3C, 0x04, 0x38, 0x04, 0x40, 0x04, 0x21, 0x00, 0x22, 0x00, 0x0D, 0x00, 0x0A, 0}

	R := TReader{
		Text:      memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size:      uint64(len(buf)),
		Index:     0,
		PrevIndex: 0}

	plexem, errorCode, _ := R.BuildLexems()

	if errorCode != 0 {
		t.Fatalf("errorCode: %d", errorCode)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
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
	if plexem.Type != ltSymbol {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "=" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
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
}

func TestCharParser(t *testing.T) {
t.Skip()
	var S string
	buf := []uint8{
		0x09, 0x00, 0x11, 0x04, 0x20, 0,
		0x3D, 0x00, 0x20, 0,
		0x27, 0x00, 0x24, 0x00, 0x27, 0x00, 0x0D, 0x00, 0x0A, 0}

	R := TReader{
		Text:      memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size:      uint64(len(buf)),
		Index:     0,
		PrevIndex: 0}

	plexem, errorCode, _ := R.BuildLexems()

	if errorCode != 0 {
		t.Fatalf("errorCode: %d", errorCode)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
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
	if plexem.Type != ltSymbol {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
	S = (*plexem).LexemAsString()
	if S != "=" {
		t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)
	}

	plexem = plexem.Next
	if plexem == nil {
		t.Fatal("Мало лексем")
	}
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
	if plexem.Type != ltEOL {
		t.Errorf("Неправильный тип: %d", plexem.Type)
	}
}

func TestReadRune(t *testing.T) {
	//var R TReader{nil, 2, 0}
}
