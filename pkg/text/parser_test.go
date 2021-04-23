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

func TestParser_expression(t *testing.T) {
	data := []struct {
		str    string
		expect Expression
	}{
		{"2", Num(2)},
		{"(2)", Num(2)},
		{"(2 + 3)", &BinOp{fakeToken("+", Addition), Num(2), Num(3)}},
		{"(2 + 3) * 4",
			&BinOp{fakeToken("*", Multiplication),
				&BinOp{fakeToken("+", Addition), Num(2), Num(3)},
				Num(4),
			},
		},
		{"2 + 3 * 4",
			&BinOp{fakeToken("+", Addition), Num(2),
				&BinOp{fakeToken("*", Multiplication), Num(3), Num(4)},
			},
		},
		{"(2 > 3) && method()",
			&BinOp{fakeToken("&&", And),
				&BinOp{fakeToken(">", GreaterThan), Num(2), Num(3)},
				&MethodCall{"method", []Expression{}, nil},
			},
		},
		{
			"new Foo()",
			&ObjectCreation{MethodCall{"Foo", []Expression{}, nil}},
		},
		{
			"new Foo[getNumber()]",
			&ArrayCreation{"Foo", &MethodCall{"getNumber", []Expression{}, nil}},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.expression()
			resStr, expStr := PrettyPrint(result), PrettyPrint(d.expect)
			if resStr != expStr {
				t.Errorf("Expecting \n%s but got\n%s", expStr, resStr)
			}
		})
	}
}

func TestParser_conditionalOrExp(t *testing.T) {
	data := []struct {
		str    string
		expect Expression
	}{
		{"true", Boolean(true)},
		{"true || false", &BinOp{fakeToken("||", Or), Boolean(true), Boolean(false)}},
		{
			"true || false && true",
			&BinOp{fakeToken("||", Or),
				Boolean(true),
				&BinOp{fakeToken("&&", And), Boolean(false), Boolean(true)},
			},
		},
		{
			"true && false || true",
			&BinOp{fakeToken("||", Or),
				&BinOp{fakeToken("&&", And), Boolean(true), Boolean(false)},
				Boolean(true),
			},
		},
		{
			"true && false && true || false",
			&BinOp{fakeToken("||", Or),
				&BinOp{fakeToken("&&", And),
					Boolean(true),
					&BinOp{fakeToken("&&", And),
						Boolean(false),
						Boolean(true),
					},
				},
				Boolean(false),
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.conditionalOrExp()
			resStr, expStr := PrettyPrint(result), PrettyPrint(d.expect)

			if resStr != expStr {
				t.Errorf("Expect \n%s but got \n%s", expStr, resStr)
			}
		})
	}
}

func TestParser_conditionalAndExp(t *testing.T) {
	data := []struct {
		str    string
		expect Expression
	}{
		{"true", Boolean(true)},
		{"true && false", &BinOp{fakeToken("&&", And), Boolean(true), Boolean(false)}},
		{
			"a > b && method()",
			&BinOp{fakeToken("&&", And),
				&BinOp{fakeToken(">", GreaterThan), &FieldAccess{"a", nil}, &FieldAccess{"b", nil}},
				&MethodCall{"method", []Expression{}, nil}},
		},
		//chained
		{
			"true && false && method()",
			&BinOp{fakeToken("&&", And),
				Boolean(true),
				&BinOp{fakeToken("&&", And),
					Boolean(false),
					&MethodCall{"method", []Expression{}, nil}},
			}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.conditionalAndExp()
			resStr, expStr := PrettyPrint(result), PrettyPrint(d.expect)
			if resStr != expStr {
				t.Errorf("Expect \n\t%s but got \n\t%s", expStr, resStr)
			}
		})
	}
}

func TestParser_relationalExp(t *testing.T) {
	data := []struct {
		str    string
		expect Expression
	}{
		{"12", Num(12)},
		{"12 == 12", &BinOp{fakeToken("==", Equal), Num(12), Num(12)}},
		{"1 >= 2", &BinOp{fakeToken(">=", GreaterThanEqual), Num(1), Num(2)}},
		{`true != this.status`, &BinOp{fakeToken("!=", NotEqual), Boolean(true), &This{&FieldAccess{"status", nil}}}},
	}
	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.relationalExp()
			resStr, expStr := PrettyPrint(result), PrettyPrint(d.expect)

			if resStr != expStr {
				t.Errorf("Expecting %s instead of %s", expStr, resStr)
			}
		})
	}
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

		{
			"this.height / width",
			&BinOp{fakeToken("/", Multiplication), &This{&FieldAccess{"height", nil}}, &FieldAccess{"width", nil}},
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

func TestParser_validName(t *testing.T) {
	data := []struct {
		str string
		exp NamedValue
	}{
		{
			"this.person",
			&This{&FieldAccess{"person", nil}},
		},

		{
			"this.person()",
			&This{&MethodCall{"person", []Expression{}, nil}},
		},

		{
			"person",
			&FieldAccess{"person", nil},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.validName()
			if r, e := PrettyPrint(result), PrettyPrint(d.exp); r != e {
				t.Errorf("Expected %s but got %s", e, r)
			}
		})
	}
}

func TestParser_primaryExp(t *testing.T) {
	data := []struct {
		str string
		exp Expression
	}{
		{"123", Num(123)},
		{`"Hello"`, String("Hello")},
		{"true", NewBoolean("true")},
		{"false", NewBoolean("false")},
		{"'c'", NewChar("c")},
		{"'你'", NewChar("你")},
		{"null", Null{}},
		{"name", &FieldAccess{"name", nil}},
		{"this.name", &This{&FieldAccess{"name", nil}}},
		{"this.name()", &This{&MethodCall{"name", []Expression{}, nil}}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			exp := p.primaryExp()
			result, expstr := PrettyPrint(exp), PrettyPrint(d.exp)
			if result != expstr {
				t.Errorf("PrimaryExp expecting %s but got %s",
					expstr,
					result,
				)
			}
		})
	}
}

func TestParser_objectInitialization(t *testing.T) {
	data := []struct {
		str string
		exp Expression
	}{
		{
			"new Hello()",
			&ObjectCreation{MethodCall{"Hello", []Expression{}, nil}},
		},

		{
			"new Nice(1, 2)",
			&ObjectCreation{MethodCall{"Nice", []Expression{Num(1), Num(2)}, nil}},
		},

		{
			"new int[6]",
			&ArrayCreation{"int", Num(6)},
		},

		{
			"new Hello[6 + 12]",
			&ArrayCreation{"Hello", &BinOp{fakeToken("+", Addition), Num(6), Num(12)}},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.objectInitialization()
			if r, e := PrettyPrint(result), PrettyPrint(d.exp); r != e {
				t.Errorf("Expected %s but got %s", e, r)
			}
		})
	}
}
