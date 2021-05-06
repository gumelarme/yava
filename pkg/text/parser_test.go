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

func TestParser_program(t *testing.T) {
	helloClass := NewEmptyClass("Hello", nil, nil)
	greetInterface := &Interface{"Greet", nil}
	expect := Program{
		"Hello": helloClass,
		"Greet": greetInterface,
	}
	str := `class Hello {} interface Greet {}`
	withParser(str, func(p *Parser) {
		program := p.Compile()
		if !expect.Equal(program) {
			t.Errorf("Program is not equal %d, %d", len(program), len(expect))
		}
	})
}
func TestParser_interface(t *testing.T) {
	method1 := MethodSignature{Public, NamedType{"int", false}, "Count", []Parameter{}}
	method2 := MethodSignature{Public, NamedType{"String", false}, "Quack", []Parameter{}}
	int1 := NewInterface("Something")

	int2 := NewInterface("Something")
	int2.Methods[method1.Signature()] = &method1

	int3 := NewInterface("Something")
	int3.Methods[method2.Signature()] = &method2

	int4 := NewInterface("Something")
	int4.Methods[method1.Signature()] = &method1
	int4.Methods[method2.Signature()] = &method2

	data := []struct {
		str    string
		expect *Interface
	}{
		{
			`interface Something {}`,
			int1,
		},
		{
			`interface Something {public int Count();}`,
			int2,
		},
		{
			`interface Something {String Quack();}`,
			int3,
		},
		{
			`interface Something {int Count(); String Quack();}`,
			int4,
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			res := p.interfaceDeclaration()
			if resStr, expStr := PrettyPrint(res), PrettyPrint(d.expect); resStr != expStr {
				t.Errorf("Expecting \n%s \n----but got----\n%s", expStr, resStr)
			}
		})
	}
}

func TestParser_interface_panic(t *testing.T) {
	str := []string{
		"interface Something{int Count(); int Count();}",
	}
	for _, s := range str {
		withParser(s, func(p *Parser) {
			defer assertPanic(t, fmt.Sprintf("Should have panic on \n%s", s))
			p.interfaceDeclaration()
		})
	}

}

func TestParser_classExtends(t *testing.T) {
	classA := NewEmptyClass("A", nil, nil)
	classB := NewEmptyClass("B", classA, nil)
	program1 := Program{"A": classA, "B": classB}

	interfaceA := Interface{"A", nil}
	classC := NewEmptyClass("C", nil, &interfaceA)
	program2 := Program{"A": &interfaceA, "C": classC}

	classBStr := `class A {} class B extends A {}`
	classCStr := `interface A {} class C implements A {}`

	withParser(classBStr, func(p *Parser) {
		program := p.Compile()
		if b, c := PrettyPrint(&program1), PrettyPrint(program); b != c {
			t.Errorf("Expecting \n%s \n----got----\n %s", b, c)
		}
	})

	withParser(classCStr, func(p *Parser) {
		program := p.Compile()
		if b, c := PrettyPrint(&program2), PrettyPrint(program); b != c {
			t.Errorf("Expecting %s got %s", b, c)
		}
	})
}

