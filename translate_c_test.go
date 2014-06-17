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
	lexems, E := stringToLexems("((42))")
	if E != nil {
		t.Fatal(E.Error())
	}
	items := make([]TLanguageItem, 0)
	E = (*lexems).translateArgument(&items)
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

	if len(items) != len(languageItems) {
		t.Fatalf("кол-во элементов языка: %d, должно быть: %d",
			len(items), len(languageItems))
	}
	for k, _ := range items {
		if items[k] != languageItems[k] {
			t.Errorf("Встретилась %d, ожидается %d, Лексема № %d",
				items[k], languageItems[k], k)
		}
	}
}
