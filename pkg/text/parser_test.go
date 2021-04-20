package text

import (
	"fmt"
	"testing"
)

func withParser(s string, parserAction func(p *Parser)) {
	withLexer(s, func(lx *Lexer) {
		p := NewParser(lx)
		parserAction(&p)
	})
}

func TestParser_fieldAccess(t *testing.T) {
	data := []struct {
		str    string
		expect NamedValue
	}{
		{"person", &FieldAccess{"person", nil}},
		{"person.age", &FieldAccess{"person", &FieldAccess{"age", nil}}},
		{"person.age.calculate", &FieldAccess{"person", &FieldAccess{"age", &FieldAccess{"calculate", nil}}}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.fieldAccess()
			if PrettyPrint(result) != PrettyPrint(d.expect) {
				t.Errorf("Got %#v instead of %#v", PrettyPrint(result), PrettyPrint(d.expect))
			}
		})
	}
}

func TestParser_methodCall(t *testing.T) {
	data := []struct {
		str    string
		expect NamedValue
	}{
		{"person()", &MethodCall{"person", []Expression{}, nil}},
		{"person(1, 2, 3)", &MethodCall{"person", []Expression{Num(1), Num(2), Num(3)}, nil}},
		{"person.age()", &FieldAccess{"person", &MethodCall{"age", []Expression{}, nil}}},
		{"person().age", &MethodCall{"person", []Expression{}, &FieldAccess{"age", nil}}},
		{"person()[0]", &MethodCall{"person", []Expression{}, &ArrayAccess{Num(0), nil}}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.fieldAccess()
			if PrettyPrint(result) != PrettyPrint(d.expect) {
				t.Errorf("Got %#v instead of %#v", PrettyPrint(result), PrettyPrint(d.expect))
			}
		})
	}
}

func TestParser_arrayAccess(t *testing.T) {
	data := []struct {
		str    string
		expect NamedValue
	}{
		{"person[1]", &FieldAccess{"person", &ArrayAccess{Num(1), nil}}},
		{"person[0].age", &FieldAccess{"person", &ArrayAccess{Num(0), &FieldAccess{"age", nil}}}},
		{"person()[0]", &MethodCall{"person", []Expression{}, &ArrayAccess{Num(0), nil}}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.fieldAccess()
			if PrettyPrint(result) != PrettyPrint(d.expect) {
				t.Errorf("`%s` \n Got %#v instead of %#v", d.str, PrettyPrint(result), PrettyPrint(d.expect))
			}
		})
	}
}

func TestParser_arrayAccess_panic(t *testing.T) {
	data := []string{
		"arr[]",
		// "arr[0][]",
	}

	for _, str := range data {
		withParser(str, func(p *Parser) {
			msg := fmt.Sprintf("Should panic on: `%s`", str)
			defer assertPanic(t, msg)
			p.fieldAccess()
		})
	}
}

func fakeToken(str string, tt TokenType) Token {
	return newToken(1, 0, str, tt)
}
func TestParser_additiveExp(t *testing.T) {
	data := []struct {
		str string
		exp Expression
	}{
		{
			"1 + 2",
			&BinOp{fakeToken("+", Addition), Num(1), Num(2)},
		},

		{
			"1 + 2 - 3",
			&BinOp{fakeToken("+", Addition), Num(1), &BinOp{fakeToken("-", Subtraction), Num(2), Num(3)}},
		},

		{
			"1 + 2 * 3",
			&BinOp{fakeToken("+", Addition), Num(1), &BinOp{fakeToken("*", Multiplication), Num(2), Num(3)}},
		},

		{
			"2 * 3",
			&BinOp{fakeToken("*", Multiplication), Num(2), Num(3)},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.additiveExp()
			if r, e := PrettyPrint(result), PrettyPrint(d.exp); r != e {
				t.Errorf("Expected %s but got %s", e, r)
			}
		})
	}
}

func TestParser_multiplicativeExp(t *testing.T) {
	data := []struct {
		str string
		exp Expression
	}{
		{
			"1 / 2",
			&BinOp{fakeToken("/", Division), Num(1), Num(2)},
		},

		{
			"1 * 2 / 3",
			&BinOp{fakeToken("*", Multiplication), Num(1), &BinOp{fakeToken("/", Modulus), Num(2), Num(3)}},
		},

		{
			"height / width",
			&BinOp{fakeToken("/", Multiplication), &FieldAccess{"height", nil}, &FieldAccess{"width", nil}},
		},

		{
			"height / width[0]",
			&BinOp{fakeToken("/", Multiplication), &FieldAccess{"height", nil}, &FieldAccess{"width", &ArrayAccess{Num(0), nil}}},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.multiplicativeExp()
			if r, e := PrettyPrint(result), PrettyPrint(d.exp); r != e {
				t.Errorf("Expected %s but got %s", e, r)
			}
		})
	}
}