func TestParser_class(t *testing.T) {
	class1 := NewEmptyClass("Hello", nil, nil)

	class2 := NewEmptyClass("Hello", nil, nil)
	class2Prop := PropertyDeclaration{Public,
		VariableDeclaration{
			NamedType{"int", false}, "a", Num(20),
		},
	}
	class2.Properties["a"] = &class2Prop

	class3 := NewEmptyClass("Hello", nil, nil)
	class3Method := NewMethodDeclaration(
		Private,
		NamedType{"void", false},
		"Nothing",
		[]Parameter{},
		StatementList{
			&JumpStatement{ReturnJump, nil},
		},
	)
	class3.Methods[class3Method.Name] = make(map[string]*MethodDeclaration)
	class3.Methods[class3Method.Name][class3Method.Signature()] = class3Method

	class4 := NewEmptyClass("Hello", nil, nil)
	class4Constructor := ConstructorDeclaration{*NewMethodDeclaration(
		Private,
		NamedType{"<this>", false},
		"Hello",
		[]Parameter{},
		StatementList{
			&JumpStatement{ReturnJump, nil},
		},
	)}
	class4.Constructor[class4Constructor.Signature()] = &class4Constructor

	class5 := NewEmptyClass("Hello", nil, nil)
	class5Main := MainMethodDeclaration{*NewMethodDeclaration(
		Public,
		NamedType{"void", false},
		"main",
		[]Parameter{
			{NamedType{"String", true}, "args"},
		},
		StatementList{
			&JumpStatement{ReturnJump, nil},
		},
	)}
	class5.MainMethod = &class5Main

	classCombined := NewEmptyClass("Hello", nil, nil)
	classCombined.Properties[class2Prop.GetName()] = &class2Prop

	classCombined.Methods[class3Method.Name] = make(map[string]*MethodDeclaration)
	classCombined.Methods[class3Method.Name][class3Method.Signature()] = class3Method

	classCombined.Constructor[class4Constructor.Signature()] = &class4Constructor
	classCombined.MainMethod = &class5Main

	classOverloading := NewEmptyClass("Hello", nil, nil)
	classOverloading.Methods[class3Method.Name] = make(map[string]*MethodDeclaration)
	classOverloading.Methods[class3Method.Name][class3Method.Signature()] = class3Method
	overLoadMethod := NewMethodDeclaration(
		Private,
		NamedType{"void", false},
		"Nothing",
		[]Parameter{
			{NamedType{"int", false}, "a"},
		},
		StatementList{
			&JumpStatement{ReturnJump, nil},
		},
	)
	classOverloading.Methods[class3Method.Name][overLoadMethod.Signature()] = overLoadMethod

	data := []struct {
		str    string
		expect *Class
	}{
		{
			"class Hello{}",
			class1,
		},
		{
			`
class Hello{
	int a = 20;
}
`,
			class2,
		},
		{
			`
class Hello{
	private void Nothing(){
		return;
	}
}
`,
			class3,
		},
		{
			`
class Hello{
	private Hello(){
		return;
	}
}
`,
			class4,
		},
		{
			`
class Hello{
	public static void main(String[] args){
		return;
	}
}
`,
			class5,
		},
		{
			`
class Hello{
	int a = 20;
	private Hello(){
		return;
	}
	private void Nothing(){
		return;
	}
	public static void main(String[] args){
		return;
	}
}
`,
			classCombined,
		},

		{
			`
class Hello{
	private void Nothing(){
		return;
	}

	private void Nothing(int a){
		return;
	}
}
`,
			classOverloading,
		},
	}
	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			res := p.classDeclaration()
			if resStr, expStr := PrettyPrint(res), PrettyPrint(d.expect); resStr != expStr {
				t.Errorf("Expecting \n%s \n----but got----\n%s", expStr, resStr)
			}
		})
	}
}

func TestParser_class_panic(t *testing.T) {
	data := []string{
		//duplicate property
		`class Something{ int a; int a;}`,
		`class Something{ int a; String a;}`,

		// member already defined as other thing
		`class Something{ int a; int a(){} }`,

		// duplicate method signature
		`class Something{
			String Speak(){}
			String Speak(){return 1;}
		}`,

		// method overloading, different type
		`class Something{
			String Speak(){}
			int Speak(){}
		}`,

		`class Something{
			String Speak(){}
			int Speak(int i){}
		}`,

		// method without return type
		`class Something{
			Nice(int a){}
		}`,

		// constructor overloading
		`class Something{
			Something(int a){}
			Something(int b){}
		}`,

		// main already defined
		`class Something{
			public static void main(String[] args){}
			public static void main(String[] a){}
		}`,

		`class Something{
			private static void main(String[] args){}
		}`,
	}

	for _, str := range data {
		withParser(str, func(p *Parser) {
			defer assertPanic(t, fmt.Sprintf("Should panic on \n%s", str))
			p.classDeclaration()
		})
	}
}

