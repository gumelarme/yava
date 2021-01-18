package text

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	Start              TokenType = iota
	Id                           // a-zA-Z 0-9 _ $
	Constant                     // 0b1111011'(2), 0173(8), 123(10), 0x7b(16)
	Keyword                      // listed below
	ArithmeticOperator           // + - * / %
	RelationOperator             // < > <= >= == !=
	BitwiseOperator              // & | ^ ~ << >> >>>
	LogicalOperator              // && || !
	AssignmentOperator           // += -= *= /= %=  <<= >>= >>>= &= |= ^=
	Separator                    // ; , . ? : @ ' " (  ) [  ] {  }
	Unknown
)

type Token struct {
	Linum int
	Value string
	Type  TokenType
}

var keywords = [...]string{
	// *	 	not used => goto, const
	// **	 	added in 1.2 => strictfp
	// ***	 	added in 1.4 => assert
	// ****	 	added in 5.0 => enum

	"abstract", "assert",
	"boolean", "break", "byte",
	"case", "catch", "char", "class", "const", "continue",
	"default", "do", "double",
	"else", "enum", "extends",
	"final", "finally", "float", "for",
	"goto", "if", "implements", "import", "instanceof", "int", "interface",
	"long", "native", "new",
	"package", "private", "protected", "public", "return",
	"short", "static", "strictfp", "super", "switch", "synchronized",
	"this", "throw", "throws", "transient", "try",
	"void", "volatile", "while",
}

func (t Token) String() string {
	return fmt.Sprintf("")
}

func (t TokenType) String() string {
	return [...]string{
		"Start", "Id", "Constant", "Keyword",
		"ArithmeticOperator", "RelationOperator",
		"Separator", "Assignment",
	}[t]
}

func Splitter(str string) chan T {
	c := make(chan T)

	go func() {
		for _, char := range str {
			c <- T{Position{0, 0}, char}
		}
		close(c)
	}()
	return c
}

func Parse(streamFn Scanner) chan Token {
	// relationOperator := "< <= > >= == !="
	c := make(chan Token)
	go func() {
		var chunk strings.Builder
		token := Start
		for streamFn.HasNext() {
			ch := streamFn.Next()
			if token == Start {
				chunk.WriteRune(ch.Char)
				token = checkStartToken(ch.Char)

				if token == ArithmeticOperator || token == Separator {
					c <- Token{ch.Position.Linum, chunk.String(), token}
					continue
				} else if token == Unknown {
					chunk.Reset()
				}
			}
		}
		close(c)
	}()
	return c
}

func checkStartToken(r rune) (token TokenType) {
	arithmetic := "+-*/"
	separator := "(){};:[].,"
	relationStart := "<>=!"

	if isLetter(r) {
		token = Id
	} else if isDigit(r) {
		token = Constant
	} else if contains(arithmetic, r) {
		token = ArithmeticOperator
	} else if contains(separator, r) {
		token = Separator
	} else if contains(relationStart, r) {
		token = RelationOperator
	} else {
		token = Unknown
	}

	return
}

func checkConstantOrId(r rune, chunk *strings.Builder, ttype TokenType, linum int) (Token, error) {
	if !(isLetter(r) || isDigit(r)) {
		for _, k := range keywords {
			if chunk.String() == k {
				ttype = Keyword
			}
		}

		if unicode.IsSpace(r) {
			return Token{linum, chunk.String(), ttype}, nil
		} else {

		}
	}
	return Token{}, errors.New("Something wrong happened while checking for constant ID")
}

func isLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func contains(chars string, r rune) bool {
	for _, c := range chars {
		if r == c {
			return true
		}
	}

	return false
}
