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

func TestTranslateArgument(t *testing.T) {
	L, E := stringToLexems("((42))")
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
	items, E := (*L).translateArgument()
	if len(items) != len(languageItems) {
		t.Fatal()
	}
	for k,_ := range items {
		if items[k] != languageItems[k] {
			t.Fatalf("Встретилась %d, ожидается %d, Лексема № %d",
			items[k], languageItems[k], k)
		}
	}
}

func ExampleTranslateAssignment() {
	S := "Результат странной формулы=(1024/((2+2*2)-(17)+((34-5)+12)*(6+7)))"
	L, E := stringToLexems(S)
	if E != nil {
		fmt.Print(E.Error())
		return
	}
	(*L).translateAssignment()
	//Output: Результат странной формулы = (1024 / ((2 + 2 * 2) - (17) + ((34 - 5) + 12) * (6 + 7)))
}
