package lang

import (
	"fmt"
	"strings"

	"github.com/gumelarme/yava/pkg/text"
)

const (
	MajorVersion     = 49
	MinorVersion     = 0
	InvokeJavaObject = "invokespecial Method java/lang/Object <init> ()V"
)

func isMathOperator(token text.TokenType) bool {
	switch token {
	case text.Addition, text.Subtraction, text.Multiplication, text.Division, text.Modulus:
		return true
	default:
		return false
	}
}

func codeConstant(exp text.Expression) (code string) {
	name, _ := exp.NodeContent()
	switch name {
	case "int":
		code = codeInt(exp.(text.Num))
	case "boolean":
		code = codeBoolean(exp.(text.Boolean))
	case "char":
		code = codeChar(exp.(text.Char))
	case "String":
		code = codeString(exp.(text.String))
	case "null":
		code = "aconst_null"
	}

	return
}

func codeInt(i text.Num) string {
	var format string
	switch {
	case i <= 5:
		format = "iconst_%d"
	case i <= 0xFF:
		format = "bipush %d"
	default:
		format = "ldc %d"
	}
	return fmt.Sprintf(format, i)
}

func codeChar(c text.Char) string {
	intVal := text.Num(c)
	return codeInt(intVal)
}

func codeBoolean(b text.Boolean) (result string) {
	if b {
		result = "iconst_1"
	} else {
		result = "iconst_0"
	}
	return
}

func codeString(s text.String) string {
	return fmt.Sprintf("ldc %#v", s)
}

func indent(code string, level int) string {
	var text []string
	for i := 1; i < level; i++ {
		text = append(text, "\t")
	}
	return strings.Join(text, "") + code
}

func labelCode(code string, labelNum int) string {
	return fmt.Sprintf("L%d:%s", labelNum, indent(code, 2))
}

type LS int

const (
	Load LS = iota
	Store
)

func loadOrStore(local Local, action LS) string {
	strArray := make([]string, 3)

	strArray[0] = "a"
	if IsPrimitive(local.Member.Type()) {
		strArray[0] = "i"
	}

	strArray[1] = "store"
	if action == Load {
		strArray[1] = "load"
	}

	strArray[2] = fmt.Sprintf("_%d", local.address)
	if local.address > 3 {
		strArray[2] = fmt.Sprintf(" %d", local.address)
	}

	return strings.Join(strArray, "")
}

func fieldDescriptor(name string, isArray bool) (result string) {
	switch name {
	case "void":
		return "V"
	case "int":
		result = "I"
	case "boolean":
		result = "Z"
	case "char":
		result = "C"
	case "String":
		result = "Ljava/lang/String;"
	default:
		result = fmt.Sprintf("L%s;", name)
	}

	if isArray {
		result = "[" + result
	}

	return result
}

type KrakatauGen struct {
	stackMax         int
	stackSize        int
	localCount       int
	labelCount       int
	outerLabel       int
	codes            []string
	codeBuffer       []string
	typeTable        TypeTable
	symbolTable      []*SymbolTable
	isScopeCreated   bool
	scopeIndex       int
	typeStack        TypeStack
	isAssignment     bool
	isObjectCreation bool
	hasField         bool
}

func NewKrakatauGen(typeTable TypeTable, symbolTables []*SymbolTable) *KrakatauGen {
	return &KrakatauGen{
		0,
		0,
		0,
		0,
		0,
		make([]string, 0),
		make([]string, 0),
		typeTable,
		symbolTables,
		false,
		-1,
		TypeStack{},
		false,
		false,
		false,
	}
}

func NewEmptyKrakatauGen() *KrakatauGen {
	typeTable := NewTypeAnalyzer().table
	return NewKrakatauGen(typeTable, nil)
}

func (c *KrakatauGen) Codes() []string {
	codes := make([]string, len(c.codes)+len(c.codeBuffer))
	i := 0
	for _, text := range c.codes {
		codes[i] = text
		i += 1
	}

	for _, text := range c.codeBuffer {
		codes[i] = text
		i += 1
	}

	return codes
}

func (c *KrakatauGen) GenerateCode() string {
	return strings.Join(c.codes, "\n")
}
func (c *KrakatauGen) Append(text string) {
	c.codes = append(c.codes, text)
}

func (c *KrakatauGen) AppendCode(text string) {
	c.codeBuffer = append(c.codeBuffer, text)
}

func (c *KrakatauGen) Lookup(name string) Local {
	// FIXME: Should this be always deep
	member, addr := c.symbolTable[c.scopeIndex].Lookup(name, true)
	return Local{member, addr}
}

func (c *KrakatauGen) incScopeIndex() {
	if !c.isScopeCreated {
		c.scopeIndex += 1
	}
}

