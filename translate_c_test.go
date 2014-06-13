package lsa

import (
	"github.com/biorhitm/memfs"
	"testing"
	"unsafe"
)

func stringToLexems(S string) (PLexem, uint, uint64) {
	buf, _ := stringToUTF8EncodedByteArray(S)
	reader := TReader{
		Text: memfs.PBigByteArray(unsafe.Pointer(&buf[0])),
		Size: uint64(len(buf))}
	return reader.BuildLexems()
}

func TestTranslateCode(t *testing.T) {
	plexem, errorCode, _ := stringToLexems(
		"Результат странной формулы=(1024/((2+2*2)-(17)+((34-5)+12)*(6+7)))")
	if errorCode != 0 {
		t.Fail()
	}
	plexem = plexem.Next
	plexem, E := (*plexem).translateAssignment()
	if E != nil {
		t.Error(E)
	}
}
