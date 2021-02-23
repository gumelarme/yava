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
				newTokenSub(1, 6, "=", Operator, AssignmentOperator),
				newTokenSub(1, 8, "5", IntegerLiteral, Decimal),
				newToken(1, 9, ";", Separator),
			},
		},
		{
			"int a = 5_10_100;",
			[]Token{
				newToken(1, 0, "int", Keyword),
				newToken(1, 4, "a", Id),
				newTokenSub(1, 6, "=", Operator, AssignmentOperator),
				newTokenSub(1, 8, "5_10_100", IntegerLiteral, Decimal),
				newToken(1, 16, ";", Separator),
			},
		},
		{
			"char a; a = 'n'",
			[]Token{
				newToken(1, 0, "char", Keyword),
				newToken(1, 5, "a", Id),
				newToken(1, 6, ";", Separator),
				newToken(1, 8, "a", Id),
				newTokenSub(1, 10, "=", Operator, AssignmentOperator),
				newToken(1, 12, "n", CharLiteral),
			},
		},
		{
			`str = "Value: " + (0xff + 20.3f) + obj.getSuffix();`,
			[]Token{
				newToken(1, 0, "str", Id),
				newTokenSub(1, 4, "=", Operator, AssignmentOperator),
				newToken(1, 6, "Value: ", StringLiteral),
				newTokenSub(1, 16, "+", Operator, ArithmeticOperator),
				newToken(1, 18, "(", Separator),
				newTokenSub(1, 19, "0xff", IntegerLiteral, Hex),
				newTokenSub(1, 24, "+", Operator, ArithmeticOperator),
				newTokenSub(1, 26, "20.3f", FloatingPointLiteral, Decimal),
				newToken(1, 31, ")", Separator),
				newTokenSub(1, 33, "+", Operator, ArithmeticOperator),
				newToken(1, 35, "obj", Id),
				newToken(1, 38, ".", Separator),
				newToken(1, 39, "getSuffix", Id),
				newToken(1, 48, "(", Separator),
				newToken(1, 49, ")", Separator),
				newToken(1, 50, ";", Separator),
			},
		},
		{
			"if(something == true)",
			[]Token{
				newToken(1, 0, "if", Keyword),
				newToken(1, 2, "(", Separator),
				newToken(1, 3, "something", Id),
				newTokenSub(1, 13, "==", Operator, RelationOperator),
				newToken(1, 16, "true", BooleanLiteral),
				newToken(1, 20, ")", Separator),
			},
		},
		{
			`float something = 90.3d;
something += 30f;`,
			[]Token{
				newToken(1, 0, "float", Keyword),
				newToken(1, 6, "something", Id),
				newTokenSub(1, 16, "=", Operator, AssignmentOperator),
				newTokenSub(1, 18, "90.3d", FloatingPointLiteral, Decimal),
				newToken(1, 23, ";", Separator),
				newToken(2, 0, "something", Id),
				newTokenSub(2, 10, "+=", Operator, AssignmentOperator),
				newTokenSub(2, 13, "30f", FloatingPointLiteral, Decimal),
				newToken(2, 16, ";", Separator),
			},
		},
		// comment should be skipped, but still count position
		{
			`//something
int a;`,
			[]Token{
				newToken(2, 0, "int", Keyword),
				newToken(2, 4, "a", Id),
				newToken(2, 5, ";", Separator),
			},
		},
		{
			`/* int a; */ float b;`,
			[]Token{
				newToken(1, 13, "float", Keyword),
				newToken(1, 19, "b", Id),
				newToken(1, 20, ";", Separator),
			},
		},
		{
			`/*
This thing should not be read by the compiler */
String a = "Hello";`,
			[]Token{
				newToken(3, 0, "String", Id),
				newToken(3, 7, "a", Id),
				newTokenSub(3, 9, "=", Operator, AssignmentOperator),
				newToken(3, 11, "Hello", StringLiteral),
				newToken(3, 18, ";", Separator),
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
