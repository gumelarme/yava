package text

import (
	"fmt"
	"testing"
)

func TestNamedValue_IdEndAs(t *testing.T) {
	data := []struct {
		expect string
		obj    NamedValue
	}{
		{
			"FieldAccess",
			&FieldAccess{"hello", nil},
		},
		{
			"MethodCall",
			&FieldAccess{"hello", &MethodCall{"nice", []Expression{}, nil}},
		},
		{
			"This",
			&This{},
		},
	}

	for _, d := range data {
		if end := IdEndsAs(d.obj); end != d.expect {
			t.Errorf("Expect %s but got %s", d.expect, end)
		}
	}
}
func TestFieldAccess_PrettyPrint(t *testing.T) {
	data := []struct {
		obj NamedValue
		str string
	}{
		{
			&FieldAccess{"page", nil},
			"(#field page)",
		},

		{
			&FieldAccess{"page", &FieldAccess{"name", nil}},
			"(#field page (#field name))",
		},
	}

	for _, d := range data {
		if PrettyPrint(d.obj) != d.str {
			t.Errorf("Expecting %s but got %s", d.str, PrettyPrint(d.obj))
		}
	}

}

func TestArrayAccess_PrettyPrint(t *testing.T) {
	data := []struct {
		obj NamedValue
		str string
	}{
		{
			&FieldAccess{"page", &ArrayAccess{Num(1), nil}},
			"(#field page (#array :at (#num 1)))",
		},

		{
			&FieldAccess{"page", &ArrayAccess{Num(0), &FieldAccess{"name", nil}}},
			"(#field page (#array :at (#num 0) (#field name)))",
		},
	}

	for _, d := range data {
		if PrettyPrint(d.obj) != d.str {
			t.Errorf("Expecting %s but got %s", d.str, PrettyPrint(d.obj))
		}
	}
}

func TestMethodCall_PrettyPrint(t *testing.T) {
	data := []struct {
		obj NamedValue
		str string
	}{
		{
			&MethodCall{"somemethod", []Expression{Num(1), Num(2)}, nil},
			"(#method-call somemethod :args [(#num 1), (#num 2)])",
		},
		{
			&MethodCall{"somemethod", []Expression{}, &FieldAccess{"page", nil}},
			"(#method-call somemethod :args [] (#field page))",
		},
		{
			&MethodCall{"somemethod", []Expression{}, &ArrayAccess{Num(1), nil}},
			"(#method-call somemethod :args [] (#array :at (#num 1)))",
		},

		{
			&FieldAccess{"page", &MethodCall{"somemethod", []Expression{}, &FieldAccess{"page", nil}}},
			"(#field page (#method-call somemethod :args [] (#field page)))",
		},
	}

	for _, d := range data {
		if PrettyPrint(d.obj) != d.str {
			t.Errorf("Expecting %s but got %s", d.str, PrettyPrint(d.obj))
		}
	}
}

func TestBasicType_PrettyPrint(t *testing.T) {
	data := []struct {
		str string
		obj Expression
	}{
		{`(#num 123)`, NumFromStr("123")},
		{`(#boolean true)`, NewBoolean("true")},
		{`(#boolean false)`, NewBoolean("false")},
		{`(#char 'c')`, NewChar("c")},
		{`(#char '你')`, NewChar("你")},
		{`(#string "Hello")`, String("Hello")},
		{`(#string "Hello \"Bro\"")`, String(`Hello "Bro"`)},
	}

	for _, d := range data {
		if pretty := PrettyPrint(d.obj); pretty != d.str {
			name, _ := d.obj.NodeContent()
			t.Errorf("%s is expected to return %s instead of %s",
				name,
				pretty,
				d.str,
			)
		}
	}
}

func TestNewBoolean_panic(t *testing.T) {
	data := "nice"
	msg := fmt.Sprintf("NewBoolean should panic on %#v", data)
	defer assertPanic(t, msg)
	NewBoolean(data)
}

func TestNewChar_panic(t *testing.T) {
	data := "morethanonechar"
	msg := fmt.Sprintf("NewChar should panic on %#v", data)
	defer assertPanic(t, msg)
	NewChar(data)
}

func TestCaseStatement_String(t *testing.T) {
	data := []struct {
		str  string
		node CaseStatement
	}{
		{
			"(#case (#num 12) :do [])",
			CaseStatement{Num(12), []Statement{}},
		},
		{
			"(#case (#num 12) :do [(#return)])",
			CaseStatement{Num(12), []Statement{&JumpStatement{ReturnJump, nil}}},
		},
		{
			"(#case (#char 'c') :do [(#break)])",
			CaseStatement{Char('c'), []Statement{&JumpStatement{BreakJump, nil}}},
		},
	}
	for _, d := range data {
		if res := d.node.String(); res != d.str {
			t.Errorf("Expecting \n%s but got \n%s", d.str, res)
		}
	}
}

func TestSwitchStatement_PrettyPrint(t *testing.T) {
	data := []struct {
		str  string
		node *SwitchStatement
	}{
		{
			"(#switch (#field age) :case [(#case (#num 12) :do [(#return)])])",
			&SwitchStatement{&FieldAccess{"age", nil},
				[]*CaseStatement{
					{Num(12), []Statement{&JumpStatement{ReturnJump, nil}}},
				},
				nil,
			},
		},
		{
			"(#switch (#field age) :case [] :default [(#return)])",
			&SwitchStatement{&FieldAccess{"age", nil},
				[]*CaseStatement{},
				[]Statement{&JumpStatement{ReturnJump, nil}},
			},
		},
	}
	for _, d := range data {
		if res := PrettyPrint(d.node); res != d.str {
			t.Errorf("Expecting \n%s but got \n%s", d.str, res)
		}
	}
}

func TestIfStatement_PrettyPrint(t *testing.T) {
	data := []struct {
		str  string
		node IfStatement
	}{
		{
			"(#if (#field name) :body (#return))",
			IfStatement{&FieldAccess{"name", nil}, &JumpStatement{ReturnJump, nil}, nil},
		},
		{
			"(#if (#field name) :body (#return) :else (#return (#num 1)))",
			IfStatement{&FieldAccess{"name", nil}, &JumpStatement{ReturnJump, nil}, &JumpStatement{ReturnJump, Num(1)}},
		},
		{
			"(#if (#field name) :body (#stmt-block (#if (#field what) :body (#return))) :else (#return (#num 1)))",
			IfStatement{&FieldAccess{"name", nil},
				&StatementList{
					&IfStatement{&FieldAccess{"what", nil}, &JumpStatement{ReturnJump, nil}, nil},
				},
				&JumpStatement{ReturnJump, Num(1)}},
		},
	}

	for _, d := range data {
		if res := PrettyPrint(&d.node); res != d.str {
			t.Errorf("Expecting \n%s but got \n%s", d.str, res)
		}
	}
}
