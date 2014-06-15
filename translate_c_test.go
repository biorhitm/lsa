package lsa

import (
	"fmt"
	"github.com/biorhitm/memfs"
	"unsafe"
)

func stringToLexems(S string) (PLexem, error) {
	buf, _ := stringToUTF8EncodedByteArray(S)
	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf))}
	return reader.BuildLexems()
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
