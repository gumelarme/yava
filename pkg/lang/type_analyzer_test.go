package lang

import (
	"fmt"
	"testing"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	interfaceCallable *text.Interface
	classHuman        *text.Class
	classPerson       *text.Class
	//properties
	propAge  *text.PropertyDeclaration
	propName *text.PropertyDeclaration
	// methods
	methodGetAge           *text.MethodDeclaration
	methodGetName          *text.MethodDeclaration
	methodGetNameWithParam *text.MethodDeclaration
)

func init() {
	propAge = &text.PropertyDeclaration{
		AccessModifier: text.Public,
		VariableDeclaration: text.VariableDeclaration{
			Type:  text.NamedType{Name: "int", IsArray: false},
			Name:  "age",
			Value: nil,
		},
	}

	propName = &text.PropertyDeclaration{
		AccessModifier: text.Public,
		VariableDeclaration: text.VariableDeclaration{
			Type:  text.NamedType{Name: "String", IsArray: false},
			Name:  "name",
			Value: text.String("Hello"),
		},
	}

	methodGetAge = text.NewMethodDeclaration(
		text.Public,
		text.NamedType{Name: "int", IsArray: false},
		"getAge",
		[]text.Parameter{},
		text.StatementList{},
	)

	methodGetName = text.NewMethodDeclaration(
		text.Public,
		text.NamedType{Name: "int", IsArray: false},
		"getName",
		[]text.Parameter{},
		text.StatementList{},
	)

	methodGetNameWithParam = text.NewMethodDeclaration(
		text.Public,
		text.NamedType{Name: "int", IsArray: false},
		"getName",
		[]text.Parameter{
			{Type: text.NamedType{Name: "String", IsArray: false}, Name: "MrOrMs"},
		},
		text.StatementList{},
	)

	interfaceCallable = text.NewInterface("Callable")
	classHuman = text.NewEmptyClass("Human", "", "")
	classPerson = text.NewEmptyClass("Person", "", "")

}

func TestNewTypeAnalyzer(t *testing.T) {
	engine := NewTypeAnalyzer()
	expect := []string{"int", "boolean", "char", "String"}
	for _, ex := range expect {
		if _, exist := engine.table[ex]; !exist {
			t.Errorf("Must have type %s initialized.", ex)
		}
	}
}

func TestTypeAnalyzer_Class(t *testing.T) {
	testNoError := func(doThings func(engine *TypeAnalyzer)) {
		engine := NewTypeAnalyzer()
		doThings(engine)
		if len(engine.error) != 0 {
			t.Error("Should not have any error, but got:\n", engine.error)
		}
	}

	testNoError(func(engine *TypeAnalyzer) {
		engine.VisitClass(classHuman)
	})

	testNoError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Extend = "Human"
		engine.VisitClass(classHuman)
		engine.VisitClass(&person)
	})

	testNoError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Implement = "Callable"
		engine.VisitInterface(interfaceCallable)
		engine.VisitClass(&person)
	})
}

func TestTypeAnalyzer_Class_errors(t *testing.T) {
	testError := func(doThings func(engine *TypeAnalyzer), msg string) {
		engine := NewTypeAnalyzer()
		doThings(engine)
		if engine.error[0] != msg {
			t.Errorf("Should have an error:\n`%s`\ninstead of:\n`%s`", msg, engine.error[0])
		}
	}

	testError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Extend = "Human"
		engine.VisitClass(&person)
	}, fmt.Sprintf(msgTypeNotExist, "Human"))

	testError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Extend = "Callable"
		engine.VisitInterface(interfaceCallable)
		engine.VisitClass(&person)

	}, msgExtendShouldBeOnClass)

	testError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Implement = "Human"
		engine.VisitClass(classHuman)
		engine.VisitClass(&person)

	}, msgImplementShouldBeOnInterface)
}

func TestTypeAnalyzer_Class_alreadyDeclared(t *testing.T) {
	engine := NewTypeAnalyzer()
	engine.VisitClass(text.NewEmptyClass("String", "", ""))
	if engine.error[0] != fmt.Sprintf(msgTypeAlreadyDeclared, "String") {
		t.Errorf("Should error on duplicate type name, %s", engine.error[0])
	}
}

func TestTypeAnalyzer_Interface_alreadyDeclared(t *testing.T) {
	engine := NewTypeAnalyzer()
	engine.VisitInterface(interfaceCallable)
	engine.VisitInterface(interfaceCallable)
	if engine.error[0] != fmt.Sprintf(msgTypeAlreadyDeclared, interfaceCallable.Name) {
		t.Errorf("Should error on duplicate type name, %s", engine.error[0])
	}
}

func TestTypeAnalyzer_AfterClass(t *testing.T) {
	callable := *interfaceCallable
	callable.AddMethod(&methodGetAge.MethodSignature)
	human := *classHuman

	engine := NewTypeAnalyzer()
	callable.Accept(engine)
	human.Accept(engine)

	if len(engine.error) != 0 {
		t.Errorf("Expected no error, but got `%s`", engine.error[0])
	}

	//reset, and add interface to human
	human.Implement = "Callable"
	engine = NewTypeAnalyzer()
	callable.Accept(engine)
	human.Accept(engine)

	msg := fmt.Sprintf(msgMustImplementMethod, methodGetAge.Signature())
	if len(engine.error) == 0 {
		t.Errorf("Should have error of: %s, but got nothing", msg)
	}
}

