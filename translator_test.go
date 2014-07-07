package lsa

import (
	"errors"
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
	AStandardItems []tLanguageItem) (mes string, result bool) {

	result = true

	for itemNo, _ := range SD.LanguageItems {
		T := SD.LanguageItems[itemNo].Type
		if result && itemNo >= len(AStandardItems) {
			mes = fmt.Sprintf("Тип элемента: %d, ожидается конец, Лексема № %d",
				T, itemNo)
			result = false
		}

		if result && T != AStandardItems[itemNo].Type {
			mes = fmt.Sprintf("Тип элемента: %d, ожидается %d, Лексема № %d",
				SD.LanguageItems[itemNo].Type,
				AStandardItems[itemNo].Type, itemNo)
			result = false
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

		if result {
			if S != AStandardItems[itemNo].Text {
				mes = fmt.Sprintf(
					"Встретился '%s', ожидается '%s', Лексема № %d",
					S, AStandardItems[itemNo].Text, itemNo)
				result = false
			}
		} else {
			if S != "" {
				S = "'" + S + "'"
			}
			fmt.Printf("[%d]: %s ", itemNo, S)
		}
	}
	if result && len(SD.LanguageItems) != len(AStandardItems) {
		mes = fmt.Sprintf("Кол-во элементов языка: %d, должно быть: %d",
			len(SD.LanguageItems), len(AStandardItems))
		result = false
	}

	return
}

func compareStringAndLanguageItems(AText string, AItems []tLanguageItem) error {
	var (
		lexems *TLexem
		sd     TSyntaxDescriptor
		E      error
	)
	if lexems, E = stringToLexems(AText); E != nil {
		return E
	}
	if sd, E = TranslateCode(lexems); E != nil {
		return E
	}
	if S, ok := compareLanguageItems(sd, AItems); !ok {
		return errors.New(S)
	}
	return nil
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
	if E := compareStringAndLanguageItems(
		"Важное число = 42 + 1 - 17 / 2.71 * 0.37",
		[]tLanguageItem{
			{ltitIdent, "Важное число"},
			{ltitAssignment, ""},
			{ltitNumber, "42"}, {ltitMathAdd, ""}, {ltitNumber, "1"},
			{ltitMathSub, ""}, {ltitNumber, "17"},
			{ltitMathDiv, ""}, {ltitNumber, "2.71"},
			{ltitMathMul, ""}, {ltitNumber, "0.37"},
		}); E != nil {
		t.Fatal(E.Error())
	}

	if E := compareStringAndLanguageItems(
		"Длина окружности = Diameter of the circle * PI * 3.14",
		[]tLanguageItem{
			{ltitIdent, "Длина окружности"},
			{ltitAssignment, ""},
			{ltitIdent, "Diameter of the circle"},
			{ltitMathMul, ""},
			{ltitIdent, "PI"},
			{ltitMathMul, ""},
			{ltitNumber, "3.14"},
		}); E != nil {
		t.Fatal(E.Error())
	}

	if E := compareStringAndLanguageItems(
		"Текст = \"Привет\" + \" \" + \"мир!\"",
		[]tLanguageItem{
			{ltitIdent, "Текст"}, {ltitAssignment, ""}, {ltitString, "Привет"},
			{ltitMathAdd, ""}, {ltitString, " "}, {ltitMathAdd, ""},
			{ltitString, "мир!"},
		}); E != nil {
		t.Fatal(E.Error())
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
	if E := compareStringAndLanguageItems(
		"функция Имя класса.Имя функции(А: Новый Тип, Б,В,Г: пакет.Тип) начало конец",
		[]tLanguageItem{
			{ltitFunction, ""},
			{ltitClassMember, ""}, {ltitIdent, "Имя класса"},
			{ltitIdent, "Имя функции"},
			{ltitParameters, ""}, {ltitIdent, "А"}, {ltitDataType, ""},
			{ltitIdent, "Новый Тип"},
			{ltitIdent, "Б"}, {ltitIdent, "В"}, {ltitIdent, "Г"},
			{ltitDataType, ""},
			{ltitPackageName, ""}, {ltitIdent, "пакет"},
			{ltitIdent, "Тип"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}

	if E := compareStringAndLanguageItems(
		"function F(А,Б: Int64): System.bool начало конец",
		[]tLanguageItem{
			{ltitFunction, ""}, {ltitIdent, "F"},
			{ltitParameters, ""}, {ltitIdent, "А"}, {ltitIdent, "Б"},
			{ltitDataType, ""}, {ltitIdent, "Int64"},
			{ltitDataType, ""},
			{ltitPackageName, ""}, {ltitIdent, "System"},
			{ltitIdent, "bool"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}

	if E := compareStringAndLanguageItems(
		"func foo(): int переменные А, Б, В: Unicode Символ начало конец",
		[]tLanguageItem{
			{ltitFunction, ""}, {ltitIdent, "foo"},
			{ltitDataType, ""}, {ltitIdent, "int"},
			{ltitVarList, ""},
			{ltitIdent, "А"}, {ltitIdent, "Б"}, {ltitIdent, "В"},
			{ltitDataType, ""}, {ltitIdent, "Unicode Символ"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}

	if E := compareStringAndLanguageItems(
		"def foo: Тип функции foo var А, Б, В: Unicode Символ {}",
		[]tLanguageItem{
			{ltitFunction, ""}, {ltitIdent, "foo"},
			{ltitDataType, ""}, {ltitIdent, "Тип функции foo"},
			{ltitVarList, ""},
			{ltitIdent, "А"}, {ltitIdent, "Б"}, {ltitIdent, "В"},
			{ltitDataType, ""}, {ltitIdent, "Unicode Символ"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}
}

func TestTranslateCode(t *testing.T) {
	if E := compareStringAndLanguageItems(
		"def foo: Тип функции foo var А, Б, В: Unicode Символ {}",
		[]tLanguageItem{
			{ltitFunction, ""}, {ltitIdent, "foo"},
			{ltitDataType, ""}, {ltitIdent, "Тип функции foo"},
			{ltitVarList, ""},
			{ltitIdent, "А"}, {ltitIdent, "Б"}, {ltitIdent, "В"},
			{ltitDataType, ""}, {ltitIdent, "Unicode Символ"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}
}

func TestIfStatement(t *testing.T) {
	if E := compareStringAndLanguageItems(
		"if S = 15 {} else {}",
		[]tLanguageItem{
			{ltitIf, ""},
			{ltitIdent, "S"}, {ltitEqual, ""}, {ltitNumber, "15"},
			{ltitBegin, ""}, {ltitEnd, ""},
			{ltitElse, ""},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}
}

func TestVarList(t *testing.T) {
	if E := compareStringAndLanguageItems(
		"переменные Возраст: целый",
		[]tLanguageItem{
			{ltitVarList, ""},
			{ltitIdent, "Возраст"}, {ltitDataType, ""}, {ltitIdent, "целый"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"var A, B, C: int, S: string, булева переменная: система.булевый",
		[]tLanguageItem{
			{ltitVarList, ""},
			{ltitIdent, "A"},
			{ltitIdent, "B"},
			{ltitIdent, "C"}, {ltitDataType, ""}, {ltitIdent, "int"},
			{ltitIdent, "S"}, {ltitDataType, ""}, {ltitIdent, "string"},
			{ltitIdent, "булева переменная"},
			{ltitDataType, ""}, {ltitPackageName, ""},
			{ltitIdent, "система"}, {ltitIdent, "булевый"},
		}); E != nil {
		t.Fatal(E.Error())
	}
}

func TestWhileStatement(t *testing.T) {
	if E := compareStringAndLanguageItems(
		"пока рак на горе = свистнет начало конец",
		[]tLanguageItem{
			{ltitWhile, ""},
			{ltitIdent, "рак на горе"}, {ltitEqual, ""},
			{ltitIdent, "свистнет"},
			{ltitBegin, ""}, {ltitEnd, ""},
		}); E != nil {
		t.Fatal(E.Error())
	}
}

func TestOperations(t *testing.T) {
	if E := compareStringAndLanguageItems(
		"A = B > C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitAbove, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"A = B >= C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitAboveEqual, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"A = B >> C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitRightShift, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"A = B < C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitBelow, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"A = B <= C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitBelowEqual, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
	if E := compareStringAndLanguageItems(
		"A = B << C",
		[]tLanguageItem{
			{ltitIdent, "A"}, {ltitAssignment, ""},
			{ltitIdent, "B"}, {ltitLeftShift, ""}, {ltitIdent, "C"},
		}); E != nil {
		t.Fatal(E.Error())
	}
}