func TestParser_declaration(t *testing.T) {
	data := []struct {
		str    string
		expect Declaration
	}{
		{
			"int a;",
			&PropertyDeclaration{Public, VariableDeclaration{
				NamedType{"int", false},
				"a",
				nil,
			}},
		},
		{
			"int[] a;",
			&PropertyDeclaration{Public, VariableDeclaration{
				NamedType{"int", true},
				"a",
				nil,
			}},
		},
		{
			"public int a;",
			&PropertyDeclaration{Public, VariableDeclaration{
				NamedType{"int", false},
				"a",
				nil,
			}},
		},
		{
			"private int a = 1;",
			&PropertyDeclaration{Private, VariableDeclaration{
				NamedType{"int", false},
				"a",
				Num(1),
			}},
		},
		{
			`String a = "Hello";`,
			&PropertyDeclaration{Public, VariableDeclaration{
				NamedType{"String", false},
				"a",
				String("Hello"),
			}},
		},
		{
			`void foo(){}`,
			NewMethodDeclaration(Public,
				NamedType{"void", false},
				"foo",
				[]Parameter{},
				StatementList{},
			),
		},
		{
			`private void foo(){}`,
			NewMethodDeclaration(Private,
				NamedType{"void", false},
				"foo",
				[]Parameter{},
				StatementList{},
			),
		},
		{
			`private int foo(int a){}`,
			NewMethodDeclaration(Private,
				NamedType{"int", false},
				"foo",
				[]Parameter{
					{NamedType{"int", false}, "a"},
				},
				StatementList{},
			),
		},
		{
			`private String foo(int a, String[] list){
return 1;
}`,
			NewMethodDeclaration(Private,
				NamedType{"String", false},
				"foo",
				[]Parameter{
					{NamedType{"int", false}, "a"},
					{NamedType{"String", true}, "list"},
				},
				StatementList{
					&JumpStatement{ReturnJump, Num(1)},
				},
			),
		},
		{
			`public static void main(String[] args){
return;
}`,
			&MainMethodDeclaration{*NewMethodDeclaration(Public,
				NamedType{"void", false},
				"main",
				[]Parameter{
					{NamedType{"String", true}, "args"},
				},
				StatementList{
					&JumpStatement{ReturnJump, nil},
				},
			),
			}},
		{
			"Hello(){}",
			NewConstructor(Public, "Hello", []Parameter{}, StatementList{}),
		},
		{
			"public Hello(int who){ return 1;}",
			NewConstructor(
				Public,
				"Hello",
				[]Parameter{
					{NamedType{"int", false}, "who"},
				},
				StatementList{
					&JumpStatement{ReturnJump, Num(1)},
				},
			),
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			decl := p.declaration()
			if res, ex := PrettyPrint(decl), PrettyPrint(d.expect); res != ex {
				t.Errorf("From `%s` Expecting \n%s but got \n%s", d.str, ex, res)
			}
		})
	}
}

func TestParser_MainMethod(t *testing.T) {
	str := []string{
		"static int main()",
		"static void nani()",
		"static void main(String a)",
		"static void main(int[] a)",
	}

	for _, s := range str {
		withParser(s, func(p *Parser) {
			defer assertPanic(t, fmt.Sprintf("Should panic on %s", s))
			p.mainMethodDeclaration()
		})
	}
}

func TestParser_statement(t *testing.T) {
	data := []struct {
		str    string
		expect Statement
	}{
		{
			"int a = 20;",
			&VariableDeclaration{NamedType{"int", false}, "a", Num(20)},
		},
		{
			"int[] a = new int[20];",
			&VariableDeclaration{NamedType{"int", true},
				"a",
				&ArrayCreation{"int", Num(20)},
			},
		},
		{
			`String a = "nice";`,
			&VariableDeclaration{NamedType{"String", false}, "a", String("nice")},
		},
		{
			`this.a = nice;`,
			&AssignmentStatement{fakeToken("=", Assignment),
				&This{&FieldAccess{"a", nil}},
				&FieldAccess{"nice", nil},
			},
		},
		{
			`switch(a){
		case 2:
		if(a + b > 12) {
		return 3;
		}
		}`,
			&SwitchStatement{&FieldAccess{"a", nil},
				[]*CaseStatement{
					{
						Num(2),
						[]Statement{
							&IfStatement{
								&BinOp{fakeToken(">", GreaterThan),
									&BinOp{fakeToken("+", Addition),
										&FieldAccess{"a", nil},
										&FieldAccess{"b", nil},
									},
									Num(12),
								},
								StatementList{&JumpStatement{ReturnJump, Num(3)}},
								nil,
							},
						},
					},
				},
				nil,
			},
		},
	}
	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			decl := p.statement()
			if res, ex := PrettyPrint(decl), PrettyPrint(d.expect); res != ex {
				t.Errorf("From `%s` Expecting \n%s but got \n%s", d.str, ex, res)
			}
		})
	}
}

