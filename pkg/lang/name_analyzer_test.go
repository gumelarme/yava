package lang

import (
	"fmt"
	"testing"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	mockTypeTable TypeTable
)

func init() {
	mockTypeTable = NewTypeAnalyzer().table
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
		expectName := "class-Human_1"
		if nameAnal.scope.level != 1 || nameAnal.scope.name != expectName {
			t.Errorf("VisitClass should result in new scope, but got: %s:%d",
				nameAnal.scope.name,
				nameAnal.scope.level,
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
		if nameAnal.scope.level != 2 || nameAnal.scope.name != expectName {
			t.Errorf("VisitMethodSignature should result in new scope but got: %s:%d",
				nameAnal.scope.name,
				nameAnal.scope.level,
			)
		}

		param := nameAnal.scope.Lookup("what").(FieldSymbol)
		if len(nameAnal.scope.table) == 0 {
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
		if len(nameAnal.scope.table) == 0 {
			t.Error("Variable 'int realAge' should exist after visiting VariableDeclaration.")
		}

		variable := nameAnal.scope.Lookup("realAge").(FieldSymbol)
		if variable.Name() != "realAge" ||
			variable.dataType.name != "int" ||
			variable.isArray != true {
			t.Errorf("Variable inserted into scope is incorrect.")
		}
	})
}

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
		if nameAnal.scope.name != expect {
			t.Errorf("Expecting scope name of `%s` but got `%s`", expect, nameAnal.scope.name)
		}
	})

	forStmt.Init = nil
	withTypeAnal(&human, func(nameAnal *NameAnalyzer) {
		expect := "method-getAge()_2"
		if nameAnal.scope.name != expect {
			t.Errorf("Expecting scope name of `%s` but got `%s`", expect, nameAnal.scope.name)
		}
	})

}
