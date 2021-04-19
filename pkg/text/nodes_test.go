package text

import "testing"

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