func TestParser_primitiveTypeVarDeclaration(t *testing.T) {
	data := []struct {
		str    string
		expect Statement
	}{
		{
			"int a = 20;",
			&VariableDeclaration{NamedType{"int", false}, "a", Num(20)},
		},
		{
			"int[] a = 20;",
			&VariableDeclaration{NamedType{"int", true}, "a", Num(20)},
		},
		{
			"boolean a = true;",
			&VariableDeclaration{NamedType{"boolean", false}, "a", Boolean(true)},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			whileStmt := p.primitiveTypeVarDeclaration()
			if res, ex := PrettyPrint(whileStmt), PrettyPrint(d.expect); res != ex {
				t.Errorf("From `%s` Expecting \n%s but got \n%s", d.str, ex, res)
			}
		})
	}
}

func TestParser_methodOrField(t *testing.T) {
	data := []struct {
		str    string
		expect Statement
	}{

		{
			"name = nice * method(1);",
			&AssignmentStatement{fakeToken("=", Assignment),
				&FieldAccess{"name", nil},
				&BinOp{fakeToken("*", Multiplication),
					&FieldAccess{"nice", nil},
					&MethodCall{"method", []Expression{Num(1)}, nil},
				},
			},
		},
		{
			"this.name = nice;",
			&AssignmentStatement{fakeToken("=", Assignment),
				&This{&FieldAccess{"name", nil}},
				&FieldAccess{"nice", nil},
			},
		},
		{
			"Hello();",
			&MethodCallStatement{&MethodCall{"Hello", []Expression{}, nil}},
		},
		{
			"this.person.hello();",
			&MethodCallStatement{
				&This{
					&FieldAccess{"person",
						&MethodCall{"hello", []Expression{}, nil},
					},
				},
			},
		},
		{
			"Something a = new Something();",
			&VariableDeclaration{NamedType{"Something", false}, "a",
				&ObjectCreation{MethodCall{"Something", []Expression{}, nil}},
			},
		},
		{
			"Something[] a = new Something[4];",
			&VariableDeclaration{NamedType{"Something", true}, "a",
				&ArrayCreation{"Something", Num(4)},
			},
		},
		{
			"something[0] = 20;",
			&AssignmentStatement{fakeToken("=", Assignment),
				&FieldAccess{"something", &ArrayAccess{Num(0), nil}},
				Num(20),
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			stmt := p.varDeclarationOrMethodOrAssignment()
			if res, ex := PrettyPrint(stmt), PrettyPrint(d.expect); res != ex {
				t.Errorf("From `%s` Expecting \n%s but got \n%s", d.str, ex, res)
			}
		})
	}

}
func TestParser_forStmt(t *testing.T) {
	data := []struct {
		str    string
		expect ForStatement
	}{
		{
			`for(;;){}`,
			ForStatement{nil, nil, nil, StatementList{}},
		},
		{
			`for(;;){
a += 1;
}`,
			ForStatement{nil, nil, nil, StatementList{
				&AssignmentStatement{
					fakeToken("+=", AdditionAssignment),
					&FieldAccess{"a", nil},
					Num(1),
				},
			}},
		},
		{
			`for(;;) for(;;) return null;`,
			ForStatement{nil, nil, nil, &ForStatement{nil, nil, nil, &JumpStatement{ReturnJump, Null{}}}},
		},
		{
			`for(int i = 0; i > 0; i += 1){}`,
			ForStatement{
				&VariableDeclaration{NamedType{"int", false}, "i", Num(0)},
				&BinOp{fakeToken(">", GreaterThan), &FieldAccess{"i", nil}, Num(0)},
				&AssignmentStatement{fakeToken("+=", AdditionAssignment), &FieldAccess{"i", nil}, Num(1)},
				StatementList{},
			},
		},
		{
			`for(i = 0; ; i += 1){}`,
			ForStatement{
				&AssignmentStatement{fakeToken("=", Assignment), &FieldAccess{"i", nil}, Num(0)},
				nil,
				&AssignmentStatement{fakeToken("+=", AdditionAssignment), &FieldAccess{"i", nil}, Num(1)},
				StatementList{},
			},
		},
		{
			`for(this.i = 0; ; i += 1){}`,
			ForStatement{
				&AssignmentStatement{fakeToken("=", Assignment), &This{&FieldAccess{"i", nil}}, Num(0)},
				nil,
				&AssignmentStatement{fakeToken("+=", AdditionAssignment), &FieldAccess{"i", nil}, Num(1)},
				StatementList{},
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			stmt := p.forStmt()
			if res, ex := PrettyPrint(stmt), PrettyPrint(&d.expect); res != ex {
				t.Errorf("Expecting \n%s but got \n%s", ex, res)
			}
		})
	}

}

