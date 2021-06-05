package lang

import (
	"fmt"
	"strings"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	msgMethodAlreadyDeclaredAsProp  = "%s is already declared as a property."
	msgTypeNotExist                 = "Type %s is not exist."
	msgTypeAlreadyDeclared          = "Type %s is already declared."
	msgExtendShouldBeOnClass        = "Extend should be on class."
	msgImplementShouldBeOnInterface = "Implement should be on interface."
	msgMustImplementMethod          = "Must implement %s method."
	msgPropertyAlreadyDeclared      = "Property %#v is already exist in class %s."
	msgMethodIsAlreadyDeclared      = "Method %s is already exist."
)

type TypeTable map[string]*TypeSymbol

func (t TypeTable) Lookup(name string) *TypeSymbol {
	val, exist := t[name]
	if !exist {
		return nil
	}
	return val
}

type TypeAnalyzer struct {
	ErrorCollector
	current *TypeSymbol
	table   TypeTable
}

var (
	PrimitiveNull    = NewType("null", Primitive)
	PrimitiveInt     = NewType("int", Primitive)
	PrimitiveBoolean = NewType("boolean", Primitive)
	PrimitiveChar    = NewType("char", Primitive)
	PrimitiveString  = NewType("String", Primitive)
)

func NewTypeAnalyzer() *TypeAnalyzer {
	return &TypeAnalyzer{
		make([]string, 0),
		nil,
		TypeTable{
			"null":    PrimitiveNull,
			"int":     PrimitiveInt,
			"boolean": PrimitiveBoolean,
			"char":    PrimitiveChar,
			"String":  PrimitiveString,
		},
	}
}

func (t *TypeAnalyzer) GetTypeTable() map[string]*TypeSymbol {
	return t.table
}

func (t *TypeAnalyzer) VisitProgram(program text.Program) {}

func (t *TypeAnalyzer) typeExist(name string) bool {
	_, exist := t.table[name]
	return exist
}

//FIXME: Code duplicate here, but if removed will be too unreadable
//fix it
func (t *TypeAnalyzer) VisitClass(class *text.Class) {
	notExistMsg := func(name string) {
		t.AddErrorf(msgTypeNotExist, name)
	}

	if _, exist := t.table[class.Name]; exist {
		t.AddErrorf(msgTypeAlreadyDeclared, class.Name)
		return
	}

	newClass := NewType(class.Name, Class)
	declareClass := func(name string, cat TypeCategory) {
		if len(name) == 0 {
			return
		}
		if !t.typeExist(name) {
			notExistMsg(name)
			return
		}

		val := t.table[name]
		if val.TypeCategory != cat {
			msg := msgExtendShouldBeOnClass
			if cat == Interface {
				msg = msgImplementShouldBeOnInterface
			}
			t.AddError(msg)
			return
		}

		if cat == Class {
			newClass.extends = val
		} else {
			newClass.implements = val
		}
	}

	declareClass(class.Extend, Class)
	declareClass(class.Implement, Interface)

	t.current = newClass
	t.table[newClass.name] = newClass

}
func (t *TypeAnalyzer) VisitAfterClass(class *text.Class) {
	t.addConstructorIfEmpty(class.Name)
	inf := t.current.implements
	if inf == nil {
		return
	}

	// check for lack of implemented methods
	for key := range inf.Methods {
		_, exist := t.current.Methods[key]
		if !exist {
			t.AddErrorf(msgMustImplementMethod, key)
		}
	}
}

func (t *TypeAnalyzer) addConstructorIfEmpty(name string) {
	for key := range t.table[name].Methods {
		fmt.Println("I Have", key)
		if strings.HasPrefix(key, name+"(") {
			return
		}
	}

	signature := name + "()"
	t.table[name].Methods[signature] = &MethodSymbol{
		DataType{t.table[name], false},
		text.Public,
		name,
		signature,
	}

}

func (t *TypeAnalyzer) VisitInterface(inf *text.Interface) {
	if _, exist := t.table[inf.Name]; exist {
		t.AddErrorf(msgTypeAlreadyDeclared, inf.Name)
		return
	}

	t.current = NewType(inf.Name, Interface)
	t.table[inf.Name] = t.current
}

func (t *TypeAnalyzer) VisitPropertyDeclaration(prop *text.PropertyDeclaration) {
	if !t.typeExist(prop.Type.Name) {
		t.AddErrorf(msgTypeNotExist, prop.Type.Name)
		return
	}

	if _, isExist := t.current.Properties[prop.Name]; isExist {
		t.AddErrorf(msgPropertyAlreadyDeclared, prop.Name, t.current.name)
		return
	}

	t.current.Properties[prop.Name] = &PropertySymbol{
		prop.AccessModifier,
		FieldSymbol{
			DataType{t.table[prop.Type.Name], prop.Type.IsArray},
			prop.Name,
		},
	}
}

