package lang

import (
	"fmt"
	"testing"

	"github.com/gumelarme/yava/pkg/text"
)

func Test_codeBoolean(t *testing.T) {
	expect := "iconst_1"
	got := codeBoolean(text.Boolean(true))
	if got != expect {
		t.Errorf("Expecting %#v on true value but got %#v", expect, got)
	}

	expect = "iconst_0"
	got = codeBoolean(text.Boolean(false))
	if got != expect {
		t.Errorf("Expecting %#v on false value but got %#v", expect, got)
	}
}

func Test_codeInt(t *testing.T) {
	data := []struct {
		num    int
		expect string
	}{
		{0, "iconst_0"},
		{3, "iconst_3"},
		{5, "iconst_5"},
		{6, "bipush 6"},
		{120, "bipush 120"},
		{255, "ldc 255"},
		{256, "ldc 256"},
		{1000_000, "ldc 1000000"},
	}

	for _, d := range data {
		result := codeInt(text.Num(d.num))
		if d.expect != result {
			t.Errorf("%d should coverted to %#v but got %#v", d.num, d.expect, result)
		}
	}
}

func Test_codeChar(t *testing.T) {
	data := []struct {
		char   rune
		expect string
	}{
		{'a', "bipush 97"},
		{'A', "bipush 65"},
		{'\u0053', "bipush 83"},
		{'你', "ldc 20320"},
	}

	for _, d := range data {
		result := codeChar(text.Char(d.char))
		if d.expect != result {
			t.Errorf("%c should coverted to %#v but got %#v", d.char, d.expect, result)
		}
	}
}

func Test_codeString(t *testing.T) {
	data := []struct {
		text   string
		expect string
	}{
		{"Hello", `ldc "Hello"`},
		{"你好", `ldc "你好"`},
	}

	for _, d := range data {
		result := codeString(text.String(d.text))
		if d.expect != result {
			t.Errorf("%s should coverted to %#v but got %#v", d.text, d.expect, result)
		}
	}
}

func Test_codeConstant(t *testing.T) {
	data := []struct {
		exp    text.Expression
		expect string
	}{
		{text.Num(1), "iconst_1"},
		{text.Boolean(true), "iconst_1"},
		{text.Char('a'), "bipush 97"},
		{text.String("Nice"), `ldc "Nice"`},
		{text.Null{}, "aconst_null"},
	}

	for _, d := range data {
		result := codeConstant(d.exp)
		if d.expect != result {
			content := text.PrettyPrint(d.exp)
			t.Errorf("%s should return %#v but got %#v", content, d.expect, result)
		}
	}
}

func Test_fieldDescriptor(t *testing.T) {
	data := []struct {
		name    string
		isArray bool
		expect  string
	}{
		{"int", false, "I"},
		{"boolean", false, "Z"},
		{"char", true, "[C"},
		{"String", false, "Ljava/lang/String;"},
		{"Hello", false, "LHello;"},
		{"AnyOtherElse", true, "[LAnyOtherElse;"},
		{"void", true, "V"},
	}

	for _, d := range data {
		result := fieldDescriptor(d.name, d.isArray)
		if result != d.expect {
			t.Errorf("Data type (%s:%v) expecting to result in %#v but got %#v ", d.name, d.isArray, d.expect, result)
		}
	}
}

func assertHasNCode(t *testing.T, gen *KrakatauGen, count int) bool {
	if codeCount := len(gen.Codes()); codeCount != count {
		t.Errorf("Should at least has %d line of codes, but got %d.", count, codeCount)
		return false
	}
	return true
}

func assertHasSameCodes(t *testing.T, gen *KrakatauGen, expect ...string) {
	if !assertHasNCode(t, gen, len(expect)) {
		for _, code := range gen.Codes() {
			t.Log(code)
		}
		return
	}

	codes := gen.Codes()
	for i, code := range expect {
		if code != codes[i] {
			t.Errorf("Expecting: \n%#v but got %#v", code, codes[i])
		}
	}
}

