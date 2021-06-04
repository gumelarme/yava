package lang

import (
	"fmt"
	"strings"

	"github.com/gumelarme/yava/pkg/text"
)

var (
	msgParameterAlreadyDeclared = "Parameter %s already declared."
	msgVariableAlreadyDeclared  = "Variable %s is already declared in this scope."
	msgVariableDoesNotExist     = "Variable %s does not exist in this scope."
	msgTypeDoesNotHaveProperty  = "Type '%s' does not have a '%s' property"
	msgMethodNotFound           = "Method '%s' not found"
	msgExpectArrayAccess        = "Expecting array acces in field '%s'"
	msgFieldIsNotArray          = "Field %s is not an array"
	msgExpectingTypeof          = "Expecting a type of '%s' but got '%s' instead."
	msgExpectingReturnTypeOf    = "Expecting a return type of '%s' but got '%s' instead."
	msgVoidDontHaveType         = "The function return type is void, but got '%s'"
	msgCantBeNull               = "Type of %s can be assigned with null value."
)

type TypeStack []DataType

func (t *TypeStack) Pop() (DataType, error) {
	if l := len(*t); l == 0 {
		return DataType{}, fmt.Errorf("Cannot pop, stack is empty.")
	}

	n := len(*t) - 1
	dataType := (*t)[n]
	*t = (*t)[:n]
	return dataType, nil
}

func (t *TypeStack) Push(d DataType) {
	*t = append(*t, d)
}

func (t *TypeStack) Overwrite(d DataType) {
	(*t)[len(*t)-1] = d
}

type NameAnalyzer struct {
	ErrorCollector
	typeTable      TypeTable
	scope          SymbolTable
	error          []string
	isScopeCreated bool
	counter        int
	Tables         []*SymbolTable
	curField       TypeMember
	stack          TypeStack
}

func NewNameAnalyzer(table map[string]*TypeSymbol) *NameAnalyzer {
	return &NameAnalyzer{
		ErrorCollector{},
		table,
		NewSymbolTable("<program>", 0, nil),
		make([]string, 0),
		false,
		0,
		make([]*SymbolTable, 0),
		nil,
		TypeStack{},
	}
}

func (n *NameAnalyzer) newScope(name string) {
	parent := n.scope
	n.counter += 1
	name = fmt.Sprintf("%s_%d", name, n.counter)
	newScope := NewSymbolTable(name, parent.level+1, &parent)
	n.scope = newScope
	n.Tables = append(n.Tables, &newScope)
}

func (n *NameAnalyzer) expectLastStackTypeOf(name string, array bool) bool {
	lastStack, _ := n.stack.Pop()
	sym := n.typeTable.Lookup(name)
	expect := DataType{sym, array}
	if lastStack != expect {
		n.AddErrorf(msgExpectingTypeof, expect, lastStack)
		return false
	}
	return true
}

func IsNullOk(dt DataType) bool {
	if dt.isArray {
		return true
	}

	switch dt.dataType.name {
	case "int", "boolean", "char":
		return false
	default:
		return true
	}
}

func (n *NameAnalyzer) getAccess() text.AccessModifier {
	access := text.Public
	if n.curField != nil && n.curField.Name() == "this" {
		access |= text.Protected
		access |= text.Private
	}
	return access
}

func (n *NameAnalyzer) VisitProgram(text.Program)      {}
func (n *NameAnalyzer) VisitInterface(*text.Interface) {}
func (n *NameAnalyzer) VisitClass(class *text.Class) {
	name := fmt.Sprintf("class-%s", class.Name)
	n.newScope(name)
	classType := n.typeTable[class.Name]
	for _, prop := range classType.Properties {
		n.scope.Insert(prop)
	}

	for _, method := range classType.Methods {
		n.scope.Insert(method)
	}

	n.stack.Push(DataType{
		classType,
		false,
	})
}

// REVIEW: Shold we pop scope here
func (n *NameAnalyzer) VisitAfterClass(*text.Class) {
	n.stack.Pop()
}
func (n *NameAnalyzer) VisitPropertyDeclaration(*text.PropertyDeclaration) {}

