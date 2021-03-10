// +build integration

package text

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func withLexerFile(fname string, test func(lx *Lexer)) {
	sc := NewFileScanner(fname)
	lx := NewLexer(sc)
	defer sc.Close()
	test(&lx)
}

func TestLexer_NextToken_text(t *testing.T) {
	data := []struct {
		text   string
		tokens []Token
	}{
		{
			"int a = 5;",
			[]Token{
				newToken(1, 0, "int", Keyword),
				newToken(1, 4, "a", Id),
				newToken(1, 6, "=", Assignment),
				newTokenSub(1, 8, "5", IntegerLiteral, Decimal),
				newToken(1, 9, ";", Semicolon),
			},
		},
		{
			"int a = 5_10_100;",
			[]Token{
				newToken(1, 0, "int", Keyword),
				newToken(1, 4, "a", Id),
				newToken(1, 6, "=", Assignment),
				newTokenSub(1, 8, "5_10_100", IntegerLiteral, Decimal),
				newToken(1, 16, ";", Semicolon),
			},
		},
		{
			"char a; a = 'n'",
			[]Token{
				newToken(1, 0, "char", Keyword),
				newToken(1, 5, "a", Id),
				newToken(1, 6, ";", Semicolon),
				newToken(1, 8, "a", Id),
				newToken(1, 10, "=", Assignment),
				newToken(1, 12, "n", CharLiteral),
			},
		},
		{
			"person.size = .32f;",
			[]Token{
				newToken(1, 0, "person", Id),
				newToken(1, 6, ".", Dot),
				newToken(1, 7, "size", Id),
				newToken(1, 12, "=", Assignment),
				newTokenSub(1, 14, ".32f", FloatingPointLiteral, Decimal),
				newToken(1, 18, ";", Semicolon),
			},
		},
		{
			`str = "Value: " + (0xff + 20.3f) + obj.getSuffix();`,
			[]Token{
				newToken(1, 0, "str", Id),
				newToken(1, 4, "=", Assignment),
				newToken(1, 6, "Value: ", StringLiteral),
				newToken(1, 16, "+", Addition),
				newToken(1, 18, "(", LeftParenthesis),
				newTokenSub(1, 19, "0xff", IntegerLiteral, Hex),
				newToken(1, 24, "+", Addition),
				newTokenSub(1, 26, "20.3f", FloatingPointLiteral, Decimal),
				newToken(1, 31, ")", RightParenthesis),
				newToken(1, 33, "+", Addition),
				newToken(1, 35, "obj", Id),
				newToken(1, 38, ".", Dot),
				newToken(1, 39, "getSuffix", Id),
				newToken(1, 48, "(", LeftParenthesis),
				newToken(1, 49, ")", RightParenthesis),
				newToken(1, 50, ";", Semicolon),
			},
		},
		{
			"if(something == true)",
			[]Token{
				newToken(1, 0, "if", Keyword),
				newToken(1, 2, "(", LeftParenthesis),
				newToken(1, 3, "something", Id),
				newToken(1, 13, "==", Equal),
				newToken(1, 16, "true", BooleanLiteral),
				newToken(1, 20, ")", RightParenthesis),
			},
		},
		{
			`float something = 90.3d;
something += 30f;`,
			[]Token{
				newToken(1, 0, "float", Keyword),
				newToken(1, 6, "something", Id),
				newToken(1, 16, "=", Assignment),
				newTokenSub(1, 18, "90.3d", FloatingPointLiteral, Decimal),
				newToken(1, 23, ";", Semicolon),
				newToken(2, 0, "something", Id),
				newToken(2, 10, "+=", AdditionAssignment),
				newTokenSub(2, 13, "30f", FloatingPointLiteral, Decimal),
				newToken(2, 16, ";", Semicolon),
			},
		},
		// comment should be skipped, but still count position
		{
			`//something
int a;`,
			[]Token{
				newToken(2, 0, "int", Keyword),
				newToken(2, 4, "a", Id),
				newToken(2, 5, ";", Semicolon),
			},
		},
		{
			`/* int a; */ float b;`,
			[]Token{
				newToken(1, 13, "float", Keyword),
				newToken(1, 19, "b", Id),
				newToken(1, 20, ";", Semicolon),
			},
		},
		{
			`/*
This thing should not be read by the compiler */
String a = "Hello";`,
			[]Token{
				newToken(3, 0, "String", Id),
				newToken(3, 7, "a", Id),
				newToken(3, 9, "=", Assignment),
				newToken(3, 11, "Hello", StringLiteral),
				newToken(3, 18, ";", Semicolon),
			},
		},
	}

	for _, d := range data {
		withLexer(d.text, func(lx *Lexer) {
			next, i := lx.NextToken, 0
			for tok, e := next(); e == nil; tok, e = next() {
				if !tok.Equal(d.tokens[i]) {
					t.Errorf("Token does not match, got %s instead of %s",
						tok,
						d.tokens[i],
					)
				}
				i++
			}
		})
	}
}

func TestLexer_NextToken_file(t *testing.T) {

	files := []string{
		"mixed_ascii_unicode",
		"full_unicode",
	}

	for _, filename := range files {
		jsonFile, _ := ioutil.ReadFile("testdata/" + filename + ".json")
		var jTokens []jsonToken
		err := json.Unmarshal(jsonFile, &jTokens)

		if err != nil {
			t.Error(err)
			return
		}

		withLexerFile("testdata/"+filename+".java", func(lx *Lexer) {
			i := 0
			for tok, e := lx.NextToken(); e == nil; tok, e = lx.NextToken() {
				jToken := jTokens[i].Token()
				if !tok.Equal(jToken) {
					t.Errorf("%s Token does not match, got %s instead of %s",
						filename,
						tok,
						jToken,
					)
				}
				i++
			}

		})
	}
}

// func TestLexer_NextToken_panic(t *testing.T) {
// 	data := []string{
// 		//incomplete
// 		`"something`,
// 		`'s`,
// 		//illegal syntax
// 		`'nice'`,
// 		`0x`,
// 		`0b`,
// 		`09`, // illegal octal
// 	}

// 	for _, str := range data {
// 		withLexer(str, func(lx *Lexer) {
// 			msg := fmt.Sprintf("Should panic on %s", str)
// 			defer assertPanic(t, msg)

// 			tok, _ := lx.NextToken()
// 			fmt.Println(tok)
// 		})
// 	}
// }