func TestKrakatauGen_Program(t *testing.T) {
	var program text.Program
	generator := NewEmptyKrakatauGen()
	program.Accept(generator)

	assertHasNCode(t, generator, 1)

	result := generator.codes[0]
	expect := fmt.Sprintf(".version %d %d", MajorVersion, MinorVersion)

	if result != expect {
		t.Errorf("Program should generate version number of: %#v but got %#v", result, expect)
	}
}

func TestKrakatauGen_Class(t *testing.T) {
	objectSuper := ".super java/lang/Object"
	data := []struct {
		class  *text.Class
		expect []string
	}{
		{
			text.NewEmptyClass("Person", "", ""),
			[]string{
				".class Person",
				objectSuper,
			},
		},

		{
			text.NewEmptyClass("Person", "Hello", ""),
			[]string{
				".class Person",
				".super Hello",
			},
		},
		{
			text.NewEmptyClass("Person", "", "ICallable"),
			[]string{
				".class Person",
				objectSuper,
				".implements ICallable",
			},
		},
		{
			text.NewEmptyClass("Person", "SomeObject", "ICallable"),
			[]string{
				".class Person",
				".super SomeObject",
				".implements ICallable",
			},
		},
	}

	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		gen.VisitClass(d.class)
		assertHasSameCodes(t, gen, d.expect...)
	}
}

func TestKrakatauGen_AfterClass(t *testing.T) {
	var class text.Class
	class.Name = "Mock"

	gen := NewEmptyKrakatauGen()
	gen.VisitAfterClass(&class)
	assertHasSameCodes(t, gen,
		".method <init> : ()V",
		".code stack 1 locals 1",
		"aload_0",
		InvokeJavaObject,
		"return",
		".end code",
		".end method",
		".end class",
	)
}

func TestKrakatauGen_PropertyDeclaration(t *testing.T) {
	data := []struct {
		prop   text.PropertyDeclaration
		expect []string
	}{
		{
			text.PropertyDeclaration{
				AccessModifier: text.Public,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "int", IsArray: false},
					Name:  "age",
					Value: nil,
				},
			},
			[]string{
				".field public age I",
			},
		},
		{
			text.PropertyDeclaration{
				AccessModifier: text.Public,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "int", IsArray: true},
					Name:  "age",
					Value: nil,
				},
			},
			[]string{
				".field public age [I",
			},
		},
		{
			text.PropertyDeclaration{
				AccessModifier: text.Private,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "String", IsArray: false},
					Name:  "name",
					Value: nil,
				},
			},
			[]string{
				".field private name Ljava/lang/String;",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			gen.VisitPropertyDeclaration(&d.prop)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakataugGen_makeDefaultConstructor(t *testing.T) {
	className := "Mock"
	mockKrakatau(func(gen *KrakatauGen) {
		typeTable := NewTypeAnalyzer().table
		gen.typeTable = typeTable
		gen.makeDefaultConstructor(className,
			&text.PropertyDeclaration{
				AccessModifier: text.Public,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "int", IsArray: false},
					Name:  "intProp",
					Value: nil,
				},
			},
			&text.PropertyDeclaration{
				AccessModifier: text.Public,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "String", IsArray: false},
					Name:  "stringProp",
					Value: nil,
				},
			},
			&text.PropertyDeclaration{
				AccessModifier: text.Public,
				VariableDeclaration: text.VariableDeclaration{
					Type:  text.NamedType{Name: "int", IsArray: false},
					Name:  "age",
					Value: text.Num(4000),
				},
			},
		)

		assertHasSameCodes(t, gen,
			".method <init> : ()V",
			".code stack 2 locals 1",
			"aload_0",
			InvokeJavaObject,
			"aload_0",
			"iconst_0",
			"putfield Field Mock intProp I",
			"aload_0",
			"aconst_null",
			"putfield Field Mock stringProp Ljava/lang/String;",
			"aload_0",
			"ldc 4000",
			"putfield Field Mock age I",
			"return",
			".end code",
			".end method",
		)
	})
}