func (n *NameAnalyzer) VisitMethodSignature(sign *text.MethodSignature) {
	n.isScopeCreated = true
	n.newScope(fmt.Sprintf("method-%s", sign.Signature()))

	classType, _ := n.stack.Pop()
	n.scope.Insert(&FieldSymbol{
		classType,
		"this",
	})

	n.stack.Push(classType)
	var returnType DataType
	if sign.ReturnType.Name == "void" {
		returnType = DataType{
			NewType("void", Primitive),
			sign.ReturnType.IsArray,
		}
	} else {
		typeof := n.typeTable.Lookup(sign.ReturnType.Name)
		returnType = DataType{
			typeof,
			sign.ReturnType.IsArray,
		}
	}
	n.stack.Push(returnType)

	for _, param := range sign.ParameterList {
		typeof := n.typeTable[param.Type.Name]

		if n.scope.Lookup(param.Name, true) != nil {
			n.AddErrorf(msgParameterAlreadyDeclared, param.Name)
			return
		}

		n.scope.Insert(&FieldSymbol{
			DataType{typeof, param.Type.IsArray},
			param.Name,
		})
	}

}

func (n *NameAnalyzer) VisitMethodDeclaration(*text.MethodDeclaration)         {}
func (n *NameAnalyzer) VisitMainMethodDeclaration(*text.MainMethodDeclaration) {}
func (n *NameAnalyzer) VisitAfterMethodDeclaration(*text.MethodDeclaration) {
	//popping method return type
	n.stack.Pop()
}
func (n *NameAnalyzer) VisitVariableDeclaration(varDecl *text.VariableDeclaration) {
	varName, typeName := varDecl.Name, varDecl.Type.Name
	typeof := n.typeExist(typeName)

	if typeof == nil {
		return
	}

	symbol := n.scope.Lookup(varName, false)
	if symbol != nil {
		n.AddErrorf(msgVariableAlreadyDeclared, varName)
		return
	}
}

func (n *NameAnalyzer) typeExist(name string) *TypeSymbol {
	typeof := n.typeTable.Lookup(name)
	if typeof == nil {
		n.AddErrorf(msgTypeNotExist, name)
		return nil
	}
	return typeof
}

func (n *NameAnalyzer) VisitAfterVariableDeclaration(varDecl *text.VariableDeclaration) {
	typeof := n.typeTable.Lookup(varDecl.Type.Name)
	canDeclare := true
	defer func() {
		if !canDeclare {
			return
		}

		n.scope.Insert(&FieldSymbol{
			DataType{
				typeof,
				varDecl.Type.IsArray,
			},
			varDecl.Name,
		})
	}()

	if varDecl.Value == nil {
		return
	}

	varType := DataType{typeof, varDecl.Type.IsArray}
	expressionType, err := n.stack.Pop()

	if err != nil {
		canDeclare = false
		return
	}

	if IsNullOk(varType) && expressionType.Name() == "null" {
		n.AddErrorf(msgCantBeNull, varType)
		canDeclare = false
		return
	}

	if varType != expressionType {
		n.AddErrorf(msgExpectingTypeof, varType, expressionType)
		canDeclare = false
		return
	}

}

func (n *NameAnalyzer) VisitStatementList(text.StatementList) {
	if n.isScopeCreated {
		n.isScopeCreated = false
	} else {
		n.newScope(fmt.Sprintf("block-%d", n.scope.level))
	}
}

func (n *NameAnalyzer) VisitAfterStatementList() {
	// pop scope
	n.scope = *n.scope.parent
}
func (n *NameAnalyzer) VisitSwitchStatement(*text.SwitchStatement) {
	n.stack.Push(n.curField.Type())
	n.curField = nil
}

func (n *NameAnalyzer) VisitAfterSwitchStatement(*text.SwitchStatement) {
	// remove switch type
	n.stack.Pop()
}

