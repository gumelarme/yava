package lang

import (
	"fmt"
	"testing"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	mockNull    DataType
	mockInt     DataType
	mockBoolean DataType
	mockChar    DataType
	mockString  DataType
	mockHuman   DataType
	mockPerson  DataType
)

func init() {
	mockNull = DataType{PrimitiveNull, false}
	mockInt = DataType{PrimitiveInt, false}
	mockBoolean = DataType{PrimitiveBoolean, false}
	mockChar = DataType{PrimitiveChar, false}
	mockString = DataType{PrimitiveString, false}

	humanClass := NewType("Human", Class)
	humanClass.Properties["age"] = &PropertySymbol{
		text.Public,
		FieldSymbol{
			DataType{PrimitiveInt, false},
			"age",
		},
	}
	mockHuman = DataType{humanClass, false}

	personClass := NewType("Person", Class)
	mockPerson = DataType{personClass, false}
}

func getMockTypeTable(templates ...text.Template) TypeTable {
	program := make(text.Program, len(templates))
	for i, t := range templates {
		program[i] = t
	}

	typeAnalyzer := NewTypeAnalyzer()
	program.Accept(typeAnalyzer)
	return typeAnalyzer.table
}

func TestTypeStack(t *testing.T) {
	var stack TypeStack
	stack.Push(mockInt)

	if len(stack) != 1 {
		t.Errorf("There should be 1 item in stack, but got %d instead", len(stack))
	}

	typeof, err := stack.Pop()
	if typeof == (DataType{}) || err != nil {
		t.Errorf("There should be 1 item in stack, but got nil")
	}

	typeof, err = stack.Pop()
	if typeof != (DataType{}) || err == nil {
		t.Errorf("There should be 0 item in stack, but got %s", typeof)
	}

	stack.Push(mockInt)
	stack.Overwrite(mockHuman)
	typeof, _ = stack.Pop()

	if typeof != mockHuman {
		t.Errorf("Item is not overwritten, got %s instead of %s", typeof, mockHuman)
	}
}

func withTypeAnal(node text.INode, doThings func(nameAnal *NameAnalyzer)) {
	typeAnal := NewTypeAnalyzer()
	node.Accept(typeAnal)
	nameAnal := NewNameAnalyzer(typeAnal.GetTypeTable())
	node.Accept(nameAnal)
	doThings(nameAnal)
}

func TestNameAnalyzer_Class(t *testing.T) {
	withTypeAnal(classHuman, func(nameAnal *NameAnalyzer) {
		expectName := "<program>"
		if nameAnal.scope.level != 0 ||
			nameAnal.scope.name != expectName ||
			nameAnal.counter != 1 {
			t.Errorf("VisitClass should result in new scope, but got: %s:%d, counted in %d instead of %d",
				nameAnal.scope.name,
				nameAnal.scope.level,
				nameAnal.counter,
				1,
			)
		}
	})
}

func TestNameAnalyzer_MethodSignature(t *testing.T) {
	human := *classHuman
	newMethodGetAge := *methodGetAge
	newMethodGetAge.ParameterList = []text.Parameter{
		{
			Type: text.NamedType{Name: "int", IsArray: false},
			Name: "what",
		},
	}
	human.Methods = append(human.Methods, &newMethodGetAge)

	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		expectName := "method-getAge(int)_2"
		scope := nameAnal.Tables[1]
		if scope.level != 2 || scope.name != expectName {
			t.Errorf("VisitMethodSignature should result in new scope but got: %s:%d",
				nameAnal.scope.name,
				nameAnal.scope.level,
			)
		}

		sym, _ := scope.Lookup("what", true)
		param := sym.(*FieldSymbol)
		if len(scope.table) == 0 {
			t.Errorf("MethodSignature should introduce new field from parameter inside a new scope")
		}

		if param.dataType.name != "int" ||
			param.isArray != false ||
			param.name != "what" {
			t.Errorf("Parameter is not equal as the method signature. ")
		}
	})
}

func TestNameAnalyzer_MethodSignature_error(t *testing.T) {
	human := *classHuman
	newMethodGetAge := *methodGetAge
	newMethodGetAge.ParameterList = []text.Parameter{
		{
			Type: text.NamedType{Name: "int", IsArray: false},
			Name: "what",
		},
		{
			Type: text.NamedType{Name: "int", IsArray: false},
			Name: "what",
		},
	}
	human.Methods = append(human.Methods, &newMethodGetAge)
	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		errors := nameAnal.Errors()
		if len(errors) == 0 {
			t.Error("Should be error when two parameter has the same name.")
		}

		expect := fmt.Sprintf(msgParameterAlreadyDeclared, "what")
		if err := errors[0].Error(); expect != err {
			t.Errorf("Expecting: %s but got:\n%s", expect, err)
		}
	})
}