func TestKrakatauGen_AfterBinOp(t *testing.T) {
	var add, sub, div, mul, mod text.Token
	add.Type = text.Addition
	sub.Type = text.Subtraction
	div.Type = text.Division
	mul.Type = text.Multiplication
	mod.Type = text.Modulus

	multiplication := text.NewBinOp(mul, text.Num(3), text.Num(4))
	data := []struct {
		binaryOp text.BinOp
		expect   []string
		stackMax int
	}{
		{
			text.NewBinOp(add, text.Num(12000), text.Num(3)),
			[]string{
				"ldc 12000",
				"iconst_3",
				"iadd",
			},
			2,
		},
		{
			text.NewBinOp(mod, text.Num(12), text.Num(3)),
			[]string{
				"bipush 12",
				"iconst_3",
				"irem",
			},
			2,
		},

		{
			text.NewBinOp(div, text.Num(12), &multiplication),
			[]string{
				"bipush 12",
				"iconst_3",
				"iconst_4",
				"imul",
				"idiv",
			},
			3,
		},
	}

	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		d.binaryOp.Accept(gen)
		assertHasSameCodes(t, gen, d.expect...)
		if gen.stackMax != d.stackMax {
			t.Errorf("Stack size should be %d, but got %d", d.stackMax, gen.stackMax)
		}
	}
}

func TestKrakatauGen_AfterBinOp_boolean(t *testing.T) {
	var gt, gte, lt, lte, eq, neq text.Token
	gt.Type = text.GreaterThan
	gte.Type = text.GreaterThanEqual
	lt.Type = text.LessThan
	lte.Type = text.LessThanEqual
	eq.Type = text.Equal
	neq.Type = text.NotEqual

	hundredsOp := text.NewBinOp(neq, text.Num(300), text.Num(200))
	data := []struct {
		bin    text.BinOp
		expect []string
	}{
		{
			text.NewBinOp(gt, text.Num(1), text.Num(2)),
			[]string{
				"iconst_1",
				"iconst_2",
				"if_icmpgt L0",
				"iconst_0",
				"goto L1",
				"L0:\ticonst_1",
				"L1:\t",
			},
		},
		{
			text.NewBinOp(eq, text.Num(1), &hundredsOp),
			[]string{
				"iconst_1",
				"ldc 300",
				"ldc 200",
				"if_icmpne L0",
				"iconst_0",
				"goto L1",
				"L0:\ticonst_1",
				"L1:\t",
				"if_icmpeq L2",
				"iconst_0",
				"goto L3",
				"L2:\ticonst_1",
				"L3:\t",
			},
		},
	}

	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		d.bin.Accept(gen)
		assertHasSameCodes(t, gen, d.expect...)
	}
}

func TestKrakatauGen_MethodSignature(t *testing.T) {
	getAge := methodGetAge.MethodSignature
	getName := methodGetNameWithParam.MethodSignature
	getName.AccessModifier = text.Private
	getName.ReturnType.Name = "String"
	getName.ParameterList = []text.Parameter{
		{Type: text.NamedType{Name: "String", IsArray: false}, Name: "name"},
	}

	method1 := text.MethodSignature{
		AccessModifier: text.Protected,
		ReturnType:     text.NamedType{Name: "void", IsArray: false},
		Name:           "things",
		ParameterList: []text.Parameter{
			{Type: text.NamedType{Name: "int", IsArray: false}, Name: "a"},
			{Type: text.NamedType{Name: "int", IsArray: true}, Name: "b"},
			{Type: text.NamedType{Name: "boolean", IsArray: false}, Name: "c"},
			{Type: text.NamedType{Name: "Human", IsArray: false}, Name: "d"},
		},
	}

	data := []struct {
		signature text.MethodSignature
		expect    string
	}{

		{
			getAge,
			".method public getAge : ()I",
		},
		{
			getName,
			".method private getName : (Ljava/lang/String;)Ljava/lang/String;",
		},
		{
			method1,
			".method protected things : (I[IZLHuman;)V",
		},
	}
	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		gen.VisitMethodSignature(&d.signature)
		assertHasSameCodes(t, gen, d.expect)
	}
}

