package lsa

import (
	"fmt"
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

// Структура для тестов, вместо номеров строк содержит текст
//   для упрощения написания тестов.
type tLanguageItem struct {
	Type TLanguageItemType
	Text string
}

// Сравнивает список элементов языка с эталонным списком
// Возвращает истину, если все элементы равны
func compareLanguageItems(SD TSyntaxDescriptor,
	AStandardItems []tLanguageItem) (string, bool) {

	if len(SD.LanguageItems) != len(AStandardItems) {
		return fmt.Sprintf("Кол-во элементов языка: %d, должно быть: %d",
			len(SD.LanguageItems), len(AStandardItems)), false
	}
	for itemNo, _ := range SD.LanguageItems {
		T := SD.LanguageItems[itemNo].Type
		if T != AStandardItems[itemNo].Type {
			return fmt.Sprintf("Встретилась %d, ожидается %d, Лексема № %d",
				SD.LanguageItems[itemNo], AStandardItems[itemNo], itemNo), false
		}

		idx := SD.LanguageItems[itemNo].Index
		var S string

		switch T {
		case ltitIdent:
			{
				S = SD.StrIdents[idx]
			}

		case ltitNumber:
			{
				S = SD.StrNumbers[idx]
			}

		default:
			S = ""
		}

		if S != AStandardItems[itemNo].Text {
			return fmt.Sprintf(
				"Встретился '%s', ожидается '%s', Лексема № %d",
				S, AStandardItems[itemNo].Text, itemNo), false
		}
	}

	return "", true
}

func TestTranslateArgument(t *testing.T) {
	var SD TSyntaxDescriptor
	var E error

	if SD.Lexem, E = stringToLexems("((42))"); E != nil {
		t.Fatal(E.Error())
	}

	if E = SD.translateArgument(); E != nil {
		t.Fatal(E.Error())
	}

	languageItems := []tLanguageItem{
		{ltitOpenParenthesis, ""}, {ltitOpenParenthesis, ""},
		{ltitNumber, "42"},
		{ltitCloseParenthesis, ""}, {ltitCloseParenthesis, ""},
	}

	if S, ok := compareLanguageItems(SD, languageItems); !ok {
		t.Fatal(S)
	}
}

func TestTranslateAssigment(t *testing.T) {
	var SD TSyntaxDescriptor
	var E error

	if SD.Lexem, E = stringToLexems("Важное число = 42"); E != nil {
		t.Fatal(E.Error())
	}

	if E = SD.translateAssignment(); E != nil {
		t.Fatal(E.Error())
	}

	languageItems := []tLanguageItem{
		{ltitIdent, "Важное число"}, {ltitAssignment, ""},
		{ltitNumber, "42"},
	}

	if S, ok := compareLanguageItems(SD, languageItems); !ok {
		t.Fatal(S)
	}
}
