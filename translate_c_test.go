package lsa

import (
	"github.com/biorhitm/memfs"
	"unsafe"
)

func stringToLexems(S string) (PLexem, uint, uint64) {
	buf, _ := stringToUTF8EncodedByteArray(S)
	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf))}
	return reader.BuildLexems()
}

func ExampleTranslateAssignment() {
	S := "Результат странной формулы=(1024/((2+2*2)-(17)+((34-5)+12)*(6+7)))"
	L, errorCode, _ := stringToLexems(S)
	if errorCode == 0 {
		(*L).translateAssignment()
	}
	//Output: Результат странной формулы = (1024 / ((2 + 2 * 2) - (17) + ((34 - 5) + 12) * (6 + 7)))
}
