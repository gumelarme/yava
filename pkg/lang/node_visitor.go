package lang

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

type NodeVisitor struct {
	curScope SymbolTable
	visitMap map[string]func(text.INode)
	errors   []error
}

func NewNodeVisitor() *NodeVisitor {
	node := NodeVisitor{}
	node.curScope = NewSymbolTable("<program>", 0, nil)
	node.errors = make([]error, 0)

	visitMap := map[string]func(text.INode){
		"class":           node.visitClass,
		"property-decl":   node.visitProperties,
		"method-decl":     node.visitMethodDeclaration,
		"var-decl":        node.visitVariableDeclaration,
		"stmt-block":      node.visitStatementList,
		"switch":          node.visitSwitchStatement,
		"if":              node.visitIfStatement,
		"for":             node.visitForStatement,
		"while":           node.visitWhileStatement,
		"assignment":      node.visitAssignmentStatement,
		"return":          node.visitJumpStatement,
		"continue":        node.visitJumpStatement,
		"break":           node.visitJumpStatement,
		"field":           node.visitFieldAccess,
		"method-call":     node.visitMethodCall,
		"array-creation":  node.visitArrayCreation,
		"object-creation": node.visitObjectCreation,
		"binop":           node.visitBinOp,
		"num":             node.visitConstant,
		"char":            node.visitConstant,
		"boolean":         node.visitConstant,
		"string":          node.visitConstant,
		"null":            node.visitConstant,
	}

	node.visitMap = visitMap
	return &node
}

func (n *NodeVisitor) initBasicTypes() {
	n.curScope.Insert(TypeSymbol{"int"})
	n.curScope.Insert(TypeSymbol{"char"})
	n.curScope.Insert(TypeSymbol{"boolean"})
	n.curScope.Insert(TypeSymbol{"String"})
}

func (n *NodeVisitor) Errors() []error {
	return n.errors
}

func (n *NodeVisitor) Traverse(program text.Program) {
	n.initBasicTypes()
	for _, decl := range program {
		n.visit(decl)
	}
}

func (n *NodeVisitor) checkType(typename string) TypeSymbol {
	dataType := n.curScope.Lookup(typename)
	if dataType == nil || dataType.Category() != Type {
		n.addError(fmt.Errorf("Undefined type %s", typename))
		return TypeSymbol{}
	}

	return dataType.(TypeSymbol)
}

func (n *NodeVisitor) addError(err error) {
	n.errors = append(n.errors, err)
	fmt.Println("Adding error: ", len(n.errors))
}
func (n *NodeVisitor) visit(node text.INode) {
	name, _ := node.NodeContent()
	n.visitMap[name](node)
}

func (n *NodeVisitor) visitClass(node text.INode) {
	class := node.(*text.Class)
	if decl := n.curScope.Lookup(class.Name); decl != nil {
		n.addError(fmt.Errorf("Type %s is already declared.", decl.Name()))
	}

	typesymbol := TypeSymbol{class.Name}
	n.curScope.Insert(typesymbol)

	name := fmt.Sprintf("[Class: %s]", class.Name)
	parent := n.curScope
	n.curScope = NewSymbolTable(name, n.curScope.level+1, &parent)

	n.curScope.Insert(FieldSymbol{
		DataType{
			&typesymbol,
			false,
		},
		"this",
	})

	for _, prop := range class.Properties {
		n.visitProperties(prop)
	}

	for _, method := range class.Methods {
		n.visitMethodDeclaration(method)
	}
}

func (n *NodeVisitor) visitProperties(node text.INode) {
	prop := node.(*text.PropertyDeclaration)
	ty := n.checkType(prop.Type.Name)
	n.curScope.Insert(FieldSymbol{
		DataType{
			&ty,
			prop.Type.IsArray,
		},
		prop.Name,
	})
}

//TODO: check interface implementations?
func (n *NodeVisitor) visitMethodDeclaration(node text.INode) {
	methodDecl := node.(*text.MethodDeclaration)
	typeof := n.checkType(methodDecl.ReturnType.Name)
	method := n.curScope.Lookup(methodDecl.Name)
	if method == nil {
		symbol := NewMethodSymbol(*methodDecl, typeof)
		symbol.AddSignature(methodDecl.MethodSignature)
		n.curScope.Insert(symbol)
	} else {
		err := n.curScope.InsertOverloadMethod(method.Name(), methodDecl.MethodSignature)
		if err != nil {
			n.errors = append(n.errors, err)
		}
	}

	n.visitStatementList(methodDecl.Body)
}

func (n *NodeVisitor) visitStatementList(node text.INode) {
	parent := n.curScope
	nextLevel := parent.level + 1
	name := fmt.Sprintf("block-%d", nextLevel)
	n.curScope = NewSymbolTable(name, nextLevel, &parent)

	list := node.(text.StatementList)
	for _, stmt := range list {
		n.visit(stmt)
	}
}