func (n *NameAnalyzer) VisitSwitchCase(cs *text.CaseStatement) {
	caseType, _ := n.stack.Pop()
	switchType, _ := n.stack.Pop()
	if caseType != switchType {
		n.AddErrorf(msgExpectingTypeof, switchType.dataType.name, caseType.dataType.name)
	}
	n.stack.Push(switchType)
}
func (n *NameAnalyzer) VisitIfStatement(*text.IfStatement) {}
func (n *NameAnalyzer) VisitAfterIfStatementCondition(*text.IfStatement) {
	n.expectLastStackTypeOf("boolean", false)
}

func (n *NameAnalyzer) VisitForStatement(forStmt *text.ForStatement) {
	if forStmt.Init == nil {
		return
	}

	name, _ := forStmt.Init.NodeContent()
	if name == "var-decl" {
		name := fmt.Sprintf("for-scope-%d", n.scope.level)
		n.newScope(name)
		n.isScopeCreated = true
	}
}

func (n *NameAnalyzer) VisitAfterForStatementCondition(forStmt *text.ForStatement) {
	n.expectLastStackTypeOf("int", false)
}

func (n *NameAnalyzer) VisitWhileStatement(*text.WhileStatement) {}
func (n *NameAnalyzer) VisitAfterWhileStatementCondition(*text.WhileStatement) {
	n.expectLastStackTypeOf("boolean", false)
}
func (n *NameAnalyzer) VisitAssignmentStatement(*text.AssignmentStatement) {}
func (n *NameAnalyzer) VisitAfterAssignmentStatement(*text.AssignmentStatement) {
	rightType, _ := n.stack.Pop()

	if IsNullOk(n.curField.Type()) && rightType.Name() == "null" {
		return
	}

	if !rightType.Equals(n.curField.Type()) {
		n.AddErrorf(msgExpectingTypeof, n.curField.Type(), rightType)
	}
}

func (n *NameAnalyzer) VisitJumpStatement(*text.JumpStatement) {}
func (n *NameAnalyzer) VisitAfterJumpStatement(jump *text.JumpStatement) {
	fmt.Println("Stack", n.stack)
	val, _ := n.stack.Pop()
	retType, _ := n.stack.Pop()
	n.stack.Push(retType)

	if retType.dataType.name == "void" && jump.Exp != nil {
		n.AddErrorf(msgVoidDontHaveType, val)
		return
	}

	if retType != val {
		n.AddErrorf(msgExpectingReturnTypeOf, retType, val)
	}
}
func (n *NameAnalyzer) VisitFieldAccess(field *text.FieldAccess) {
	if n.curField == nil {
		sym := n.scope.Lookup(field.Name, true)
		if sym == nil {
			n.AddErrorf(msgVariableDoesNotExist, field.Name)
		} else {
			n.curField = sym
			n.stack.Push(sym.Type())
		}
		return
	}

	if n.curField.Type().isArray {
		n.AddErrorf(msgExpectArrayAccess, field.Name)
		return
	}

	subField := n.curField.Type().dataType.Properties[field.Name]
	propNotFoundMsg := fmt.Sprintf(msgTypeDoesNotHaveProperty, n.curField.Type().Name(), field.Name)
	if subField == nil {
		n.AddError(propNotFoundMsg)
		return
	}

	access := n.getAccess()
	if n.curField != nil && subField.AccessModifier&access == 0 {
		n.AddError(propNotFoundMsg)
		return
	}

	n.curField = subField
	n.stack.Overwrite(subField.DataType)
}

func (n *NameAnalyzer) VisitArrayAccess(arr *text.ArrayAccess) {
	if !n.curField.Type().isArray {
		n.AddErrorf(msgFieldIsNotArray, n.curField.Name())
	}
}

func (n *NameAnalyzer) VisitAfterArrayAccess(arr *text.ArrayAccess) {
	n.expectLastStackTypeOf("int", false)
}

