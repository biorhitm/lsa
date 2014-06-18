package lsa

import (
	"github.com/biorhitm/memfs"
	"testing"
	"unsafe"
)

func stringToLexems(S string) (PLexem, error) {
	buf, _ := stringToUTF8EncodedByteArray(S)
	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf))}
	return reader.BuildLexems()
}

func TestTranslateArgument(t *testing.T) {
	var SD TSyntaxDescriptor
	var E error

	SD.Lexem, E = stringToLexems("((42))")
	if E != nil {
		t.Fatal(E.Error())
	}

	E = SD.translateArgument()
	if E != nil {
		t.Fatal(E.Error())
	}

	languageItems := []TLanguageItem{
		{ltitOpenParenthesis, 0},
		{ltitOpenParenthesis, 0},
		{ltitNumber, 0},
		{ltitCloseParenthesis, 0},
		{ltitCloseParenthesis, 0},
	}

	if len(SD.LanguageItems) != len(languageItems) {
		t.Fatalf("кол-во элементов языка: %d, должно быть: %d",
			len(SD.LanguageItems), len(languageItems))
	}
	for k, _ := range SD.LanguageItems {
		if SD.LanguageItems[k] != languageItems[k] {
			t.Errorf("Встретилась %d, ожидается %d, Лексема № %d",
				SD.LanguageItems[k], languageItems[k], k)
		}
	}
}