func (n *NodeVisitor) visitVariableDeclaration(node text.INode) {
	varDecl := node.(*text.VariableDeclaration)
	typeof := n.checkType(varDecl.Type.Name)
	n.curScope.Insert(FieldSymbol{DataType{&typeof, varDecl.Type.IsArray}, varDecl.Name})

	if value := varDecl.ChildNode(); value != nil {
		n.visit(value)
	}
}

func (n *NodeVisitor) visitBinOp(node text.INode) {
	binOp := node.(*text.BinOp)
	n.visit(binOp.Left)
	n.visit(binOp.Right)
}

func (n *NodeVisitor) visitFieldAccess(node text.INode) {
	fieldLookup := node.(*text.FieldAccess)
	symbol := n.curScope.Lookup(fieldLookup.Name)
	if symbol == nil || symbol.Category() != Field {
		n.addError(fmt.Errorf("Undefined field %s", fieldLookup.Name))
		return
	}

	//TODO: check type and members
	if fieldLookup.Child != nil {
		n.visitNamedValueChild(fieldLookup.Child, symbol)
	}
}

func (n *NodeVisitor) visitMethodCall(node text.INode) {
	methodCall := node.(*text.MethodCall)
	symbol := n.curScope.Lookup(methodCall.Name)
	if symbol == nil || symbol.Category() != Method {
		n.addError(fmt.Errorf("Undefined method %s", methodCall.Name))
		return
	}

	//TODO: check method signature
	//method := symbol.(*MethodSymbol)
	for _, arg := range methodCall.Args {
		n.visit(arg)
	}

	if methodCall.Child != nil {
		n.visitNamedValueChild(methodCall.Child, symbol)
	}
}

func (n *NodeVisitor) visitNamedValueChild(node text.INode, symbol Symbol) {
	name, _ := node.NodeContent()
	if name == "array" {
		if !symbol.Type().isArray {
			n.addError(fmt.Errorf("%s %s is not an array", symbol.Category(), symbol.Name()))
			return
		}
		n.visitArrayAccess(node, symbol)
		return
	}
	n.visit(node)
}

func (n *NodeVisitor) visitArrayAccess(node text.INode, symbol Symbol) {
	array := node.(*text.ArrayAccess)
	n.visit(array.At)
	if array.ChildNode() != nil {
		n.visitNamedValueChild(array.ChildNode(), symbol)
	}
}

func (n *NodeVisitor) visitConstant(node text.INode) {
	_, content := node.NodeContent()
	fmt.Printf("Constant: %s\n", content)
}

func (n *NodeVisitor) visitArrayCreation(node text.INode) {
	array := node.(*text.ArrayCreation)
	n.checkType(array.Type)
	n.visit(array.Length)
}

func (n *NodeVisitor) visitObjectCreation(node text.INode) {
	object := node.(*text.ObjectCreation)
	n.checkType(object.Name)
	for _, arg := range object.Args {
		n.visit(arg)
	}

	if object.ChildNode() != nil {
		n.visit(object.ChildNode())
	}
}

func (n *NodeVisitor) visitJumpStatement(node text.INode) {
	jump := node.(*text.JumpStatement)
	if jump.Exp != nil {
		n.visit(jump.Exp)
	}
}

func (n *NodeVisitor) visitAssignmentStatement(node text.INode) {
	assign := node.(*text.AssignmentStatement)
	n.visit(assign.Left)
	n.visit(assign.Right)
}

func (n *NodeVisitor) visitSwitchStatement(node text.INode) {
	switchStmt := node.(*text.SwitchStatement)
	n.visit(switchStmt.ValueToCompare)
	for _, c := range switchStmt.CaseList {
		n.visit(c.Value)
		for _, stmt := range c.StatementList {
			n.visit(stmt)
		}
	}

	for _, d := range switchStmt.DefaultCase {
		n.visit(d.ChildNode())
	}
}

func (n *NodeVisitor) visitIfStatement(node text.INode) {
	ifStmt := node.(*text.IfStatement)
	n.visit(ifStmt.Condition)
	n.visit(ifStmt.Body)
	if ifStmt.Else != nil {
		n.visit(ifStmt.Else)
	}
}

func (n *NodeVisitor) visitForStatement(node text.INode) {
	forStmt := node.(*text.ForStatement)
	if forStmt.Init != nil {
		n.visit(forStmt.Init)
	}
	if forStmt.Condition != nil {
		n.visit(forStmt.Condition)
	}
	if forStmt.Update != nil {
		n.visit(forStmt.Update)
	}
	n.visit(forStmt.Body)

}
func (n *NodeVisitor) visitWhileStatement(node text.INode) {
	while := node.(*text.WhileStatement)
	n.visit(while.Condition)
	n.visit(while.Body)
}