func (n *NameAnalyzer) VisitArrayAccessDelegate(text.NamedValue) {}
func (n *NameAnalyzer) VisitMethodCall(*text.MethodCall)         {}
func (n *NameAnalyzer) VisitAfterMethodCall(method *text.MethodCall) {
	argCount := make([]string, len(method.Args))

	for i := range argCount {
		typeof, _ := n.stack.Pop()
		argCount[len(argCount)-i-1] = typeof.String()
	}

	signature := fmt.Sprintf("%s(%s)", method.Name, strings.Join(argCount, ", "))
	var emptyMethod *MethodSymbol
	var exist TypeMember
	if n.curField != nil {
		exist = n.curField.Type().dataType.LookupMethod(signature)
	} else {
		exist = n.scope.Lookup(signature, true)
	}

	// weird emptyMethod
	if exist == nil || exist == emptyMethod {
		n.AddErrorf(msgMethodNotFound, signature)
		return
	}

	access := n.getAccess()

	methodSym := exist.(*MethodSymbol)
	if methodSym != nil && methodSym.accessMod&access == 0 {
		n.AddErrorf(msgMethodNotFound, signature)
		return
	}

	if n.curField == nil {
		n.stack.Push(methodSym.DataType)
	} else {
		n.stack.Overwrite(methodSym.DataType)
	}

	if method.ChildNode() != nil {
		n.curField = methodSym
	}
}

func (n *NameAnalyzer) VisitArrayCreation(arr *text.ArrayCreation) {
	if n.typeExist(arr.Type) == nil {
		return
	}
}

func (n *NameAnalyzer) VisitAfterArrayCreation(arr *text.ArrayCreation) {
	if !n.expectLastStackTypeOf("int", false) {
		return
	}

	typeof := n.typeTable.Lookup(arr.Type)
	n.stack.Push(DataType{
		typeof,
		true,
	})
}

func (n *NameAnalyzer) VisitObjectCreation(*text.ObjectCreation) {}
func (n *NameAnalyzer) VisitBinOp(bin *text.BinOp)               {}

func (n *NameAnalyzer) mustBeTypeof(left, right DataType, types ...string) bool {
	leftName, rightName := left.dataType.name, right.dataType.name
	if leftName != rightName {
		n.AddErrorf(msgExpectingTypeof, leftName, rightName)
		return false
	}

	typeOk := false
	for _, name := range types {
		if leftName == name {
			typeOk = true
		}
	}

	typeString := strings.Join(types, ", ")
	if !typeOk {
		n.AddErrorf(msgExpectingTypeof, typeString, leftName)
		return false
	}

	return true
}

func (n *NameAnalyzer) VisitAfterBinOp(bin *text.BinOp) {
	right, _ := n.stack.Pop()
	left, _ := n.stack.Pop()
	evaluate := func(being string, types ...string) {
		if !n.mustBeTypeof(left, right, types...) {
			return
		}

		n.stack.Push(DataType{
			n.typeTable.Lookup(being),
			false,
		})
	}

	operator := bin.GetOperator()
	switch operator.Type {
	case text.Addition,
		text.Subtraction,
		text.Multiplication,
		text.Division,
		text.Modulus:
		evaluate("int", "int")

	case text.GreaterThan,
		text.GreaterThanEqual,
		text.LessThan,
		text.LessThanEqual:
		evaluate("boolean", "int")

	case text.Or, text.And:
		evaluate("boolean", "boolean")

	case text.Equal, text.NotEqual:
		evaluate("boolean", "boolean", "char", "int")
	}
}

func (n *NameAnalyzer) VisitConstant(ex text.Expression) {
	typeof, _ := ex.NodeContent()
	switch typeof {
	case "String", "int", "char", "boolean", "null":
		dataType := DataType{
			n.typeTable.Lookup(typeof),
			false,
		}
		n.stack.Push(dataType)
	case "this":
		this := n.scope.Lookup("this", true)
		n.stack.Push(this.Type())
		n.curField = &FieldSymbol{
			DataType{this.Type().dataType, false},
			"this",
		}
	}
}

func (n *NameAnalyzer) VisitSystemOut()      {}
func (n *NameAnalyzer) VisitAfterSystemOut() {}
