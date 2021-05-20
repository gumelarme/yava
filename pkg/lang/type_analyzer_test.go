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
		errors := engine.Errors()
		if len(errors) != 0 {
			t.Error("Should not have any error, but got:\n", errors[0])
		}
	}

	testNoError(func(engine *TypeAnalyzer) {
		engine.VisitClass(classHuman)
		table := engine.GetTypeTable()
		if _, exist := table["Human"]; !exist {
			t.Errorf("Expecting 'Human' class to be present in the table")
		}
	})

	testNoError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Extend = "Human"
		engine.VisitClass(classHuman)
		engine.VisitClass(&person)

		humanExist := engine.typeExist("Human")
		if !humanExist {
			t.Errorf("Expecting 'Human' class to be present in the table")
		}

		personExist := engine.typeExist("Person")
		if !personExist {
			t.Errorf("Expecting 'Person' class to be present in the table\n%#v", engine.table)
		}

		personType := engine.table["Person"]
		humanType := engine.table["Human"]
		if personType.extends != humanType {
			t.Errorf("'Person' should have a pointer to 'Human' Type symbol")
		}
	})

	testNoError(func(engine *TypeAnalyzer) {
		person := *classPerson
		person.Implement = "Callable"
		engine.VisitInterface(interfaceCallable)
		engine.VisitClass(&person)

		callableExist := engine.typeExist("Callable")
		if !callableExist {
			t.Errorf("Expecting 'Callable' class to be present in the table")
		}

		personExist := engine.typeExist("Person")
		if !personExist {
			t.Errorf("Expecting 'Person' class to be present in the table\n%#v", engine.table)
		}

		personType := engine.table["Person"]
		callableType := engine.table["Callable"]
		if personType.implements != callableType {
			t.Errorf("'Person' should have a pointer to 'Callable' interface")
		}

	})
}

func TestTypeAnalyzer_Class_errors(t *testing.T) {
	testError := func(doThings func(engine *TypeAnalyzer), msg string) {
		engine := NewTypeAnalyzer()
		doThings(engine)
		errors := engine.Errors()
		if len(errors) > 0 && errors[0].Error() != msg {
			t.Errorf("Should have an error:\n`%s`\ninstead of:\n`%s`", msg, errors[0])
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
	errors := engine.Errors()
	if len(errors) > 0 && errors[0].Error() != fmt.Sprintf(msgTypeAlreadyDeclared, "String") {
		t.Errorf("Should error on duplicate type name, %s", errors[0])
	}
}

func TestTypeAnalyzer_Interface_alreadyDeclared(t *testing.T) {
	engine := NewTypeAnalyzer()
	engine.VisitInterface(interfaceCallable)
	engine.VisitInterface(interfaceCallable)
	errors := engine.Errors()
	if len(errors) > 0 && errors[0].Error() != fmt.Sprintf(msgTypeAlreadyDeclared, interfaceCallable.Name) {
		t.Errorf("Should error on duplicate type name, %s", errors[0])
	}
}

func TestTypeAnalyzer_AfterClass(t *testing.T) {
	callable := *interfaceCallable
	callable.AddMethod(&methodGetAge.MethodSignature)
	human := *classHuman

	engine := NewTypeAnalyzer()
	callable.Accept(engine)
	human.Accept(engine)

	errors := engine.Errors()
	if len(errors) != 0 {
		t.Errorf("Expected no error, but got `%s`", errors[0])
	}

	//reset, and add interface to human
	human.Implement = "Callable"
	engine = NewTypeAnalyzer()
	callable.Accept(engine)
	human.Accept(engine)

	msg := fmt.Sprintf(msgMustImplementMethod, methodGetAge.Signature())
	if len(engine.Errors()) == 0 {
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
	errors := visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
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

	errors = visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
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
	errors := visitor.Errors()
	if len(errors) > 0 {
		t.Errorf("Should be error free but got:\n %s", errors[0])
	}

	humanType := visitor.table["Human"]
	method := humanType.Methods[newMethodAge.Signature()]
	if method == nil {
		t.Errorf("Expecting %s method to be present in %s class",
			newMethodAge.Signature(),
			humanType.name,
		)
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
	errors = visitor.Errors()
	if len(errors) > 0 {
		t.Errorf("Should be error free but got:\n %s", errors[0])
	}

	humanType = visitor.table["Human"]
	for _, sign := range []string{newMethodAge.Signature(), methodGetAge.Signature()} {
		method := humanType.Methods[sign]
		if method == nil {
			t.Errorf("Expecting %s method to be present in %s class",
				newMethodAge.Signature(),
				humanType.name,
			)
		}
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
	errors := visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
		t.Errorf("Should have got error of :\n\t%s \nbut got: \n%s", msg, errors[0])
	}

	// duplicate method
	human = *classHuman
	human.Methods = append(human.Methods, methodGetAge, methodGetAge)
	visitor = NewTypeAnalyzer()
	human.Accept(visitor)

	msg = fmt.Sprintf(msgMethodIsAlreadyDeclared, methodGetAge.Signature())

	errors = visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
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

	errors = visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
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

	errors = visitor.Errors()
	if len(errors) == 0 || errors[0].Error() != msg {
		t.Errorf("Should have got error of :\n\t%s", msg)
	}
}