func TestNameAnalyzer_VariableDeclaration(t *testing.T) {
	human := *classHuman
	newMethodGetAge := *methodGetAge
	human.Methods = append(human.Methods, &newMethodGetAge)
	newMethodGetAge.Body = append(newMethodGetAge.Body, &text.VariableDeclaration{
		Type:  text.NamedType{Name: "int", IsArray: true},
		Name:  "realAge",
		Value: nil,
	})

	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		scope := nameAnal.Tables[1]
		if len(scope.table) < 2 {
			t.Error("Variable 'int realAge' should exist after visiting VariableDeclaration.")
		}

		sym, _ := scope.Lookup("realAge", true)
		if sym == nil {
			t.Error("Variable 'int realAge' is not exist")
		}

		variable := sym.(*FieldSymbol)
		if variable.Name() != "realAge" ||
			variable.dataType.name != "int" ||
			variable.isArray != true {
			t.Errorf("Variable inserted into scope is incorrect.")
		}
	})
}

// func withNameAnalyzerFromText(content string, do func(*NameAnalyzer, text.Program)) {
// 	lex := text.NewLexer(text.NewStringScanner(content))
// 	parser := text.NewParser(&lex)
// 	ast := parser.Compile()

// 	typeAnalyzer := NewTypeAnalyzer()
// 	ast.Accept(typeAnalyzer)

// 	nameAnalyzer := NewNameAnalyzer(typeAnalyzer.table)
// 	do(nameAnalyzer, ast)
// }

// func TestNameAnalyzer_VariableDeclaration_badType(t *testing.T) {
// 	template := `
// class Person {
// 	public int a;
// 	public int getA(){
// 		%s
// 	}
// }
// `
// 	content := fmt.Sprintf(template, "int a = 3 + 3;")
// 	withNameAnalyzerFromText(content, func(nameAnalyzer *NameAnalyzer, program text.Program) {
// 		class := program[0].(*text.Class)
// 		nameAnalyzer.VisitClass(class)
// 		nameAnalyzer.VisitMethodSignature(&class.MainMethod.MethodSignature)
// 		nameAnalyzer.VisitAfterClass(class)
// 	})
// }

func TestNameAnalyzer_VariableDeclaration_error(t *testing.T) {
	checkHasError := func(class *text.Class, errorMsg string) {
		withTypeAnal(class, func(nameAnal *NameAnalyzer) {
			errors := nameAnal.Errors()
			if len(errors) == 0 {
				t.Errorf("Got no error, expecting: `%s`", errorMsg)
			}

			if errors[0].Error() != errorMsg {
				t.Errorf("Expecting error message of:\n`%s`\nbut got:\n%s", errorMsg, errors[0].Error())
			}
		})
	}

	human := *classHuman
	newMethodGetAge := *methodGetAge
	human.Methods = append(human.Methods, &newMethodGetAge)
	variable := &text.VariableDeclaration{
		Type:  text.NamedType{Name: "int", IsArray: true},
		Name:  "realAge",
		Value: nil,
	}
	newMethodGetAge.Body = append(newMethodGetAge.Body, variable, variable)
	checkHasError(&human, fmt.Sprintf(msgVariableAlreadyDeclared, "realAge"))

	// type not exist
	variable.Type.Name = "SomethingDidNotExist"
	newMethodGetAge.Body = []text.Statement{variable}
	checkHasError(&human, fmt.Sprintf(msgTypeNotExist, "SomethingDidNotExist"))
}

func TestNameAnalyzer_ForStatement(t *testing.T) {
	human := *classHuman
	newMethodGetAge := *methodGetAge
	human.Methods = append(human.Methods, &newMethodGetAge)
	forStmt := text.ForStatement{
		Init: &text.VariableDeclaration{
			Type:  text.NamedType{Name: "int", IsArray: false},
			Name:  "something",
			Value: nil,
		},
		Condition: nil,
		Update:    nil,
		Body: &text.JumpStatement{
			Type: text.ReturnJump,
			Exp:  nil,
		},
	}
	newMethodGetAge.Body = text.StatementList{&forStmt}

	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		expect := "for-scope-2_3"
		scope := nameAnal.Tables[2]
		if scope.name != expect {
			t.Errorf("Expecting scope name of `%s` but got `%s`", expect, scope.name)
		}
	})

	forStmt.Init = nil
	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		expect := "method-getAge()_2"
		scope := nameAnal.Tables[1]
		if scope.name != expect {
			t.Errorf("Expecting scope name of `%s` but got `%s`", expect, scope.name)
		}
	})
}

func TestNameAnalyzer_VisitBinOp(t *testing.T) {
	data := []struct {
		tokens      []text.TokenType
		result      DataType
		operandType []DataType
	}{
		{
			[]text.TokenType{text.Addition, text.Subtraction, text.Multiplication, text.Division, text.Modulus},
			mockInt,
			[]DataType{mockInt, mockInt},
		},
		{
			[]text.TokenType{text.GreaterThan, text.GreaterThanEqual, text.LessThan, text.LessThanEqual},
			mockBoolean,
			[]DataType{mockInt, mockInt},
		},
		{
			[]text.TokenType{text.Or, text.And},
			mockBoolean,
			[]DataType{mockBoolean, mockBoolean},
		},
		{
			[]text.TokenType{text.Equal, text.NotEqual},
			mockBoolean,
			[]DataType{mockBoolean, mockBoolean},
		},
	}

	for _, d := range data {
		for _, token := range d.tokens {
			nameAnalyzer := NewNameAnalyzer(nil)
			nameAnalyzer.stack = append(nameAnalyzer.stack, d.operandType...)

			operator := text.Token{}
			operator.Type = token

			binop := text.NewBinOp(operator, nil, nil)
			nameAnalyzer.VisitAfterBinOp(&binop)
		}
	}
}