func TestKrakatauGen_AfterMethodDeclaration(t *testing.T) {
	gen := NewEmptyKrakatauGen()
	gen.VisitAfterMethodDeclaration(nil)
	assertHasSameCodes(t, gen,
		".code stack 0 locals 0",
		"return",
		".end code",
		".end method",
	)

	gen = NewEmptyKrakatauGen()
	gen.stackMax = 3
	gen.localCount = 4
	gen.codeBuffer = []string{
		"bipush 11",
		"bipush 12",
		"iadd",
		"ireturn",
	}

	gen.VisitAfterMethodDeclaration(nil)
	assertHasSameCodes(t, gen,
		".code stack 3 locals 4",
		"bipush 11",
		"bipush 12",
		"iadd",
		"ireturn",
		".end code",
		".end method",
	)

}

func TestKrakatauGen_MainMethodDeclaration(t *testing.T) {
	gen := NewEmptyKrakatauGen()
	gen.VisitMainMethodDeclaration(nil)
	assertHasSameCodes(t, gen, ".method public static main : ([Ljava/lang/String;)V")
}

func TestKrakatauGen_stackSize(t *testing.T) {
	gen := NewEmptyKrakatauGen()
	gen.incStackSize(3)
	gen.incStackSize(1)
	if gen.stackSize != 4 && gen.stackMax != 4 {
		t.Errorf("Expecting stack size = 4, stack max = 4, but got %d and %d instead.", gen.stackSize, gen.stackMax)
	}

	gen.decStackSize(2)
	if gen.stackSize != 2 && gen.stackMax != 4 {
		t.Errorf("Expecting stack size = 2, stack max = 4, but got %d and %d instead.", gen.stackSize, gen.stackMax)
	}

	gen.resetStackSize()
	if gen.stackSize != 0 && gen.stackMax != 0 {
		t.Errorf("Expecting stack size & stack max to be 0, but got %d and %d instead.", gen.stackSize, gen.stackMax)
	}
}

func mockSysout(args ...text.Expression) *text.FieldAccess {
	return &text.FieldAccess{
		Name: "System",
		Child: &text.FieldAccess{
			Name: "out",
			Child: &text.MethodCall{
				Name:  "println",
				Args:  args,
				Child: nil,
			},
		},
	}
}

func TestKrakatauGen_SystemOut(t *testing.T) {
	var gt text.Token
	gt.Type = text.GreaterThan
	condition := text.NewBinOp(gt, text.Num(1), text.Num(2))

	data := []struct {
		lastDT DataType
		arg    text.Expression
		expect []string
	}{
		{
			mockInt,
			text.Num(1),
			[]string{
				"getstatic Field java/lang/System out Ljava/io/PrintStream;",
				"iconst_1",
				"invokevirtual Method java/io/PrintStream println (I)V",
			},
		},
		{
			mockString,
			text.String("Hello"),
			[]string{
				"getstatic Field java/lang/System out Ljava/io/PrintStream;",
				`ldc "Hello"`,
				"invokevirtual Method java/io/PrintStream println (Ljava/lang/String;)V",
			},
		},
		{
			mockBoolean,
			&condition,
			[]string{
				"getstatic Field java/lang/System out Ljava/io/PrintStream;",
				"iconst_1",
				"iconst_2",
				"if_icmpgt L0",
				"iconst_0",
				"goto L1",
				"L0:\ticonst_1",
				"L1:\t",
				"invokevirtual Method java/io/PrintStream println (Z)V",
			},
		},
	}
	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		gen.typeStack.Push(d.lastDT)
		sys := mockSysout(d.arg)
		sys.Accept(gen)
		assertHasSameCodes(t, gen, d.expect...)
	}
}