func TestTypeAnalyzer_PropertyDeclaration(t *testing.T) {
	human := *classHuman
	human.Properties = append(human.Properties, propAge)

	humanType := NewType("Human", Class)
	humanType.Properties["age"] = &PropertySymbol{
		text.Public,
		FieldSymbol{
			DataType: DataType{
				NewType("int", Primitive),
				propAge.Type.IsArray,
			},
			name: "age",
		},
	}

	expect := map[string]*TypeSymbol{
		"Human": humanType,
	}

	visitor := NewTypeAnalyzer()
	human.Accept(visitor)
	table := visitor.GetTypeTable()
	for typeName, typeSymbol := range expect {
		symbol, exist := table[typeName]
		if !exist {
			t.Errorf("Type %s should have exist.", typeName)
		}

		for propName, property := range typeSymbol.Properties {
			_, exist := symbol.Properties[propName]
			if !exist {
				t.Errorf("Property %s should have exist.", property)
			}
		}
	}

}

func TestTypeAnalyzer_PropertyDeclaration_error(t *testing.T) {
	human := *classHuman
	human.Properties = append(human.Properties, propAge)
	human.Properties = append(human.Properties, propAge)
	visitor := NewTypeAnalyzer()
	human.Accept(visitor)
	msg := fmt.Sprintf(msgPropertyAlreadyDeclared, propAge.Name, human.Name)
	if len(visitor.error) < 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of `%s `", msg)
	}

	// Type Not exist error
	newPropName := *propName
	newPropName.Type.Name = "Hello"

	human = *classHuman
	human.Properties = append(human.Properties, &newPropName)
	msg = fmt.Sprintf(msgTypeNotExist, "Hello")
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)
	if len(visitor.error) < 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of `%s `", msg)
	}

}

func TestTypeAnalyzer_MethodDeclaration(t *testing.T) {
	newMethodAge := *methodGetAge
	newMethodAge.ReturnType.Name = "void"
	newMethodAge.ParameterList = []text.Parameter{
		{
			Type: text.NamedType{Name: "int", IsArray: false},
			Name: "a",
		},
	}
	human := *classHuman
	human.Methods = append(human.Methods, &newMethodAge)
	visitor := NewTypeAnalyzer()
	human.Accept(visitor)

	if len(visitor.error) > 0 {
		t.Errorf("Should be error free but got:\n %s", visitor.error[0])
	}

	newMethodAge = *methodGetAge
	newMethodAge.ReturnType.Name = "void"
	newMethodAge.ParameterList = []text.Parameter{
		{
			Type: text.NamedType{Name: "int", IsArray: false},
			Name: "a",
		},
	}

	//overloading
	human = *classHuman
	human.Methods = append(human.Methods, methodGetAge, &newMethodAge)
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)

	if len(visitor.error) > 0 {
		t.Errorf("Should be error free but got:\n %s", visitor.error[0])
	}

}
func TestTypeAnalyzer_MethodDeclaration_error(t *testing.T) {
	//Already declared as property
	newMethodAge := *methodGetAge
	newMethodAge.Name = "age"

	human := *classHuman
	human.Properties = append(human.Properties, propAge)
	human.Methods = append(human.Methods, &newMethodAge)

	visitor := NewTypeAnalyzer()
	human.Accept(visitor)
	msg := fmt.Sprintf(msgMethodAlreadyDeclaredAsProp, newMethodAge.Name)
	if len(visitor.error) == 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of :\n\t%s \nbut got: \n%s", msg, visitor.error[0])
	}

	// duplicate method
	human = *classHuman
	human.Methods = append(human.Methods, methodGetAge, methodGetAge)
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)

	msg = fmt.Sprintf(msgMethodIsAlreadyDeclared, methodGetAge.Signature())
	if len(visitor.error) == 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of :\n\t%s \n", msg)
	}

	// unkown type
	newMethodAge = *methodGetAge
	newMethodAge.ReturnType.Name = "Nice"
	human = *classHuman
	human.Methods = append(human.Methods, &newMethodAge)
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)
	msg = fmt.Sprintf(msgTypeNotExist, newMethodAge.ReturnType.Name)
	if len(visitor.error) == 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of :\n\t%s", msg)
	}

	// parameter unknown type
	newMethodAge = *methodGetAge
	newMethodAge.ParameterList = []text.Parameter{
		{
			Type: text.NamedType{Name: "Nice", IsArray: false},
			Name: "Hello",
		},
	}
	human = *classHuman
	human.Methods = append(human.Methods, &newMethodAge)
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)
	msg = fmt.Sprintf(msgTypeNotExist, newMethodAge.ParameterList[0].Name)
	if len(visitor.error) == 0 || visitor.error[0] != msg {
		t.Errorf("Should have got error of :\n\t%s", msg)
	}
}
