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
	items, E := (*L).translateArgument()
	if len(items) != 5 {
		t.Fatal()
	}
	if items[0].Type != ltitOpenParenthesis ||
		items[1].Type != ltitOpenParenthesis {
		t.Fatal()
	}
	if items[2].Type != ltitNumber &&
		strNumbers[items[2].Index] != "42" {
		t.Fatal()
	}
	if items[3].Type != ltitCloseParenthesis ||
		items[4].Type != ltitCloseParenthesis {
		t.Fatal()
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
