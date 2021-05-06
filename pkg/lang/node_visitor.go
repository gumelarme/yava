package lang

import (
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

type NodeVisitor struct {
	curScope SymbolTable
	visit    map[string]func(text.INode)
}

func NewNodeVisitor() NodeVisitor {
	node := NodeVisitor{}
	node.curScope = NewSymbolTable("<program>", 0, nil)

	visitorMap := map[string]func(text.INode){
		"class":        node.visitClass,
		"propertyDecl": node.visitProperties,
	}
	node.visit = visitorMap
	return node
}

func (n *NodeVisitor) initBasicTypes() {
	n.curScope.Insert(TypeSymbol{"int"})
	n.curScope.Insert(TypeSymbol{"char"})
	n.curScope.Insert(TypeSymbol{"boolean"})
	n.curScope.Insert(TypeSymbol{"String"})
}

func (n *NodeVisitor) Traverse(program text.Program) {
	n.initBasicTypes()
	for _, decl := range program {
		typeof, _ := decl.Describe()
		n.visit[typeof](decl)
	}
}

func (n *NodeVisitor) checkType(typename string) TypeSymbol {
	dataType := n.curScope.Lookup(typename)
	if dataType == nil || dataType.Category() != Type {
		msg := fmt.Sprintf("Undefined type %s", typename)
		panic(msg)
	}

	return dataType.(TypeSymbol)
}

func (n *NodeVisitor) visitClass(node text.INode) {
	class := node.(*text.Class)
	typesymbol := TypeSymbol{class.Name}
	n.curScope.Insert(typesymbol)

	name := fmt.Sprintf("<%s>", class.Name)
	parent := n.curScope
	n.curScope = NewSymbolTable(name, n.curScope.level+1, &parent)
	for _, prop := range class.Properties {
		n.visitProperties(prop)
	}
}

func (n *NodeVisitor) visitProperties(node text.INode) {
	prop := node.(*text.PropertyDeclaration)
	ty := n.checkType(prop.Type.Name)
	n.curScope.Insert(FieldSymbol{
		prop.Name,
		&ty,
		prop.Type.IsArray,
	})
}
