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
	typeTable        TypeTable
	scope            SymbolTable
	error            []string
	isScopeCreated   bool
	counter          int
	Tables           []*SymbolTable
	curField         TypeMember
	stack            TypeStack
	localCount       int
	fieldBuffer      TypeMember
	isInterface      bool
	isObjectCreation bool
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
		0,
		nil,
		false,
		false,
	}
}

func (n *NameAnalyzer) popScope() {
	// every popped is added to the tables again, so the access could be linear
	n.Tables = append(n.Tables, n.scope.parent)
	// pop scope
	n.scope = *n.scope.parent
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

func (n *NameAnalyzer) Insert(member TypeMember) {
	n.scope.Insert(member, n.localCount)
	n.localCount += 1
}

func (n *NameAnalyzer) VisitProgram(text.Program) {}
func (n *NameAnalyzer) VisitInterface(i *text.Interface) {
	n.isScopeCreated = false
	n.newScope(fmt.Sprintf("interface-%s", i.Name))
	n.isInterface = true
	interfaceType := n.typeTable[i.Name]
	n.stack.Push(DataType{
		interfaceType,
		false,
	})
}

func (n *NameAnalyzer) VisitAfterInterface(*text.Interface) {
	n.isInterface = false
	n.popScope()
	n.stack.Pop()
}

func (n *NameAnalyzer) VisitClass(class *text.Class) {
	n.isScopeCreated = false
	name := fmt.Sprintf("class-%s", class.Name)
	n.newScope(name)
	classType := n.typeTable[class.Name]
	for _, prop := range classType.Properties {
		n.localCount = 0
		n.Insert(prop)
	}

	for _, method := range classType.Methods {
		n.localCount = 0
		n.Insert(method)
	}
	n.stack.Push(DataType{
		classType,
		false,
	})
}

// REVIEW: Shold we pop scope here
func (n *NameAnalyzer) VisitAfterClass(*text.Class) {
	n.popScope()
	n.stack.Pop()
}
func (n *NameAnalyzer) VisitPropertyDeclaration(*text.PropertyDeclaration) {}

func (n *NameAnalyzer) VisitMethodSignature(sign *text.MethodSignature) {
	defer func() {
		if n.isInterface {
			n.popScope()
			n.stack.Pop()
		}
	}()

	n.localCount = 0 // reset
	n.isScopeCreated = true
	n.newScope(fmt.Sprintf("method-%s", sign.Signature()))

	classType, _ := n.stack.Pop()
	n.Insert(&FieldSymbol{
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
	n.registerParam(sign.ParameterList)
}

func (n *NameAnalyzer) registerParam(params []text.Parameter) {
	for _, param := range params {
		typeof := n.typeTable[param.Type.Name]

		if exist, _ := n.scope.Lookup(param.Name, false); exist != nil {
			n.AddErrorf(msgParameterAlreadyDeclared, param.Name)
			return
		}

		n.Insert(&FieldSymbol{
			DataType{typeof, param.Type.IsArray},
			param.Name,
		})
	}
}

func (n *NameAnalyzer) VisitMethodDeclaration(*text.MethodDeclaration) {}
func (n *NameAnalyzer) VisitConstructor(con *text.ConstructorDeclaration) {
	n.localCount = 0 // reset
	n.isScopeCreated = true
	n.newScope(fmt.Sprintf("constructor-%s", con.Signature()))

	classType, _ := n.stack.Pop()
	n.Insert(&FieldSymbol{
		classType,
		"this",
	})
	n.stack.Push(classType)
	n.stack.Push(DataType{
		NewType("void", Primitive),
		false,
	})
	n.registerParam(con.ParameterList)
}

func (n *NameAnalyzer) VisitAfterConstructor(*text.ConstructorDeclaration) {
	n.stack.Pop()
}

func (n *NameAnalyzer) VisitMainMethodDeclaration(*text.MainMethodDeclaration) {
	n.localCount = 0
	returnType := DataType{
		NewType("void", Primitive),
		false,
	}
	n.stack.Push(returnType)
}
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

	symbol, _ := n.scope.Lookup(varName, false)
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

		n.Insert(&FieldSymbol{
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

	if !isTypeValid(varType, expressionType) {
		n.AddErrorf(msgExpectingTypeof, varType, expressionType)
		canDeclare = false
		return
	}

}

func isTypeValid(varType, expressionType DataType) bool {
	if varType.isArray != expressionType.isArray {
		return false
	}

	if varType == expressionType {
		return true
	}

	if expressionType.dataType.isDescendantOf(varType.dataType) {
		return true
	}

	if expressionType.dataType.isImplementing(varType.dataType) {
		return true
	}

	return false
}

func (n *NameAnalyzer) VisitStatementList(text.StatementList) {
	if n.isScopeCreated {
		n.isScopeCreated = false
	} else {
		n.newScope(fmt.Sprintf("block-%d", n.scope.level))
	}
}

func (n *NameAnalyzer) VisitAfterStatementList() {
	n.popScope()
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

func (n *NameAnalyzer) VisitAfterIfStatementBody(*text.IfStatement)   {}
func (n *NameAnalyzer) VisitAfterIfStatement(*text.IfStatement)       {}
func (n *NameAnalyzer) VisitAfterElseStatementBody(*text.IfStatement) {}
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

func (n *NameAnalyzer) VisitAfterForStatementInit(*text.ForStatement) {}
func (n *NameAnalyzer) VisitAfterForStatementCondition(forStmt *text.ForStatement) {
	n.expectLastStackTypeOf("boolean", false)
}
func (n *NameAnalyzer) VisitBeforeForStatementUpdate(*text.ForStatement) {}
func (n *NameAnalyzer) VisitAfterForStatement(*text.ForStatement)        {}
func (n *NameAnalyzer) VisitWhileStatement(*text.WhileStatement)         {}
func (n *NameAnalyzer) VisitAfterWhileStatement(*text.WhileStatement)    {}
func (n *NameAnalyzer) VisitAfterWhileStatementCondition(*text.WhileStatement) {
	n.expectLastStackTypeOf("boolean", false)
}
func (n *NameAnalyzer) VisitAssignmentStatement(*text.AssignmentStatement) {}
func (n *NameAnalyzer) VisitAfterAssignmentStatement(*text.AssignmentStatement) {
	targetType, _ := n.stack.Pop()
	rightType, _ := n.stack.Pop()

	if IsNullOk(targetType) && rightType.Name() == "null" {
		return
	}

	if !rightType.Equals(targetType) {
		n.AddErrorf(msgExpectingTypeof, targetType, rightType)
	}
}

func (n *NameAnalyzer) VisitJumpStatement(*text.JumpStatement) {}
func (n *NameAnalyzer) VisitAfterJumpStatement(jump *text.JumpStatement) {
	if jump.Type != text.ReturnJump {
		return
	}

	// return type always at the second index
	retType := n.stack[1]
	if retType.dataType.name == "void" {
		if jump.Exp != nil {
			val, _ := n.stack.Pop()
			n.AddErrorf(msgVoidDontHaveType, val)
		}
		return
	}

	val, _ := n.stack.Pop()
	if retType != val {
		n.AddErrorf(msgExpectingReturnTypeOf, retType, val)
	}
}

func (n *NameAnalyzer) VisitFieldAccess(field *text.FieldAccess) {
	if n.curField == nil {
		sym, _ := n.scope.Lookup(field.Name, true)
		if sym == nil {
			n.AddErrorf(msgVariableDoesNotExist, field.Name)
		} else {
			if field.Child != nil {
				n.curField = sym
			}
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

	if field.Child != nil {
		n.curField = subField
	} else {
		n.curField = nil
	}

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
func (n *NameAnalyzer) VisitMethodCall(*text.MethodCall) {
	n.fieldBuffer = n.curField
	n.curField = nil
}

func (n *NameAnalyzer) getFittingMethod(args []DataType, name string) (method *MethodSymbol) {
	argStr := make([]string, len(args))
	for i, a := range args {
		argStr[i] = a.String()
	}
	signature := fmt.Sprintf("%s(%s)", name, strings.Join(argStr, ", "))
	return n.getMethodBySignature(signature)
}

func (n *NameAnalyzer) getMethodBySignature(signature string) *MethodSymbol {
	var method TypeMember
	if n.curField != nil {
		method = n.curField.Type().dataType.LookupMethod(signature)
	} else {
		method, _ = n.scope.Lookup(signature, true)
	}

	var emptyMethod *MethodSymbol
	if method == nil || method == emptyMethod {
		n.AddErrorf(msgMethodNotFound, signature)
		return nil
	}

	return method.(*MethodSymbol)
}

func (n *NameAnalyzer) VisitAfterMethodCall(method *text.MethodCall) {
	if n.isObjectCreation {
		n.isObjectCreation = false
		return
	}

	if n.fieldBuffer != nil {
		n.curField = n.fieldBuffer
		n.fieldBuffer = nil
	}

	args := make([]DataType, len(method.Args))
	for i := range args {
		typeof, _ := n.stack.Pop()
		args[len(method.Args)-i-1] = typeof
	}

	methodSym := n.getFittingMethod(args, method.Name)
	access := n.getAccess()
	if methodSym != nil && methodSym.accessMod&access == 0 {
		argStr := make([]string, len(args))
		for i, a := range args {
			argStr[i] = a.String()
		}

		signature := fmt.Sprintf("%s(%s)", method.Name, strings.Join(argStr, ", "))
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
	} else {
		n.curField = nil
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

func (n *NameAnalyzer) VisitObjectCreation(*text.ObjectCreation) {
	n.isObjectCreation = true
}

func (n *NameAnalyzer) VisitAfterObjectCreation(o *text.ObjectCreation) {
	//TODO: implement parameterize object creation
	objectSymbol := n.typeTable[o.Name]
	if objectSymbol == nil {
		n.AddErrorf(msgTypeNotExist, o.Name)
		return
	}

	n.stack.Push(DataType{objectSymbol, false})
}

func (n *NameAnalyzer) VisitBinOp(bin *text.BinOp) {}

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
		this, _ := n.scope.Lookup("this", true)
		n.stack.Push(this.Type())
		n.curField = &FieldSymbol{
			DataType{this.Type().dataType, false},
			"this",
		}
	}
}

func (n *NameAnalyzer) VisitSystemOut() {}
func (n *NameAnalyzer) VisitAfterSystemOut() {
	n.curField = nil
	n.stack.Pop()
}