func (c *KrakatauGen) resetStackSize() {
	c.stackSize, c.stackMax = 0, 0
}

func (c *KrakatauGen) decStackSize(count int) {
	c.stackSize -= count
}

func (c *KrakatauGen) incStackSize(count int) {
	c.stackSize += count
	if c.stackSize > c.stackMax {
		c.stackMax = c.stackSize
	}
}

func (c *KrakatauGen) getLabel() (result int) {
	result = c.labelCount
	c.labelCount += 1
	return
}

func (c *KrakatauGen) isCodeEndsWithReturn() bool {
	length := len(c.codeBuffer) - 1

	if length == -1 {
		return false
	}

	switch c.codeBuffer[length] {
	case "ireturn", "return", "areturn":
		return true
	default:
		return false
	}
}

func (c *KrakatauGen) getDefaultInitialization(t text.NamedType) (DataType, string) {
	typeof := c.typeTable.Lookup(t.Name)
	dt := DataType{typeof, t.IsArray}
	if IsPrimitive(dt) {
		return dt, "iconst_0"
	} else {
		return dt, "aconst_null"
	}
}

func (c *KrakatauGen) VisitProgram(program text.Program) {
	c.Append(fmt.Sprintf(".version %d %d", MajorVersion, MinorVersion))
}

func (c *KrakatauGen) VisitClass(class *text.Class) {
	c.incScopeIndex()
	declareClass := fmt.Sprintf(".class %s", class.Name)

	super := "java/lang/Object"
	if len(class.Extend) != 0 {
		super = class.Extend
	}

	declareSuper := fmt.Sprintf(".super %s", super)

	c.Append(declareClass)
	c.Append(declareSuper)

	if len(class.Implement) > 0 {
		c.Append(fmt.Sprintf(".implements %s", class.Implement))
	}
}

func (c *KrakatauGen) VisitAfterClass(class *text.Class) {
	c.makeDefaultConstructor(class.Name, class.Properties...)
	c.Append(".end class")
}

func (c *KrakatauGen) VisitInterface(*text.Interface) {}
func (c *KrakatauGen) VisitPropertyDeclaration(prop *text.PropertyDeclaration) {
	c.Append(fmt.Sprintf(".field %s %s %s",
		prop.AccessModifier,
		prop.Name,
		fieldDescriptor(prop.Type.Name, prop.Type.IsArray),
	))
}

func (c *KrakatauGen) makeDefaultConstructor(className string, prop ...*text.PropertyDeclaration) {
	c.localCount = 1
	// FIXME: Do something about the parameter
	header := fmt.Sprintf(".method <init> : (%s)V", "")
	c.Append(header)

	c.AppendCode("aload_0")
	c.AppendCode(InvokeJavaObject)
	c.incStackSize(1)
	c.decStackSize(1)

	for _, p := range prop {
		c.AppendCode("aload_0")
		c.incStackSize(1)

		if p.Value == nil {
			dt, code := c.getDefaultInitialization(p.Type)
			c.AppendCode(code)
			c.typeStack.Push(dt)
			c.incStackSize(1)
		} else {
			p.Value.Accept(c)
		}

		c.AppendCode(fmt.Sprintf("putfield Field %s %s %s",
			className,
			p.Name,
			fieldDescriptor(p.Type.Name, p.Type.IsArray),
		))
		// remove aload_0 and the value
		c.decStackSize(2)
	}

	c.AppendCode("return")
	//count stack
	c.Append(fmt.Sprintf(".code stack %d locals %d", c.stackMax, c.localCount))

	for _, code := range c.codeBuffer {
		c.Append(code)
	}

	c.codeBuffer = []string{}
	c.Append(".end code")
	c.Append(".end method")
}

func (c *KrakatauGen) VisitMethodSignature(signature *text.MethodSignature) {
	c.incScopeIndex()
	c.isScopeCreated = true
	c.localCount = 1
	params := make([]string, len(signature.ParameterList))
	for i, p := range signature.ParameterList {
		params[i] = fieldDescriptor(p.Type.Name, p.Type.IsArray)
	}

	returnType := fieldDescriptor(signature.ReturnType.Name, signature.ReturnType.IsArray)

	code := fmt.Sprintf(".method %s %s : (%s)%s",
		signature.AccessModifier,
		signature.Name,
		strings.Join(params, ""),
		returnType,
	)

	c.Append(code)
}

func (c *KrakatauGen) VisitMethodDeclaration(*text.MethodDeclaration) {}
func (c *KrakatauGen) VisitMainMethodDeclaration(*text.MainMethodDeclaration) {
	c.localCount = 1
	c.Append(".method public static main : ([Ljava/lang/String;)V")
}

