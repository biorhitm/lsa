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
			return fmt.Sprintf("Тип элемента: %d, ожидается %d, Лексема № %d",
				SD.LanguageItems[itemNo].Type,
				AStandardItems[itemNo].Type, itemNo), false
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

		case ltitString:
			{
				S = SD.StrStrings[idx]
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
	var (
		SD TSyntaxDescriptor
		E  error
		S  string
		ok bool
	)

	//**************************************************************************
	S = "Важное число = 42 + 1 - 17 / 2.71 * 0.37"
	//**************************************************************************
	if SD.Lexem, E = stringToLexems(S); E != nil {
		t.Fatal(E.Error())
	}

	if E = SD.translateAssignment(); E != nil {
		t.Fatal(E.Error())
	}

	S, ok = compareLanguageItems(SD, []tLanguageItem{
		{ltitIdent, "Важное число"},
		{ltitAssignment, ""},
		{ltitNumber, "42"},
		{ltitMathAdd, ""},
		{ltitNumber, "1"},
		{ltitMathSub, ""},
		{ltitNumber, "17"},
		{ltitMathDiv, ""},
		{ltitNumber, "2.71"},
		{ltitMathMul, ""},
		{ltitNumber, "0.37"},
	})

	if !ok {
		t.Fatal(S)
	}

	//**************************************************************************
	S = "Длина окружности = Diameter of the circle * PI * 3.14"
	//**************************************************************************
	SD.Init()
	if SD.Lexem, E = stringToLexems(S); E != nil {
		t.Fatal(E.Error())
	}

	if E = SD.translateAssignment(); E != nil {
		t.Fatal(E.Error())
	}

	S, ok = compareLanguageItems(SD, []tLanguageItem{
		{ltitIdent, "Длина окружности"},
		{ltitAssignment, ""},
		{ltitIdent, "Diameter of the circle"},
		{ltitMathMul, ""},
		{ltitIdent, "PI"},
		{ltitMathMul, ""},
		{ltitNumber, "3.14"},
	})

	if !ok {
		t.Fatal(S)
	}

	//**************************************************************************
	S = "Текст = \"Привет\" + \" \" + \"мир!\""
	//**************************************************************************
	SD.Init()
	if SD.Lexem, E = stringToLexems(S); E != nil {
		t.Fatal(E.Error())
	}
	if E = SD.translateAssignment(); E != nil {
		t.Fatal(E.Error())
	}
	S, ok = compareLanguageItems(SD, []tLanguageItem{
		{ltitIdent, "Текст"},
		{ltitAssignment, ""},
		{ltitString, "Привет"},
		{ltitMathAdd, ""},
		{ltitString, " "},
		{ltitMathAdd, ""},
		{ltitString, "мир!"},
	})
	if !ok {
		t.Fatal(S)
	}
}

func Test_addUnique(t *testing.T) {
	var list TStringArray
	list.addUnique("test")
	list.addUnique("2")
	list.addUnique("test")
	list.addUnique("3")
	list.addUnique("2")

	if list[0] != "test" || list[1] != "2" || list[2] != "3" {
		t.Fail()
	}
}

func TestTranslateFunctionDeclaration(t *testing.T) {
	var (
		SD TSyntaxDescriptor
		E  error
		S  string
		ok bool
	)

	//**************************************************************************
	S = "функция Сравнить элементы массива(Массив): Логический"
	//**************************************************************************
	if SD.Lexem, E = stringToLexems(S); E != nil {
		t.Fatal(E.Error())
	}

	if E = SD.translateFunctionDeclaration(); E != nil {
		t.Fatal(E.Error())
	}

	S, ok = compareLanguageItems(SD, []tLanguageItem{
		{ltitFunctionDeclaration, ""},
		{ltitIdent, "Сравнить элементы массива"},
		{ltitOpenParenthesis, ""},
		{ltitIdent, "Массив"},
		{ltitCloseParenthesis, ""},
		{ltitColon, ""},
		{ltitIdent, "Логический"},
	})

	if !ok {
		t.Fatal(S)
	}
}