func TestKrakatauGen_JumpStatement(t *testing.T) {
	data := []struct {
		jump   text.JumpStatement
		expect []string
	}{
		{
			text.JumpStatement{Type: text.ReturnJump, Exp: nil},
			[]string{"return"},
		},
		{
			text.JumpStatement{Type: text.ReturnJump, Exp: text.String("Nice")},
			[]string{`ldc "Nice"`, "areturn"},
		},
		{
			text.JumpStatement{Type: text.ReturnJump, Exp: text.Num(1)},
			[]string{"iconst_1", "ireturn"},
		},
		{
			text.JumpStatement{Type: text.ReturnJump, Exp: text.Char('\u0001')},
			[]string{"iconst_1", "ireturn"},
		},
		{
			text.JumpStatement{Type: text.ReturnJump, Exp: text.Boolean(true)},
			[]string{"iconst_1", "ireturn"},
		},
	}

	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		d.jump.Accept(gen)
		assertHasSameCodes(t, gen, d.expect...)
	}
}

func TestKrakatauGen_IfStatement(t *testing.T) {
	getStatic := "getstatic Field java/lang/System out Ljava/io/PrintStream;"
	invokeSysout := "invokevirtual Method java/io/PrintStream println (I)V"
	data := []struct {
		ifstmt text.IfStatement
		expect []string
	}{
		{
			text.IfStatement{
				Condition: text.Boolean(true),
				Body:      &text.MethodCallStatement{Method: mockSysout(text.Num(12))},
				Else:      nil,
			},
			[]string{
				"iconst_1",
				"ifne L0", // if true
				"goto L1",
				"L0:\t",
				getStatic,
				"bipush 12",
				invokeSysout,
				"goto L1",
				"L1:\t",
			},
		},
		{
			text.IfStatement{
				Condition: text.Boolean(true),
				Body:      text.StatementList{},
				Else:      text.StatementList{},
			},
			[]string{
				"iconst_1",
				"ifne L0", // if true
				"goto L2",
				"L0:\t",
				// if body
				"goto L1",
				"L2:\t",
				// else body
				"goto L1",
				"L1:\t",
				// outer
			},
		},
		{
			text.IfStatement{
				Condition: text.Boolean(true),
				Body:      text.StatementList{},
				Else: &text.IfStatement{
					Condition: text.Boolean(true),
					Body:      text.StatementList{},
					Else:      nil,
				},
			},
			[]string{
				"iconst_1",
				"ifne L0", // if true
				"goto L2",
				"L0:\t",
				"goto L1", // if body
				"L2:\t",   // else if
				"iconst_1",
				"ifne L3", //condition
				"goto L1",
				"L3:\t",
				"goto L1",
				"L1:\t",
			},
		},
	}

	for _, d := range data {
		gen := NewEmptyKrakatauGen()
		d.ifstmt.Accept(gen)
		assertHasSameCodes(t, gen, d.expect...)
	}
}

func newNamedType(name string, isarray bool) text.NamedType {
	return text.NamedType{Name: name, IsArray: isarray}
}

func mockKrakatau(do func(gen *KrakatauGen)) {
	// content := ``
	// lexer := text.NewLexer(text.NewStringScanner(content))
	// parser := text.NewParser(&lexer)
	gen := NewEmptyKrakatauGen()
	gen.typeTable = NewEmptyKrakatauGen().typeTable
	do(gen)
}