func TestParser_whileStmt(t *testing.T) {
	data := []struct {
		str    string
		expect WhileStatement
	}{
		{
			"while(true) return 20;",
			WhileStatement{Boolean(true), &JumpStatement{ReturnJump, Num(20)}},
		},
		{
			`while(x > 0){
while(true) return 20;
}`,
			WhileStatement{&BinOp{fakeToken(">", GreaterThan), &FieldAccess{"x", nil}, Num(0)},
				StatementList{
					&WhileStatement{Boolean(true), &JumpStatement{ReturnJump, Num(20)}}},
			},
		},
		{
			`while(isOk) if(isStillOk) break;
`,
			WhileStatement{&FieldAccess{"isOk", nil},
				&IfStatement{
					&FieldAccess{"isStillOk", nil},
					&JumpStatement{BreakJump, nil},
					nil,
				},
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			stmt := p.whileStmt()
			if res, ex := PrettyPrint(stmt), PrettyPrint(&d.expect); res != ex {
				t.Errorf("Expecting \n%s but got \n%s", ex, res)
			}
		})
	}
}

func TestParser_ifStmt(t *testing.T) {
	data := []struct {
		str    string
		expect IfStatement
	}{
		{
			"if(name) return 20;",
			IfStatement{&FieldAccess{"name", nil}, &JumpStatement{ReturnJump, Num(20)}, nil},
		},

		{
			"if(name) return 20; else return 1;",
			IfStatement{&FieldAccess{"name", nil}, &JumpStatement{ReturnJump, Num(20)}, &JumpStatement{ReturnJump, Num(1)}},
		},
		{
			`if(name > 20){
break;
} else {
return 12;
}`,
			IfStatement{
				&BinOp{fakeToken(">", GreaterThan),
					&FieldAccess{"name", nil},
					Num(20),
				},
				&StatementList{
					&JumpStatement{BreakJump, nil},
				},
				&StatementList{
					&JumpStatement{ReturnJump, Num(12)},
				},
			},
		},

		{
			`if(name){
break;
} else if (isOk()){
return 12;
}`,
			IfStatement{&FieldAccess{"name", nil},
				&StatementList{
					&JumpStatement{BreakJump, nil},
				},
				// else if block
				&IfStatement{&MethodCall{"isOk", []Expression{}, nil},
					&StatementList{
						&JumpStatement{ReturnJump, Num(12)},
					},
					nil,
				},
			},
		},
		{
			`if(name){
break;
} else if (isOk()){
return 12;
} else {
return 1;
}`,
			IfStatement{&FieldAccess{"name", nil},
				&StatementList{
					&JumpStatement{BreakJump, nil},
				},
				// else if block
				&IfStatement{&MethodCall{"isOk", []Expression{}, nil},
					&StatementList{
						&JumpStatement{ReturnJump, Num(12)},
					},
					// else block
					&StatementList{
						&JumpStatement{ReturnJump, Num(1)},
					},
				},
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			ifres := p.ifStmt()
			if res, ex := PrettyPrint(ifres), PrettyPrint(&d.expect); res != ex {
				t.Errorf("Expecting \n%s but got \n%s", ex, res)
			}
		})
	}

}

