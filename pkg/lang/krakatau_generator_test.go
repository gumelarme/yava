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
		{255, "bipush 255"},
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
	if codeCount := len(gen.codes); codeCount != count {
		t.Errorf("Should at least has %d line of codes, but got %d.", count, codeCount)
		return false
	}
	return true
}

func assertHasSameCodes(t *testing.T, gen *KrakatauGen, expect []string) {
	if !assertHasNCode(t, gen, len(expect)) {
		return
	}

	for i, code := range expect {
		if code != gen.codes[i] {
			t.Errorf("Expecting: \n%#v but got %#v", code, gen.codes[i])
		}
	}
}

func TestKrakatauGen_Program(t *testing.T) {
	var program text.Program
	generator := NewKrakatauGenerator()
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
		gen := NewKrakatauGenerator()
		gen.VisitClass(d.class)
		assertHasSameCodes(t, gen, d.expect)
	}
}

func TestKrakatauGen_AfterClass(t *testing.T) {
	var class *text.Class
	gen := NewKrakatauGenerator()
	gen.VisitAfterClass(class)
	assertHasSameCodes(t, gen, []string{
		".end class",
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
	}{
		{
			text.NewBinOp(add, text.Num(12000), text.Num(3)),
			[]string{
				"ldc 12000",
				"iconst_3",
				"iadd",
			},
		},
		{
			text.NewBinOp(mod, text.Num(12), text.Num(3)),
			[]string{
				"bipush 12",
				"iconst_3",
				"imod",
			},
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
		},
	}

	for _, d := range data {
		gen := NewKrakatauGenerator()
		d.binaryOp.Accept(gen)
		assertHasSameCodes(t, gen, d.expect)
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
		gen := NewKrakatauGenerator()
		gen.VisitMethodSignature(&d.signature)
		assertHasSameCodes(t, gen, []string{d.expect})
	}
}

func TestKrakatauGen_AfterMethodDeclaration(t *testing.T) {
	gen := NewKrakatauGenerator()
	gen.VisitAfterMethodDeclaration(nil)
	assertHasSameCodes(t, gen, []string{
		".end code",
		".end method",
	})
}
