package text

import (
	"fmt"
	"testing"
)

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