func (c *KrakatauGen) VisitAfterMethodDeclaration(*text.MethodDeclaration) {
	defer func() {
		c.codeBuffer = make([]string, 0)
	}()

	c.Append(fmt.Sprintf(".code stack %d locals %d", c.stackMax, c.localCount))

	for _, code := range c.codeBuffer {
		c.Append(code)
	}

	if !c.isCodeEndsWithReturn() {
		// FIXME: determine the return type
		c.Append("return")
	}

	c.Append(".end code")
	c.Append(".end method")
}

func (c *KrakatauGen) VisitVariableDeclaration(v *text.VariableDeclaration) {}

func (c *KrakatauGen) VisitAfterVariableDeclaration(varDecl *text.VariableDeclaration) {
	// assign default value
	if varDecl.Value == nil {
		c.incStackSize(1)
		dt, code := c.getDefaultInitialization(varDecl.Type)
		c.AppendCode(code)
		c.typeStack.Push(dt)
	}

	c.localCount += 1
	local := c.Lookup(varDecl.Name)
	c.AppendCode(loadOrStore(local, Store))
	c.decStackSize(1)
}

func (c *KrakatauGen) VisitStatementList(text.StatementList) {
	c.incScopeIndex()
}
func (c *KrakatauGen) VisitAfterStatementList() {
	c.isScopeCreated = false
	c.incScopeIndex()
}

func (c *KrakatauGen) VisitSwitchStatement(*text.SwitchStatement)      {}
func (c *KrakatauGen) VisitSwitchCase(*text.CaseStatement)             {}
func (c *KrakatauGen) VisitAfterSwitchStatement(*text.SwitchStatement) {}

func gotoLabel(number int) string {
	return fmt.Sprintf("goto L%d", number)
}

func (c *KrakatauGen) VisitIfStatement(*text.IfStatement) {}
func (c *KrakatauGen) VisitAfterIfStatementCondition(ifStmt *text.IfStatement) {
	trueLabel, falseLabel := c.getLabel(), c.getLabel()
	if c.outerLabel == 0 {
		c.outerLabel = falseLabel
	}

	nextJump := falseLabel
	if ifStmt.Else != nil {
		nextJump = c.getLabel()
	} else if c.outerLabel > 0 {
		nextJump = c.outerLabel
	}

	c.AppendCode(fmt.Sprintf("ifne L%d", trueLabel))
	c.AppendCode(gotoLabel(nextJump))      // if false
	c.AppendCode(labelCode("", trueLabel)) // if true body
}

func (c *KrakatauGen) VisitAfterIfStatementBody(ifStmt *text.IfStatement) {
	c.AppendCode(gotoLabel(c.outerLabel))
	if ifStmt.Else != nil {
		c.AppendCode(labelCode("", c.labelCount-1)) // equals to nextJump
	}
}

func (c *KrakatauGen) VisitAfterElseStatementBody(ifStmt *text.IfStatement) {
	c.AppendCode(gotoLabel(c.outerLabel))
}

func (c *KrakatauGen) VisitAfterIfStatement(ifStmt *text.IfStatement) {
	c.AppendCode(labelCode("", c.outerLabel))
	c.outerLabel = 0
}

func (c *KrakatauGen) VisitForStatement(forStmt *text.ForStatement) {
	name, _ := forStmt.Init.NodeContent()
	if name == "var-decl" {
		c.incScopeIndex()
		c.isScopeCreated = true
	}
}
func (c *KrakatauGen) VisitAfterForStatementCondition(*text.ForStatement)     {}
func (c *KrakatauGen) VisitWhileStatement(*text.WhileStatement)               {}
func (c *KrakatauGen) VisitAfterWhileStatementCondition(*text.WhileStatement) {}
func (c *KrakatauGen) VisitAssignmentStatement(*text.AssignmentStatement) {
	c.isAssignment = true
}
func (c *KrakatauGen) VisitAfterAssignmentStatement(a *text.AssignmentStatement) {
	defer func() { c.isAssignment = false }()
	if a.Left.ChildNode() == nil {
		field := a.Left.(*text.FieldAccess)
		local := c.Lookup(field.Name)
		c.AppendCode(loadOrStore(local, Store))
	}
}
func (c *KrakatauGen) VisitJumpStatement(*text.JumpStatement) {}
func (c *KrakatauGen) VisitAfterJumpStatement(jump *text.JumpStatement) {
	defer c.decStackSize(1)
	if jump.ChildNode() == nil {
		c.AppendCode("return")
		return
	}

	data, _ := c.typeStack.Pop()
	if IsPrimitive(data) {
		c.AppendCode("ireturn")
	} else {
		c.AppendCode("areturn")
	}
}