func TestParser_switchStmt(t *testing.T) {
	data := []struct {
		str    string
		expect SwitchStatement
	}{
		{
			`switch(age){}`,
			SwitchStatement{&FieldAccess{"age", nil}, nil, nil},
		},

		{
			`switch(age){
case 12:
return 20;
}`,
			SwitchStatement{&FieldAccess{"age", nil},
				[]*CaseStatement{
					{Num(12), []Statement{&JumpStatement{ReturnJump, Num(20)}}},
				},
				nil,
			},
		},

		{
			`switch(age){
default:
return 20;
}`,
			SwitchStatement{&FieldAccess{"age", nil},
				nil,
				[]Statement{&JumpStatement{ReturnJump, Num(20)}},
			},
		},

		{
			`switch(age.year()){
case 1998:
return 20;
case 20:
return 2;
default:
break;
}`,
			SwitchStatement{&FieldAccess{"age", &MethodCall{"year", []Expression{}, nil}},
				[]*CaseStatement{
					{Num(1998), []Statement{&JumpStatement{ReturnJump, Num(20)}}},
					{Num(20), []Statement{&JumpStatement{ReturnJump, Num(2)}}},
				},
				[]Statement{&JumpStatement{BreakJump, nil}},
			},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			sw := p.switchStmt()
			if res, ex := PrettyPrint(sw), PrettyPrint(&d.expect); res != ex {
				t.Errorf("Expecting \n%s but got \n%s", ex, res)
			}
		})
	}
}

func TestParser_caseStmt(t *testing.T) {
	data := []struct {
		str    string
		expect CaseStatement
	}{
		{
			`case 1:
return;
`,
			CaseStatement{Num(1), []Statement{
				&JumpStatement{ReturnJump, nil},
			}},
		},

		{
			`case true:
		return 8;
		`,
			CaseStatement{Boolean(true), []Statement{
				&JumpStatement{ReturnJump, Num(8)},
			}},
		},

		{
			`case 8: {
return 900;
}
		`,
			CaseStatement{Num(8), []Statement{
				StatementList{
					&JumpStatement{ReturnJump, Num(900)},
				},
			}},
		},

		{
			`case 'A':
		return;
		`,
			CaseStatement{Char('A'), []Statement{
				&JumpStatement{ReturnJump, nil},
			}},
		},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.caseStmt()
			if r, expect := result.String(), d.expect.String(); r != expect {
				t.Errorf("Expecting \n%s but got\n%s", expect, r)
			}
		})
	}
}

func TestParser_jumpStmt(t *testing.T) {
	data := []struct {
		str    string
		expect Statement
	}{
		{"return;", &JumpStatement{ReturnJump, nil}},
		{"break;", &JumpStatement{BreakJump, nil}},
		{"continue;", &JumpStatement{ContinueJump, nil}},
		{"return 1;", &JumpStatement{ReturnJump, Num(1)}},
		{"return new Hello();",
			&JumpStatement{ReturnJump,
				&ObjectCreation{MethodCall{"Hello", []Expression{}, nil}},
			},
		},
	}
	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.jumpStmt()
			resStr, expectStr := PrettyPrint(result), PrettyPrint(d.expect)
			if resStr != expectStr {
				t.Errorf("Expecting: \n%s but got \n%s", expectStr, resStr)
			}
		})
	}
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

func TestParser_fieldAccessFrom(t *testing.T) {
	data := []struct {
		name   string
		str    string
		expect NamedValue
	}{
		{"person", "", &FieldAccess{"person", nil}},
		{"person", ".age", &FieldAccess{"person", &FieldAccess{"age", nil}}},
		{"person", "()", &MethodCall{"person", []Expression{}, nil}},
	}

	for _, d := range data {
		withParser(d.str, func(p *Parser) {
			result := p.fieldAccessFrom(d.name)
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
