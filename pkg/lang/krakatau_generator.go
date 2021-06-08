package lang

import (
	"fmt"
	"strings"

	"github.com/gumelarme/yava/pkg/text"
)

const (
	MajorVersion = 49
	MinorVersion = 0
)

func invokeDefaultConstructor(name string) string {
	return fmt.Sprintf("invokespecial Method %s <init> ()V", name)
}

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
	case i <= 128:
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

type IntStack []int

func (i *IntStack) Pop() int {
	n := len(*i) - 1
	val := (*i)[n]
	*i = (*i)[:n]
	return val
}

func (i *IntStack) Push(val int) {
	*i = append(*i, val)
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
	loopOuter        IntStack
	loopHead         IntStack
	isInterface      bool
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
		0,
		TypeStack{},
		false,
		false,
		false,
		make([]int, 0),
		make([]int, 0),
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
	} else {
		c.isScopeCreated = false
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

func (c *KrakatauGen) combineCodes() {
	for _, code := range c.codeBuffer {
		c.Append(code)
	}
	// empty it
	c.codeBuffer = []string{}
}

func (c *KrakatauGen) getStackAndLocalCount() string {
	return fmt.Sprintf(".code stack %d locals %d", c.stackMax, c.localCount)
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
	c.makeDefaultConstructor(*class)
	c.Append(".end class")
	c.incScopeIndex()
}

func (c *KrakatauGen) VisitInterface(i *text.Interface) {
	c.isInterface = true
	c.Append(fmt.Sprintf(".class interface abstract %s", i.Name))
	c.Append(".super java/lang/Object")
	c.incScopeIndex()
}
func (c *KrakatauGen) VisitAfterInterface(*text.Interface) {
	c.Append(".end class")
	c.isInterface = false
	c.incScopeIndex()
}

func (c *KrakatauGen) VisitPropertyDeclaration(prop *text.PropertyDeclaration) {
	c.Append(fmt.Sprintf(".field %s %s %s",
		prop.AccessModifier,
		prop.Name,
		fieldDescriptor(prop.Type.Name, prop.Type.IsArray),
	))
}

func (c *KrakatauGen) makeDefaultConstructor(class text.Class) {
	c.localCount = 1
	// FIXME: Do something about the parameter
	header := fmt.Sprintf(".method <init> : (%s)V", "")
	c.Append(header)
	c.AppendCode("aload_0")

	object := "java/lang/Object"
	if len(class.Extend) > 0 {
		object = class.Extend
	}
	c.AppendCode(invokeDefaultConstructor(object))

	c.incStackSize(1)
	c.decStackSize(1)

	for _, p := range class.Properties {
		c.putProperties(class.Name, *p)
	}
	c.AppendCode("return")

	//count stack
	c.Append(c.getStackAndLocalCount())
	c.combineCodes()
	c.Append(".end code")
	c.Append(".end method")
}

func (c *KrakatauGen) putProperties(className string, p text.PropertyDeclaration) {
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

func (c *KrakatauGen) VisitMethodSignature(signature *text.MethodSignature) {
	c.incScopeIndex()
	c.isScopeCreated = true
	c.localCount = len(signature.ParameterList) + 1
	params := make([]string, len(signature.ParameterList))
	for i, p := range signature.ParameterList {
		params[i] = fieldDescriptor(p.Type.Name, p.Type.IsArray)
	}

	returnType := fieldDescriptor(signature.ReturnType.Name, signature.ReturnType.IsArray)

	var abstractModifier string
	if c.isInterface {
		abstractModifier = " abstract"
	}

	code := fmt.Sprintf(".method %s%s %s : (%s)%s",
		signature.AccessModifier,
		abstractModifier,
		signature.Name,
		strings.Join(params, ""),
		returnType,
	)

	c.Append(code)

	if c.isInterface {
		c.Append(".end method")
		c.incScopeIndex()
	}
}

func (c *KrakatauGen) VisitMethodDeclaration(*text.MethodDeclaration) {}
func (c *KrakatauGen) VisitMainMethodDeclaration(*text.MainMethodDeclaration) {
	c.localCount = 1
	c.Append(".method public static main : ([Ljava/lang/String;)V")
}

func (c *KrakatauGen) VisitAfterMethodDeclaration(*text.MethodDeclaration) {
	c.Append(c.getStackAndLocalCount())
	if !c.isCodeEndsWithReturn() {
		// FIXME: determine the return type
		c.AppendCode("return")
	}
	c.combineCodes()
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
	if !c.isScopeCreated {
		c.incScopeIndex()
	} else {
		c.isScopeCreated = false
	}
}
func (c *KrakatauGen) VisitAfterStatementList() {
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
	if forStmt.Init == nil {
		return
	}

	name, _ := forStmt.Init.NodeContent()
	if name == "var-decl" {
		c.incScopeIndex()
		c.isScopeCreated = true
	}
}

func (c *KrakatauGen) VisitAfterForStatementInit(forStmt *text.ForStatement) {
	conditionLabel := c.getLabel()
	c.loopHead.Push(conditionLabel)
	c.AppendCode(labelCode("", conditionLabel))
}
func (c *KrakatauGen) VisitAfterForStatementCondition(forStmt *text.ForStatement) {
	body, outer := c.getLabel(), c.getLabel()
	c.loopOuter.Push(outer)
	c.AppendCode(fmt.Sprintf("ifne L%d", body))
	c.AppendCode(gotoLabel(outer))
	c.AppendCode(labelCode("", body))

	if forStmt.Update != nil {
		forUpdateLabel := c.getLabel()
		c.loopHead.Push(forUpdateLabel)
	}
}

func (c *KrakatauGen) VisitBeforeForStatementUpdate(forStmt *text.ForStatement) {
	forUpdateLabel := c.loopHead.Pop()
	c.AppendCode(labelCode("", forUpdateLabel))
}
func (c *KrakatauGen) VisitAfterForStatement(forStmt *text.ForStatement) {
	conditionLabel := c.loopHead.Pop()
	c.AppendCode(gotoLabel(conditionLabel))

	if forStmt.Condition != nil {
		c.AppendCode(labelCode("", c.loopOuter.Pop()))
	}
}

func (c *KrakatauGen) VisitWhileStatement(*text.WhileStatement) {
	head := c.getLabel()
	c.AppendCode(labelCode("", head))
	c.loopHead.Push(head)
}

func (c *KrakatauGen) VisitAfterWhileStatementCondition(*text.WhileStatement) {
	whileBody, outer := c.getLabel(), c.getLabel()
	c.AppendCode(fmt.Sprintf("ifne L%d", whileBody))

	c.loopOuter.Push(outer)

	c.AppendCode(gotoLabel(outer))
	c.AppendCode(labelCode("", whileBody))
}

func (c *KrakatauGen) VisitAfterWhileStatement(*text.WhileStatement) {
	// popping
	head := c.loopHead.Pop()
	outer := c.loopOuter.Pop()

	c.AppendCode(gotoLabel(head))
	c.AppendCode(labelCode("", outer))
}

func (c *KrakatauGen) VisitAssignmentStatement(*text.AssignmentStatement) {
	c.isAssignment = true
}
func (c *KrakatauGen) VisitAfterAssignmentStatement(a *text.AssignmentStatement) {
	defer func() { c.isAssignment = false }()
	if a.Left.ChildNode() == nil {
		field := a.Left.(*text.FieldAccess)
		local := c.Lookup(field.Name)
		c.AppendCode(loadOrStore(local, Store))
		return
	}
	//pop the right one
	c.typeStack.Pop()
	// the parentField
	// the paretnField type name
	var parentField text.NamedValue
	parentField = a.Left
	leftType, _ := c.typeStack.Pop()
	parentFieldTypeName := leftType.dataType.name

	for {
		child := parentField.GetChild()
		if child != nil && child.GetChild() == nil {
			break
		}
		parentField = child
	}

	lastField := parentField.GetChild().(*text.FieldAccess)
	typeOfField := c.typeTable.Lookup(parentFieldTypeName)
	prop := typeOfField.LookupProperty(lastField.Name)
	c.AppendCode(fmt.Sprintf("putfield Field %s %s %s",
		parentFieldTypeName,
		lastField.Name,
		fieldDescriptor(prop.dataType.name, prop.isArray),
	))
}

func (c *KrakatauGen) VisitJumpStatement(*text.JumpStatement) {}
func (c *KrakatauGen) VisitAfterJumpStatement(jump *text.JumpStatement) {
	// assume its inside a loop
	if jump.Type != text.ReturnJump {
		labelSource := &c.loopHead
		if jump.Type == text.BreakJump {
			labelSource = &c.loopOuter
		}

		lbl := labelSource.Pop()
		c.AppendCode(gotoLabel(lbl))
		labelSource.Push(lbl) //put it back
		return
	}

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
		c.hasField = field.Child != nil
	}()

	if c.isAssignment && field.Child == nil {
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
func (c *KrakatauGen) VisitMethodCall(method *text.MethodCall) {
	c.hasField = method.Child != nil
}

func (c *KrakatauGen) VisitAfterMethodCall(method *text.MethodCall) {
	if c.isObjectCreation {
		c.isObjectCreation = false
		return
	}

	methodArgs := make([]string, len(method.Args))
	javaMethodArgs := make([]string, len(method.Args))
	for i := 0; i < len(method.Args); i++ {
		dt, _ := c.typeStack.Pop()
		javaMethodArgs[i] = fieldDescriptor(dt.dataType.name, dt.isArray)
		methodArgs[i] = dt.String()
	}
	methodSignature := fmt.Sprintf("%s(%s)", method.Name, strings.Join(methodArgs, ", "))
	javaMethodSignature := strings.Join(javaMethodArgs, "")

	dt, _ := c.typeStack.Pop()
	objectReferenceType := dt.dataType.name
	methodSymbol := dt.dataType.LookupMethod(methodSignature)
	returnType := fieldDescriptor(methodSymbol.Type().dataType.name, methodSymbol.isArray)
	c.AppendCode(fmt.Sprintf("invokevirtual Method %s %s (%s)%s",
		objectReferenceType,
		method.Name,
		javaMethodSignature,
		returnType,
	))
	c.typeStack.Push(methodSymbol.Type())
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
	text.Modulus:          "irem",
	text.GreaterThan:      "if_icmpgt",
	text.GreaterThanEqual: "if_icmpge",
	text.LessThan:         "if_icmplt",
	text.LessThanEqual:    "if_icmple",
	text.Equal:            "if_icmpeq",
	text.NotEqual:         "if_icmpne",
	// text.And:              "if_icmpgt",
	// text.Or:               "if_icmpgt",
}

func (c *KrakatauGen) VisitAfterBinOp(bin *text.BinOp) {
	// use (remove) two operand, and place the result in the stack
	defer c.decStackSize(1)
	c.typeStack.Pop()
	c.typeStack.Pop()
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
	defer c.incStackSize(1)
	typename, _ := e.NodeContent()
	if typename != "this" {
		symbol := c.typeTable.Lookup(typename)
		c.typeStack.Push(DataType{symbol, false})
		c.AppendCode(codeConstant(e))
		return
	}
	fmt.Println(c.symbolTable[c.scopeIndex].name, c.scopeIndex)

	local := c.Lookup("this")
	c.hasField = true
	c.AppendCode("aload_0")
	c.typeStack.Push(local.Member.Type())
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
