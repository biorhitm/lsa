package lsa

import (
	"github.com/biorhitm/memfs"
	"testing"
	"unsafe"
)

func TestIdentifierParser(t *testing.T) {
	var S string
	buf := []uint8{0xFF, 0xFE,
		0x44, 0x04, 0x43, 0x04, 0x3D, 0x04, 0x3A, 0x04, 0x46, 0x04, 0x38, 0x04, 0x4F, 0x04, 0x20, 0,
		0x41, 0x04, 0x40, 0x04, 0x35, 0x04, 0x34, 0x04, 0x3D, 0x04, 0x35, 0x04, 0x35, 0x04, 0x20, 0,
		0x30, 0x04, 0x40, 0x04, 0x38, 0x04, 0x44, 0x04, 0x3C, 0x04, 0x35, 0x04, 0x42, 0x04, 0x38, 0x04, 0x47, 0x04, 0x35, 0x04, 0x41, 0x04, 0x3A, 0x04, 0x3E, 0x04, 0x35, 0x04, 0x28, 0x00, 0x30, 0x04}

	p := memfs.PBigByteArray(unsafe.Pointer(&buf[0]))

	plexem, errorCode, _ := BuildLexems(p, uint64(len(buf)))

	if errorCode != 0 {t.Fatalf("errorCode: %d", errorCode)}
	
	plexem = plexem.Next
	if plexem == nil {t.Fatal("Мало лексем")}
	if plexem.Type != ltIdent {t.Errorf("Неправильный тип: %d", plexem.Type)}
	S = (*plexem).LexemAsString()
	if S != "функция" {t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)}

	plexem = plexem.Next
	if plexem == nil {t.Fatal("Мало лексем")}
	if plexem.Type != ltIdent {t.Errorf("Неправильный тип: %d", plexem.Type)}
	S = (*plexem).LexemAsString()
	if S != "среднее" {t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)}

	plexem = plexem.Next
	if plexem == nil {t.Fatal("Мало лексем")}
	if plexem.Type != ltIdent {t.Errorf("Неправильный тип: %d", plexem.Type)}
	S = (*plexem).LexemAsString()
	if S != "арифметическое" {t.Errorf("Лексема содержит неправильный текст: \"%s\"", S)}
}