func TestKrakatauGen_VariableDeclaration(t *testing.T) {
	data := []struct {
		symbol  Local
		varDecl text.VariableDeclaration
		expect  []string
	}{
		{
			Local{&FieldSymbol{mockInt, "count"}, 1},
			text.VariableDeclaration{
				newNamedType("int", false),
				"count",
				text.Num(1),
			},
			[]string{
				"iconst_1",
				"istore_1",
			},
		},
		{
			Local{&FieldSymbol{mockString, "name"}, 12},
			text.VariableDeclaration{
				newNamedType("String", false),
				"name",
				text.String("Hello"),
			},
			[]string{
				`ldc "Hello"`,
				"astore 12",
			},
		},
		{
			Local{&FieldSymbol{mockString, "name"}, 1},
			text.VariableDeclaration{
				newNamedType("String", false),
				"name",
				nil,
			},
			[]string{
				`aconst_null`,
				"astore_1",
			},
		},
		{
			Local{&FieldSymbol{mockBoolean, "name"}, 1},
			text.VariableDeclaration{
				newNamedType("boolean", false),
				"name",
				nil,
			},
			[]string{
				`iconst_0`,
				"istore_1",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			table := NewSymbolTable("mock", 0, nil)
			table.Insert(d.symbol.Member, d.symbol.address)
			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}
			d.varDecl.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatauGen_FieldAccess(t *testing.T) {
	data := []struct {
		symbol  Local
		varDecl text.FieldAccess
		expect  []string
	}{
		{
			Local{&FieldSymbol{mockInt, "count"}, 12},
			text.FieldAccess{Name: "count", Child: nil},
			[]string{
				"iload 12",
			},
		},
		{
			Local{&FieldSymbol{mockHuman, "human"}, 1},
			text.FieldAccess{Name: "human",
				Child: &text.FieldAccess{
					Name:  "age",
					Child: nil,
				}},
			[]string{
				"aload_1",
				"getfield Field Human age I",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			table := NewSymbolTable("mock", 0, nil)
			table.Insert(d.symbol.Member, d.symbol.address)
			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}

			gen.typeTable = NewTypeAnalyzer().table
			gen.typeTable["Human"] = mockHuman.dataType

			d.varDecl.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatauGen_AssignmentStatement(t *testing.T) {
	var eq text.Token
	eq.Type = text.Assignment

	data := []struct {
		symbol     Local
		assignment text.AssignmentStatement
		expect     []string
	}{
		{
			Local{&FieldSymbol{mockInt, "count"}, 1},
			text.AssignmentStatement{Operator: eq,
				Left:  &text.FieldAccess{Name: "count", Child: nil},
				Right: text.Num(1),
			},
			[]string{
				"iconst_1",
				"istore_1",
			},
		},
		{
			Local{&FieldSymbol{mockHuman, "human"}, 3},
			text.AssignmentStatement{Operator: eq,
				Left: &text.FieldAccess{
					Name: "human",
					Child: &text.FieldAccess{
						Name:  "age",
						Child: nil,
					},
				},
				Right: text.Num(1),
			},
			[]string{
				"aload_3",
				"iconst_1",
				"putfield Field Human age I",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			table := NewSymbolTable("mock", 0, nil)
			table.Insert(d.symbol.Member, d.symbol.address)
			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}

			gen.typeTable = NewTypeAnalyzer().table
			gen.typeTable["Human"] = mockHuman.dataType

			d.assignment.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatau_ObjecCreation(t *testing.T) {
	data := []struct {
		obj    text.ObjectCreation
		expect []string
	}{
		{
			text.ObjectCreation{
				text.MethodCall{
					Name:  "Human",
					Args:  []text.Expression{},
					Child: nil,
				},
			},
			[]string{
				"new Human",
				"dup",
				"invokespecial Method Human <init> ()V",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			gen.VisitObjectCreation(&d.obj)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatauGen_MethodCall(t *testing.T) {
	data := []struct {
		local  Local
		method text.NamedValue
		expect []string
	}{
		{
			Local{&FieldSymbol{mockHuman, "human"}, 3},
			&text.FieldAccess{Name: "human",
				Child: &text.MethodCall{
					Name:  "getAge",
					Args:  []text.Expression{},
					Child: nil,
				}},
			[]string{
				"aload_3",
				"invokevirtual Method Human getAge ()I",
			},
		},
		{
			Local{&FieldSymbol{mockHuman, "human"}, 3},
			&text.FieldAccess{Name: "human",
				Child: &text.MethodCall{
					Name: "getName",
					Args: []text.Expression{
						text.String("Hello"),
					},
					Child: nil,
				}},
			[]string{
				"aload_3",
				`ldc "Hello"`,
				"invokevirtual Method Human getName (Ljava/lang/String;)I",
			},
		},
	}

	human := mockHuman.dataType
	human.Methods[methodGetAge.Signature()] = NewMethodSymbol(methodGetAge.MethodSignature, *mockInt.dataType)
	human.Methods[methodGetNameWithParam.Signature()] = NewMethodSymbol(methodGetName.MethodSignature, *mockInt.dataType)

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			gen.typeTable = NewTypeAnalyzer().table
			gen.typeTable["Human"] = human

			table := NewSymbolTable("mock", 0, nil)
			table.Insert(d.local.Member, d.local.address)
			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}

			d.method.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatauGen_This(t *testing.T) {
	data := []struct {
		namedValue text.NamedValue
		expect     []string
	}{
		{
			&text.This{Child: &text.FieldAccess{Name: "age", Child: nil}},
			[]string{
				"aload_0",
				"getfield Field Human age I",
			},
		},
		{
			&text.This{Child: &text.MethodCall{
				Name: "getName",
				Args: []text.Expression{
					text.String("Hello"),
				},
				Child: nil,
			}},
			[]string{
				"aload_0",
				`ldc "Hello"`,
				"invokevirtual Method Human getName (Ljava/lang/String;)I",
			},
		},
	}

	human := mockHuman.dataType
	human.Properties["age"] = &PropertySymbol{
		propAge.AccessModifier,
		FieldSymbol{
			mockInt,
			"age",
		},
	}

	human.Methods[methodGetAge.Signature()] = NewMethodSymbol(methodGetAge.MethodSignature, *mockInt.dataType)
	human.Methods[methodGetNameWithParam.Signature()] = NewMethodSymbol(methodGetName.MethodSignature, *mockInt.dataType)

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			gen.typeTable = NewTypeAnalyzer().table
			gen.typeTable["Human"] = human

			table := NewSymbolTable("mock", 0, nil)
			table.Insert(&PropertySymbol{
				text.Public,
				FieldSymbol{mockHuman, "this"},
			}, 0)

			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}

			d.namedValue.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}

func TestKrakatauGen_WhileStatement(t *testing.T) {
	var add, lt, assign text.Token
	add.Type = text.Addition
	lt.Type = text.LessThan
	assign.Type = text.Assignment

	condition := text.NewBinOp(lt, &text.FieldAccess{Name: "a", Child: nil}, text.Num(10))
	increment := text.NewBinOp(add, &text.FieldAccess{Name: "a", Child: nil}, text.Num(1))

	data := []struct {
		local  Local
		while  text.WhileStatement
		expect []string
	}{
		{
			Local{&FieldSymbol{mockInt, "a"}, 1},
			text.WhileStatement{
				Condition: text.Boolean(true),
				Body:      text.StatementList{},
			},
			[]string{
				"L0:\t",
				"iconst_1",
				"ifne L1", // true
				"goto L2",
				"L1:\t",
				"goto L0",
				"L2:\t",
			},
		},
		{
			Local{&FieldSymbol{mockInt, "a"}, 1},
			text.WhileStatement{
				Condition: &condition,
				Body: &text.AssignmentStatement{
					Operator: assign,
					Left: &text.FieldAccess{
						Name:  "a",
						Child: nil,
					},
					Right: &increment,
				},
			},
			[]string{
				"L0:\t",
				"iload_1",
				"bipush 10",
				"if_icmplt L1",
				"iconst_0",
				"goto L2",
				"L1:\ticonst_1",
				"L2:\t",
				"ifne L3", // true
				"goto L4",
				"L3:\t",
				"iload_1",
				"iconst_1",
				"iadd",
				"istore_1",
				"goto L0",
				"L4:\t",
			},
		},
	}

	for _, d := range data {
		mockKrakatau(func(gen *KrakatauGen) {
			table := NewSymbolTable("mock", 0, nil)
			table.Insert(d.local.Member, d.local.address)
			gen.scopeIndex = 0
			gen.symbolTable = []*SymbolTable{
				&table,
			}

			d.while.Accept(gen)
			assertHasSameCodes(t, gen, d.expect...)
		})
	}
}
