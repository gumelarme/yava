package lang

import (
	"errors"
	"fmt"

	"github.com/gumelarme/yava/pkg/text"
)

type ErrorCollector []string

func (e ErrorCollector) Errors() []error {
	errs := make([]error, len(e))
	for i, msg := range e {
		errs[i] = errors.New(msg)
	}
	return errs
}

func (e *ErrorCollector) AddError(err string) {
	(*e) = append(*e, err)
}
func (e *ErrorCollector) AddErrorf(err string, i ...interface{}) {
	(*e) = append(*e, fmt.Sprintf(err, i...))
}

type SymbolCategory int

const (
	Type SymbolCategory = iota
	Field
	Method
)

func (s SymbolCategory) String() string {
	return []string{
		"Type",
		"Field",
		"Method",
	}[s]
}

type Symbol interface {
	Name() string
	Category() SymbolCategory
	String() string
}

type TypeMember interface {
	Symbol
	Type() DataType
}
type TypeCategory int

const (
	Primitive TypeCategory = iota
	Class
	Interface
)

type TypeSymbol struct {
	name         string
	extends      *TypeSymbol
	implements   *TypeSymbol
	Properties   map[string]*PropertySymbol
	Methods      map[string]*MethodSymbol
	TypeCategory TypeCategory
}

func NewType(name string, category TypeCategory) *TypeSymbol {
	return &TypeSymbol{name, nil, nil,
		make(map[string]*PropertySymbol),
		make(map[string]*MethodSymbol),
		category,
	}
}

func (t TypeSymbol) Name() string {
	return t.name
}

func (t TypeSymbol) Category() SymbolCategory {
	return Type
}

func (t TypeSymbol) String() string {
	return fmt.Sprintf("<%s>", t.name)
}

func (t *TypeSymbol) LookupProperty(name string) *PropertySymbol {
	//FIXME: Java get the parent first, but here we get the child prop first
	prop := t.Properties[name]
	if prop != nil {
		return prop
	}

	if t.extends != nil {
		return t.extends.LookupProperty(name)
	}
	return nil
}

func (t *TypeSymbol) LookupMethod(signature string) *MethodSymbol {
	method := t.Methods[signature]
	if method != nil {
		return method
	}

	if t.extends != nil {
		return t.extends.LookupMethod(signature)
	}
	return nil
}

type DataType struct {
	dataType *TypeSymbol
	isArray  bool
}

func (d DataType) String() string {
	str := d.dataType.name
	if d.isArray {
		str += "[]"
	}
	return str
}

type FieldSymbol struct {
	DataType
	name string
}

func (f FieldSymbol) Name() string {
	return f.name
}

func (f FieldSymbol) Type() DataType {
	return f.DataType
}

func (f FieldSymbol) Category() SymbolCategory {
	return Field
}

func (f FieldSymbol) String() string {
	return fmt.Sprintf("%s: %s", f.dataType, f.name)
}

type PropertySymbol struct {
	text.AccessModifier
	FieldSymbol
}

func (p PropertySymbol) String() string {
	return p.FieldSymbol.String()
}

type MethodSymbol struct {
	DataType
	accessMod text.AccessModifier
	name      string
	parameter []*FieldSymbol
}

func NewMethodSymbol(signature text.MethodSignature, returnType TypeSymbol) *MethodSymbol {
	return &MethodSymbol{
		DataType{
			&returnType,
			signature.ReturnType.IsArray,
		},
		signature.AccessModifier,
		signature.Name,
		nil,
	}
}

func (m MethodSymbol) Name() string {
	return m.name
}

func (m MethodSymbol) Type() DataType {
	return m.DataType
}

func (m MethodSymbol) Category() SymbolCategory {
	return Method
}

func (m MethodSymbol) String() string {
	return fmt.Sprintf("%s %s()", m.DataType, m.name)
}

type SymbolTable struct {
	name      string
	level     int
	table     map[string]Symbol
	parent    *SymbolTable
	isVerbose bool
}

func NewSymbolTable(name string, level int, parent *SymbolTable) SymbolTable {
	return SymbolTable{
		name,
		level,
		make(map[string]Symbol),
		parent,
		true,
	}
}

func (s *SymbolTable) Insert(sym Symbol) {
	s.table[sym.Name()] = sym

	if s.isVerbose {
		fmt.Printf("Insert %s @%s\n", sym.Name(), s.name)
	}
}

func (s *SymbolTable) Lookup(name string) Symbol {
	if s.isVerbose {
		fmt.Printf("Lookup %s @%s\n", name, s.name)
	}

	if val, found := s.table[name]; found {
		return val
	}

	if s.parent != nil {
		return s.parent.Lookup(name)
	}

	return nil
}