func (c *KrakatauGen) VisitFieldAccess(field *text.FieldAccess) {
	defer func() {
		if field.Child != nil {
			c.hasField = true
		}
	}()

	if c.isAssignment {
		c.isAssignment = false
		return
	}

	if !c.hasField {
		local := c.Lookup(field.Name)
		c.typeStack.Push(local.Member.Type())
		c.AppendCode(loadOrStore(local, Load))
		c.incStackSize(1)
		return
	}

	dt, _ := c.typeStack.Pop()
	prop := dt.dataType.LookupProperty(field.Name)

	c.AppendCode(fmt.Sprintf("getfield Field %s %s %s",
		dt.dataType.name,
		prop.name,
		fieldDescriptor(prop.dataType.name, prop.isArray),
	))
	c.typeStack.Push(prop.DataType)
}

func (c *KrakatauGen) VisitArrayAccess(*text.ArrayAccess)       {}
func (c *KrakatauGen) VisitAfterArrayAccess(*text.ArrayAccess)  {}
func (c *KrakatauGen) VisitArrayAccessDelegate(text.NamedValue) {}
func (c *KrakatauGen) VisitMethodCall(*text.MethodCall)         {}
func (c *KrakatauGen) VisitAfterMethodCall(method *text.MethodCall) {
	if c.isObjectCreation {
		c.isObjectCreation = false
		return
	}

	var typeof, paramSignature, returnType string
	c.AppendCode(fmt.Sprintf("invokevirtual Method %s %s (%s)%s",
		typeof,
		method.Name,
		paramSignature,
		returnType,
	))
}

func (c *KrakatauGen) VisitArrayCreation(*text.ArrayCreation)      {}
func (c *KrakatauGen) VisitAfterArrayCreation(*text.ArrayCreation) {}
func (c *KrakatauGen) VisitObjectCreation(obj *text.ObjectCreation) {
	c.isObjectCreation = true
	c.AppendCode(fmt.Sprintf("new %s", obj.Name))
	c.AppendCode("dup")
	c.incStackSize(2)

	c.AppendCode(fmt.Sprintf("invokespecial Method %s <init> ()V", obj.Name))
	c.decStackSize(1)
}
func (c *KrakatauGen) VisitBinOp(*text.BinOp) {}

var opString = map[text.TokenType]string{
	text.Addition:         "iadd",
	text.Subtraction:      "isub",
	text.Multiplication:   "imul",
	text.Division:         "idiv",
	text.Modulus:          "imod",
	text.GreaterThan:      "if_icmpgt",
	text.GreaterThanEqual: "if_icmpgte",
	text.LessThan:         "if_icmplt",
	text.LessThanEqual:    "if_icmplte",
	text.Equal:            "if_icmpeq",
	text.NotEqual:         "if_icmpne",
	// text.And:              "if_icmpgt",
	// text.Or:               "if_icmpgt",
}

func (c *KrakatauGen) VisitAfterBinOp(bin *text.BinOp) {
	// use (remove) two operand, and place the result in the stack
	defer c.decStackSize(1)
	strOperator := opString[bin.GetOperator().Type]
	if isMathOperator(bin.GetOperator().Type) {
		c.AppendCode(strOperator)
		c.typeStack.Push(DataType{PrimitiveInt, false})
		return
	}

	// from here its boolean operation
	trueLabel, falseLabel := c.getLabel(), c.getLabel()
	c.AppendCode(fmt.Sprintf("%s L%d", strOperator, trueLabel))
	c.AppendCode(codeBoolean(false))
	c.AppendCode(fmt.Sprintf("goto L%d", falseLabel))
	c.AppendCode(labelCode(codeBoolean(true), trueLabel))
	c.AppendCode(labelCode("", falseLabel))
	c.typeStack.Push(DataType{PrimitiveBoolean, false})
}

func (c *KrakatauGen) VisitConstant(e text.Expression) {
	typename, _ := e.NodeContent()
	symbol := c.typeTable.Lookup(typename)
	c.typeStack.Push(DataType{symbol, false})

	c.incStackSize(1)
	c.AppendCode(codeConstant(e))
}

func (c *KrakatauGen) VisitSystemOut() {
	c.AppendCode("getstatic Field java/lang/System out Ljava/io/PrintStream;")
	c.incStackSize(1)
}

func sysOutDescriptor(dt DataType) string {
	if IsPrimitive(dt) {
		return fieldDescriptor(dt.dataType.name, dt.isArray)
	} else if dt.dataType == PrimitiveString && dt.isArray == false {
		return "Ljava/lang/String;"
	} else {
		return "Ljava/lang/Object;"
	}
}

func (c *KrakatauGen) VisitAfterSystemOut() {
	dt, _ := c.typeStack.Pop()
	argtype := sysOutDescriptor(dt)
	invoke := fmt.Sprintf("invokevirtual Method java/io/PrintStream println (%s)V", argtype)
	c.AppendCode(invoke)
	c.decStackSize(2)
}
