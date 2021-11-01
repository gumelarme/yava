package lang

import (
	"errors"
	"fmt"
	"strings"

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

func (t *TypeSymbol) isDescendantOf(val *TypeSymbol) bool {
	for parent := t.extends; parent != nil; parent = parent.extends {
		if val == parent {
			return true
		}
	}
	return false
}

func (t *TypeSymbol) isImplementing(val *TypeSymbol) bool {
	parent := t
	for {
		if parent.implements == val {
			return true
		}

		if parent.extends != nil {
			parent = parent.extends
			continue
		}

		break
	}
	return false
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

func (t *TypeSymbol) LookupMethodByArgs(name string, args []DataType) *MethodSymbol {
	sameNameMethods := t.getMethodsByName(name)

	//filter by number of args
	argLen := len(args)
	var sameNameSameArgCount []*MethodSymbol
	for _, method := range sameNameMethods {
		if argLen == len(method.args) {
			sameNameSameArgCount = append(sameNameSameArgCount, method)
		}
	}

	for _, method := range sameNameSameArgCount {
		if method.CanAccept(args) {
			return method
		}
	}
	return nil
}

func (t *TypeSymbol) getMethodsByName(name string) []*MethodSymbol {
	var methods []*MethodSymbol
	for _, m := range t.Methods {
		if m.name == name {
			methods = append(methods, m)
		}
	}

	if t.extends != nil {
		parentMethods := t.extends.getMethodsByName(name)
		methods = append(methods, parentMethods...)
	}

	return methods
}

func IsPrimitive(dt DataType) bool {
	if dt.isArray {
		return false
	}

	switch dt.Name() {
	case "int", "boolean", "char":
		return true
	default:
		return false
	}
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
func (d DataType) Name() string {
	return d.dataType.name
}

func (d DataType) Equals(val DataType) bool {
	return d.isArray == val.isArray && d.dataType.name == val.dataType.name
}

type FieldSymbol struct {
	DataType
	name string
}

func (f *FieldSymbol) Name() string {
	return f.name
}

func (f *FieldSymbol) Type() DataType {
	return f.DataType
}

func (f *FieldSymbol) Category() SymbolCategory {
	return Field
}

func (f *FieldSymbol) String() string {
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
	args      []DataType
}

func NewMethodSymbol(signature text.MethodSignature, returnType TypeSymbol) *MethodSymbol {
	return &MethodSymbol{
		DataType{
			&returnType,
			signature.ReturnType.IsArray,
		},
		signature.AccessModifier,
		signature.Name,
		make([]DataType, 0),
	}
}

func (m *MethodSymbol) Name() string {
	return m.String()
}

func (m *MethodSymbol) Type() DataType {
	return m.DataType
}

func (m *MethodSymbol) Category() SymbolCategory {
	return Method
}

func (m *MethodSymbol) CanAccept(args []DataType) bool {
	if len(m.args) != len(args) {
		return false
	}

	for i, mArg := range m.args {
		expect := args[i]
		if expect == mArg {
			continue //	accepted
		}

		if expect.dataType.isDescendantOf(mArg.dataType) &&
			expect.isArray == mArg.isArray {
			continue //	accepted
		}

		if expect.dataType.isImplementing(mArg.dataType) &&
			expect.isArray == mArg.isArray {
			continue //	accepted
		}
		return false
	}
	return true
}

func (m *MethodSymbol) String() string {
	argString := make([]string, len(m.args))
	for i, a := range m.args {
		argString[i] = a.String()
	}
	// return fmt.Sprintf("%s(%s)", m.dataType)
	return fmt.Sprintf("%s(%s)", m.name, strings.Join(argString, ", "))
}

// FIXME: Change to meaningful name
type Local struct {
	Member  TypeMember
	address int
}

type SymbolTable struct {
	name      string
	level     int
	table     map[string]Local
	parent    *SymbolTable
	isVerbose bool
}

func NewSymbolTable(name string, level int, parent *SymbolTable) SymbolTable {
	return SymbolTable{
		name,
		level,
		make(map[string]Local),
		parent,
		false,
	}
}

func (s *SymbolTable) Insert(sym TypeMember, address int) {
	s.table[sym.Name()] = Local{sym, address}

	if s.isVerbose {
		fmt.Printf("Insert %s @%s\n @%d", sym.Name(), s.name, address)
	}
}

func (s *SymbolTable) Lookup(name string, deep bool) (TypeMember, int) {
	if s.isVerbose {
		fmt.Printf("Lookup %s @%s\n", name, s.name)
	}

	if val, found := s.table[name]; found {
		return val.Member, val.address
	}

	if s.parent != nil && deep {
		return s.parent.Lookup(name, deep)
	}

	return nil, -1
}

func (s *SymbolTable) LookupMethod(name string, args []DataType, deep bool) (*MethodSymbol, int) {
	if s.isVerbose {
		fmt.Printf("Lookup %s @%s\n", name, s.name)
	}

	for _, local := range s.table {
		if local.Member.Category() == Method && local.Member.Name() == name {
			m := local.Member.(*MethodSymbol)
			if m.CanAccept(args) {
				return m, local.address
			}
		}
	}

	if s.parent != nil && deep {
		return s.parent.LookupMethod(name, args, deep)
	}

	return nil, -1
}
