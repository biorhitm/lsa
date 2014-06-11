package lsa

import (
	"fmt"
)

func getLexemAfterLexem(ALexem PLexem, _type TLexemType, text string) PLexem {
	for ALexem != nil {
		if ALexem.Type == _type && (*ALexem).LexemAsString() == text {
			return ALexem.Next
		}

		ALexem = ALexem.Next
	}
	return ALexem
}

/*
  'функция' <имя функции> '(' {<имя параметра> ':' <тип>} ')'
  <локальные переменные>
  'начало' <тело функции> 'конец'
*/
func generateFunction(ALexem PLexem) {
	var S string
	var L PLexem
	functionName := ALexem

	parameter := getLexemAfterLexem(ALexem, ltSymbol, "(")
	L = parameter
	for L != nil {
		if L.Type == ltSymbol && L.Text[0] == ')' {
			L = L.Next
			break
		}
		L = L.Next
	}

	if L != nil && L.Text[0] == ':' {
		L = L.Next
	}

	// печатаю тип функции
	for L != nil {
		if L.Type == ltEOL {
			L = L.Next
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)

		L = L.Next
	}

	localVars := L

	// печатаю имя функции
	L = functionName
	for L != nil {
		if L.Type == ltSymbol && L.Text[0] == '(' {
			fmt.Print(" (")
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)
		L = L.Next
	}

	// печатаю параметры функции
	L = parameter
	for L != nil {
		if L.Type == ltSymbol && L.Text[0] == ')' {
			fmt.Print(" )")
			break
		}

		S = (*L).LexemAsString()
		fmt.Printf(" %s", S)
		L = L.Next
	}

	L = getLexemAfterLexem(localVars, ltIdent, "начало")

	fmt.Print(" {\n")

	// печатаю тело функции
	for L != nil {
		if L.Type == ltIdent && (*L).LexemAsString() == "конец" {
			break
		}
		
		S = (*L).LexemAsString()
		fmt.Printf("\t%s\n", S)
		L = L.Next
	}
	fmt.Print("}\n")
}

func GenerateCode(ALexem PLexem) {
	for ALexem != nil {
		switch ALexem.Type {
		case ltIdent:
			{
				S := (*ALexem).LexemAsString()
				if S == "функция" {
					generateFunction(ALexem.Next)
				}
			}

		case ltEOL:
			{
				fmt.Println()
			}

		default:
			if ALexem.Size > 0 {
				fmt.Printf("Лехема: %d size: %d %s ",
					ALexem.Type, ALexem.Size, (*ALexem).LexemAsString())
			}
		}

		ALexem = ALexem.Next
	}

	fmt.Printf("----------EOF-----------\n")
}