func TestNameAnalyzer_mustBeTypeOf(t *testing.T) {
	left, right := mockString, mockString
	nameAnalyzer := NewNameAnalyzer(nil)
	nameAnalyzer.mustBeTypeof(left, right, "boolean", "int")

	err := nameAnalyzer.Errors()
	expect := fmt.Sprintf(msgExpectingTypeof, "boolean, int", left)

	msgError := err[0].Error()
	if msgError != expect {
		t.Errorf("Expecting error of: \n%s \nbut got: \n%s", expect, msgError)
	}
}

func TestIsNullOk(t *testing.T) {
	intArray := mockInt
	intArray.isArray = true

	data := []struct {
		data   DataType
		expect bool
	}{
		{mockInt, false},
		{mockBoolean, false},
		{mockChar, false},
		{mockString, true},
		{mockHuman, true},
		{intArray, true},
	}

	for _, d := range data {
		if IsNullOk(d.data) != d.expect {
			t.Errorf("Data type %s expected to return %v", d.data, d.expect)
		}
	}
}

func TestNameAnalyzer_expecLastStackTypeOf(t *testing.T) {
	table := NewTypeAnalyzer().table

	nameAnalyzer := NewNameAnalyzer(table)
	nameAnalyzer.stack = append(nameAnalyzer.stack, mockString)
	if nameAnalyzer.expectLastStackTypeOf("int", false) != false {
		t.Error("Expected to return false.")
	}

	intArray := mockInt
	intArray.isArray = true
	nameAnalyzer.stack = append(nameAnalyzer.stack, intArray)
	if nameAnalyzer.expectLastStackTypeOf("int", true) != true {
		t.Errorf("Expected to return true.")
	}

}

func TestNameAnalyzer_VisitBinOp_error(t *testing.T) {
	operator := text.Token{}
	operator.Type = text.Addition

	left, right := mockInt, mockBoolean
	binOp := text.NewBinOp(operator, nil, nil)

	nameAnalyzer := NewNameAnalyzer(nil)
	nameAnalyzer.stack = append(nameAnalyzer.stack, left, right)
	nameAnalyzer.VisitAfterBinOp(&binOp)

	if size := len(nameAnalyzer.stack); size != 0 {
		t.Errorf("Stack should be empty on failed operation, but got %d item.", size)
	}

	err := nameAnalyzer.Errors()
	if len(err) == 0 {
		t.Errorf("Failed operator should add errors.")
	}

	expect := fmt.Sprintf(msgExpectingTypeof, left, right)
	if err[0].Error() != expect {
		t.Errorf("Failed operator should add error of: \n%s \nbut got: \n%s", err[0].Error(), expect)
	}
}

func TestNameAnalyzer_VisitConstant(t *testing.T) {
	// primitives types
	data := []struct {
		expression text.Expression
		expected   DataType
	}{

		{text.Null{}, mockNull},
		{text.Num(1), mockInt},
		{text.Boolean(true), mockBoolean},
		{text.Char('a'), mockChar},
		{text.String("Hello"), mockString},
	}

	for _, d := range data {
		table := NewTypeAnalyzer().table
		nameAnalyzer := NewNameAnalyzer(table)
		d.expression.Accept(nameAnalyzer)
		typeof, err := nameAnalyzer.stack.Pop()
		if err != nil {
			t.Error("Constant should put type on stack")
		}

		if typeof != d.expected {
			t.Errorf("Expected type of %s but got %s.", d.expected, typeof)
		}
	}

	// this
	human := *classHuman
	typeAnalyzer := NewTypeAnalyzer()
	human.Accept(typeAnalyzer)

	table := typeAnalyzer.table // mock table provide this

	expected := DataType{table["Human"], false}
	nameAnalyzer := NewNameAnalyzer(table)
	nameAnalyzer.newScope("func")
	nameAnalyzer.scope.Insert(&FieldSymbol{expected, "this"}, 0) //FIXME: change to another address

	this := &text.This{Child: &text.FieldAccess{Name: "Hello", Child: nil}}
	this.Accept(nameAnalyzer)

	typeof, err := nameAnalyzer.stack.Pop()
	if err != nil {
		t.Error("This should be put in the stack but got error.")
	}

	if typeof != expected {
		t.Errorf("Should be type of %s, but got %s.", expected, typeof)
	}

	thisSymbol := &FieldSymbol{expected, "this"}
	if nameAnalyzer.curField.String() != thisSymbol.String() {
		t.Errorf("name.curField should contain a field symbol but got %s instead of %s", nameAnalyzer.curField, thisSymbol)
	}
}
