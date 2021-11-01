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
			"(#field page (#array :at (#int 1)))",
		},

		{
			&FieldAccess{"page", &ArrayAccess{Num(0), &FieldAccess{"name", nil}}},
			"(#field page (#array :at (#int 0) (#field name)))",
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
			"(#method-call somemethod :args [(#int 1), (#int 2)])",
		},
		{
			&MethodCall{"somemethod", []Expression{}, &FieldAccess{"page", nil}},
			"(#method-call somemethod :args [] (#field page))",
		},
		{
			&MethodCall{"somemethod", []Expression{}, &ArrayAccess{Num(1), nil}},
			"(#method-call somemethod :args [] (#array :at (#int 1)))",
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
		{`(#int 123)`, NumFromStr("123")},
		{`(#boolean true)`, NewBoolean("true")},
		{`(#boolean false)`, NewBoolean("false")},
		{`(#char 'c')`, NewChar("c")},
		{`(#char '你')`, NewChar("你")},
		{`(#String "Hello")`, String("Hello")},
		{`(#String "Hello \"Bro\"")`, String(`Hello "Bro"`)},
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
			"(#case (#int 12) :do [])",
			CaseStatement{Num(12), []Statement{}},
		},
		{
			"(#case (#int 12) :do [(#return)])",
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
			"(#switch (#field age) :case [(#case (#int 12) :do [(#return)])])",
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
			"(#if (#field name) :body (#return) :else (#return (#int 1)))",
			IfStatement{&FieldAccess{"name", nil}, &JumpStatement{ReturnJump, nil}, &JumpStatement{ReturnJump, Num(1)}},
		},
		{
			"(#if (#field name) :body (#stmt-block (#if (#field what) :body (#return))) :else (#return (#int 1)))",
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

//TODO: Do more equality test
func TestMethodSignature_Equal(t *testing.T) {
	m1 := MethodSignature{Public, NamedType{"void", false}, "Hello", []Parameter{}}
	m2 := MethodSignature{Public, NamedType{"void", false}, "Hello", []Parameter{}}

	if !m2.Equal(m1) {
		t.Errorf("Method signature should be equal")
	}

	m3 := MethodSignature{Public, NamedType{"int", true}, "getAge", []Parameter{}}
	m4 := MethodSignature{Public, NamedType{"int", true}, "getAge", []Parameter{{NamedType{"int", false}, "a"}}}

	if m3.Equal(m4) {
		t.Errorf("Method signature with different parameter count should be unequal")
	}

	m5 := MethodSignature{Public, NamedType{"int", true}, "getName", []Parameter{}}
	m6 := MethodSignature{Public, NamedType{"int", true}, "getAge", []Parameter{}}

	if m5.Equal(m6) {
		t.Errorf("Method signature with different name should be unequal")
	}

	m7 := MethodSignature{Public, NamedType{"int", true}, "getAge", []Parameter{
		{NamedType{"int", false}, "a"},
	}}

	m8 := MethodSignature{Public, NamedType{"int", true}, "getAge", []Parameter{
		{NamedType{"char", false}, "a"},
	}}

	if m7.Equal(m8) {
		t.Errorf("Method signature with different parameter list should be unequal")
	}
}

func TestClass_Members(t *testing.T) {
	class := NewEmptyClass("Person", "", "")
	prop := &PropertyDeclaration{
		Public,
		VariableDeclaration{
			NamedType{"int", false},
			"age",
			nil,
		},
	}

	method := &MethodDeclaration{
		MethodSignature{
			Public,
			NamedType{"int", false},
			"getAge",
			[]Parameter{},
		},
		nil,
	}

	class.addProperty(prop)
	class.addMethod(method)
	method2 := &MethodDeclaration{
		MethodSignature{
			Public,
			NamedType{"int", false},
			"getAge",
			[]Parameter{
				{NamedType{"int", false}, "a"},
			},
		},
		nil,
	}
	class.addMethod(method2)

	expect := []Declaration{prop, method, method2}
	members := class.Members()

	if lex, lem := len(expect), len(members); lex != lem {
		t.Errorf("Number of member does not match, expect %d got %d", lex, lem)
	}

	for i, ex := range expect {
		mem := members[i]
		if mem.GetName() != ex.GetName() ||
			mem.GetAccessModifier() != ex.GetAccessModifier() ||
			mem.DeclType() != ex.DeclType() ||
			mem.TypeOf() != ex.TypeOf() {

			t.Errorf("Members are not equal")

		}

	}
}

//TODO: Move checking to node_visitor
// func TestClass_checkInterfaceImplementations(t *testing.T) {
// 	interfaceA := NewInterface("A")
// 	class1 := NewEmptyClass("Person", nil, nil)

// 	if err := class1.checkInterfaceImplementations(); err != nil {
// 		t.Error("Should return nil if class does not implement any interface.")
// 	}

// 	sign1 := MethodSignature{Public, NamedType{"int", false}, "getA", []Parameter{}}
// 	interfaceA.AddMethod(&sign1)

// 	class1.Implement = interfaceA
// 	class1.addMethod(&MethodDeclaration{sign1, StatementList{}})

// 	if err := class1.checkInterfaceImplementations(); err != nil {
// 		t.Errorf("Methods are implemented but got error of %s", err)
// 	}

// 	sign2 := MethodSignature{Public, NamedType{"int", true}, "getB", []Parameter{}}
// 	interfaceA.AddMethod(&sign2)

// 	if err := class1.checkInterfaceImplementations(); err == nil {
// 		t.Errorf("Methods are NOT implemented but got no errors")
// 	}

// 	class1.addMethod(&MethodDeclaration{
// 		MethodSignature{
// 			Public,
// 			NamedType{"int", true},
// 			"getB",
// 			[]Parameter{
// 				{NamedType{"int", false}, "b"},
// 			},
// 		},
// 		StatementList{},
// 	})

// 	if err := class1.checkInterfaceImplementations(); err == nil {
// 		t.Errorf("Same method name with different parameter should return error.")
// 	}
// }