func (t *TypeAnalyzer) VisitMethodSignature(signature *text.MethodSignature) {
	var typeof *TypeSymbol

	if _, exist := t.current.Properties[signature.Name]; exist {
		t.AddErrorf(msgMethodAlreadyDeclaredAsProp, signature.Name)
	} else if _, exist := t.current.Methods[signature.Signature()]; exist {
		t.AddErrorf(msgMethodIsAlreadyDeclared, signature.Signature())
	}

	if signature.ReturnType.Name == "void" {
		typeof = NewType("void", Primitive)
	} else if !t.typeExist(signature.ReturnType.Name) {
		t.AddErrorf(msgTypeNotExist, signature.ReturnType.Name)
		return
	} else {
		typeof = t.table[signature.ReturnType.Name]
	}

	parameters := make([]*FieldSymbol, len(signature.ParameterList))
	for i, param := range signature.ParameterList {
		if !t.typeExist(param.Type.Name) {
			t.AddErrorf(msgTypeNotExist, param.Name)
			continue
		}

		parameters[i] = &FieldSymbol{
			DataType{
				t.table[param.Type.Name],
				param.Type.IsArray,
			},
			param.Name,
		}
	}

	t.current.Methods[signature.Signature()] = &MethodSymbol{
		DataType{typeof, signature.ReturnType.IsArray},
		signature.AccessModifier,
		signature.Name,
		signature.Signature(),
	}
}

func (t *TypeAnalyzer) VisitMainMethodDeclaration(*text.MainMethodDeclaration)  {}
func (t *TypeAnalyzer) VisitMethodDeclaration(*text.MethodDeclaration)          {}
func (t *TypeAnalyzer) VisitAfterMethodDeclaration(*text.MethodDeclaration)     {}
func (t *TypeAnalyzer) VisitVariableDeclaration(*text.VariableDeclaration)      {}
func (t *TypeAnalyzer) VisitAfterVariableDeclaration(*text.VariableDeclaration) {}
func (t *TypeAnalyzer) VisitStatementList(text.StatementList)                   {}
func (t *TypeAnalyzer) VisitAfterStatementList()                                {}
func (t *TypeAnalyzer) VisitSwitchStatement(*text.SwitchStatement)              {}
func (t *TypeAnalyzer) VisitAfterSwitchStatement(*text.SwitchStatement)         {}
func (t *TypeAnalyzer) VisitSwitchCase(*text.CaseStatement)                     {}
func (t *TypeAnalyzer) VisitIfStatement(*text.IfStatement)                      {}
func (t *TypeAnalyzer) VisitAfterIfStatementCondition(*text.IfStatement)        {}
func (t *TypeAnalyzer) VisitAfterIfStatementBody(*text.IfStatement)             {}
func (t *TypeAnalyzer) VisitAfterElseStatementBody(*text.IfStatement)           {}
func (t *TypeAnalyzer) VisitAfterIfStatement(*text.IfStatement)                 {}
func (t *TypeAnalyzer) VisitForStatement(*text.ForStatement)                    {}
func (t *TypeAnalyzer) VisitAfterForStatementCondition(*text.ForStatement)      {}
func (t *TypeAnalyzer) VisitWhileStatement(*text.WhileStatement)                {}
func (t *TypeAnalyzer) VisitAfterWhileStatementCondition(*text.WhileStatement)  {}
func (t *TypeAnalyzer) VisitAssignmentStatement(*text.AssignmentStatement)      {}
func (t *TypeAnalyzer) VisitAfterAssignmentStatement(*text.AssignmentStatement) {}
func (t *TypeAnalyzer) VisitJumpStatement(*text.JumpStatement)                  {}
func (t *TypeAnalyzer) VisitAfterJumpStatement(*text.JumpStatement)             {}
func (t *TypeAnalyzer) VisitFieldAccess(*text.FieldAccess)                      {}
func (t *TypeAnalyzer) VisitArrayAccess(*text.ArrayAccess)                      {}
func (t *TypeAnalyzer) VisitAfterArrayAccess(*text.ArrayAccess)                 {}
func (t *TypeAnalyzer) VisitArrayAccessDelegate(text.NamedValue)                {}
func (t *TypeAnalyzer) VisitMethodCall(*text.MethodCall)                        {}
func (t *TypeAnalyzer) VisitAfterMethodCall(*text.MethodCall)                   {}
func (t *TypeAnalyzer) VisitArrayCreation(*text.ArrayCreation)                  {}
func (t *TypeAnalyzer) VisitAfterArrayCreation(*text.ArrayCreation)             {}
func (t *TypeAnalyzer) VisitObjectCreation(*text.ObjectCreation)                {}
func (t *TypeAnalyzer) VisitBinOp(*text.BinOp)                                  {}
func (t *TypeAnalyzer) VisitAfterBinOp(*text.BinOp)                             {}
func (t *TypeAnalyzer) VisitConstant(text.Expression)                           {}
func (t *TypeAnalyzer) VisitSystemOut()                                         {}
func (t *TypeAnalyzer) VisitAfterSystemOut()                                    {}
