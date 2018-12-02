// Package aster is golang coding efficiency engine.
//
// Copyright 2018 henrylee2cn. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package aster

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// Module packages AST
type Module struct {
	FileSet *token.FileSet
	Dir     string
	filter  func(os.FileInfo) bool
	Pkgs    map[string]*Package // <package name, *Package>
	mode    parser.Mode
}

// A Package node represents a set of source files
// collectively building a Go package.
//
type Package struct {
	module  *Module // nil when not existed
	FileSet *token.FileSet
	Dir     string
	Name    string                 // package name
	Scope   *ast.Scope             // package scope across all files
	Imports map[string]*ast.Object // map of package id -> package object
	Files   map[string]*File       // Go source files by filename
	mode    parser.Mode
}

// A File node represents a Go source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.
//
// For correct printing of source code containing comments (using packages
// go/format and go/printer), special care must be taken to update comments
// when a File's syntax tree is modified: For printing, comments are interspersed
// between tokens based on their position. If syntax tree nodes are
// removed or moved, relevant comments in their vicinity must also be removed
// (from the File.Comments list) or moved accordingly (by updating their
// positions). A CommentMap may be used to facilitate some of these operations.
//
// Whether and how a comment is associated with a node depends on the
// interpretation of the syntax tree by the manipulating program: Except for Doc
// and Comment comments directly associated with nodes, the remaining comments
// are "free-floating" (see also issues #18593, #20744).
//
type File struct {
	*ast.File
	pkg      *Package // nil when not existed
	PkgName  string
	FileSet  *token.FileSet
	Filename string
	Src      []byte
	mode     parser.Mode
	Imports  []*Import
	Blocks   map[token.Pos]Block // <type node pos, Block>
}

// Import import info
type Import struct {
	*ast.ImportSpec
	Name string
	Path string
	Doc  *ast.CommentGroup
}

type (
	// Block the basic sub-interface based on ast.Node extension,
	// is the supertype of other extended interfaces.
	Block interface {
		CommBlockMethods
		FuncBlockMethods
		TypeBlockMethods
		blockIdentify() // only as identify method
	}
	// FuncBlock is the representation of a Go function or method.
	// NOTE: Kind = Func
	FuncBlock interface {
		CommBlockMethods
		FuncBlockMethods
		funcBlockIdentify() // only as identify method
	}
	// TypeBlock is the representation of a Go type node.
	// NOTE: Kind != Func
	TypeBlock interface {
		CommBlockMethods
		TypeBlockMethods
		typeBlockIdentify() // only as identify method
	}
)

type (
	// CommBlockMethods is the common methods of block interface.
	CommBlockMethods interface {
		// Node returns origin AST node.
		Node() ast.Node

		// Name returns the type's name within its package for a defined type.
		// For other (non-defined) types it returns the empty string.
		Name() string

		// Kind returns the specific kind of this type.
		Kind() Kind

		// Doc returns lead comment.
		Doc() string
	}

	// TypeBlockMethods is the representation of a Go type node.
	// NOTE: Kind != Func
	TypeBlockMethods interface {
		// IsAssign is there `=` for declared type?
		IsAssign() bool

		// NumMethod returns the number of exported methods in the type's method set.
		NumMethod() int

		// Method returns the i'th method in the type's method set.
		// For a non-interface type T or *T, the returned Method's Type and Func
		// fields describe a function whose first argument is the receiver.
		//
		// For an interface type, the returned Method's Type field gives the
		// method signature, without a receiver, and the Func field is nil.
		Method(int) (FuncBlock, bool)

		// MethodByName returns the method with that name in the type's
		// method set and a boolean indicating if the method was found.
		//
		// For a non-interface type T or *T, the returned Method's Type and Func
		// fields describe a function whose first argument is the receiver.
		//
		// For an interface type, the returned Method's Type field gives the
		// method signature, without a receiver, and the Func field is nil.
		MethodByName(string) (FuncBlock, bool)

		// Implements reports whether the type implements the interface type u.
		Implements(u TypeBlock) bool

		// addMethod adds a FuncBlock as method.
		//
		// Returns error if the FuncBlock is already exist or receiver is not the TypeBlock.
		addMethod(FuncBlock) error

		// -------------- Only for Kind=Struct ---------------

		// NumField returns a struct type's field count.
		// It panics if the type's Kind is not Struct.
		NumField() int

		// Field returns a struct type's i'th field.
		Field(i int) (field *StructField, found bool)

		// FieldByName returns the struct field with the given name
		// and a boolean indicating if the field was found.
		FieldByName(name string) (field *StructField, found bool)
	}

	// FuncBlockMethods is the representation of a Go function or method.
	// NOTE: Kind = Func
	FuncBlockMethods interface {
		// NumParam returns a function type's input parameter count.
		NumParam() int

		// NumResult returns a function type's output parameter count.
		NumResult() int

		// Param returns the type of a function type's i'th input parameter.
		Param(int) (*FuncField, bool)

		// Result returns the type of a function type's i'th output parameter.
		Result(int) (*FuncField, bool)

		// IsVariadic reports whether a function type's final input parameter
		// is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
		// implicit actual type []T.
		//
		// For concreteness, if t represents func(x int, y ... float64), then
		//
		//	f.NumParam() == 2
		//	f.Param(0) is the Type for "int"
		//	f.Param(1) is the Type for "[]float64"
		//	f.IsVariadic() == true
		//
		IsVariadic() bool

		// Recv returns receiver (methods); or returns false (functions)
		Recv() (*FuncField, bool)
	}
)

// FuncField function params or results.
type FuncField struct {
	Name     string
	TypeName string // not contain `*`
}

// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
type Kind uint

// Kind enumerate
const (
	Invalid Kind = iota
	Suspense
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	Interface
	Chan
	Array
	Slice
	Map
	Func
	Struct
	// Ptr
)

func getBasicKind(basicName string) (k Kind, found bool) {
	found = true
	switch basicName {
	case "bool":
		k = Bool
	case "int":
		k = Int
	case "int8":
		k = Int8
	case "int16":
		k = Int16
	case "int32":
		k = Int32
	case "int64":
		k = Int64
	case "uint":
		k = Uint
	case "uint8":
		k = Uint8
	case "uint16":
		k = Uint16
	case "uint32":
		k = Uint32
	case "uint64":
		k = Uint64
	case "uintptr":
		k = Uintptr
	case "float32":
		k = Float32
	case "float64":
		k = Float64
	case "complex64":
		k = Complex64
	case "complex128":
		k = Complex128
	case "string":
		k = String
	default:
		return Invalid, false
	}
	return
}

// NilNode nil Node
type NilNode struct{}

// Pos .
func (NilNode) Pos() token.Pos { return token.NoPos }

// End .
func (NilNode) End() token.Pos { return token.NoPos }

// super common node extension info
type super struct {
	file    *File
	kind    Kind
	namePtr *string
	doc     *ast.CommentGroup
}

func (f *File) newSuper(namePtr *string, kind Kind, doc *ast.CommentGroup) *super {
	return &super{
		file:    f,
		kind:    kind,
		namePtr: namePtr,
		doc:     doc,
	}
}

func (s *super) blockIdentify() {}

// Kind returns the facade kind of this node.
func (s *super) Kind() Kind {
	return s.kind
}

// Name returns the type's name within its package for a defined type.
// For other (non-defined) types it returns the empty string.
func (s *super) Name() string {
	if s.namePtr == nil {
		return ""
	}
	return *s.namePtr
}

// Doc returns lead comment.
func (s *super) Doc() string {
	if s.doc == nil {
		return ""
	}
	return s.doc.Text()
}

// ------------------------ Kind: Func ------------------------

// NumParam returns a function type's input parameter count.
func (s *super) NumParam() int {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// NumResult returns a function type's output parameter count.
func (s *super) NumResult() int {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// Param returns the type of a function type's i'th input parameter.
func (s *super) Param(int) (*FuncField, bool) {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// Result returns the type of a function type's i'th output parameter.
func (s *super) Result(int) (*FuncField, bool) {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// IsVariadic reports whether a function type's final input parameter
// is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
// implicit actual type []T.
//
// For concreteness, if t represents func(x int, y ... float64), then
//
//	f.NumParam() == 2
//	f.Param(0) is the Type for "int"
//	f.Param(1) is the Type for "[]float64"
//	f.IsVariadic() == true
//
func (s *super) IsVariadic() bool {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// Recv returns receiver (methods); or returns false (functions)
func (s *super) Recv() (*FuncField, bool) {
	if s.kind != Func {
		panic("Kind must be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// IsFuncBlock returns true if b is implementd FuncBlock.
func IsFuncBlock(b Block) bool {
	_, ok := b.(FuncBlock)
	return ok
}

// IsTypeBlock returns true if b is implementd TypeBlock.
func IsTypeBlock(b Block) bool {
	_, ok := b.(TypeBlock)
	return ok
}

// ------------------------ Type ------------------------

// IsAssign is there `=` for declared type?
func (s *super) IsAssign() bool {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// NumMethod returns the number of exported methods in the type's method set.
func (s *super) NumMethod() int {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// Method returns the i'th method in the type's method set.
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (s *super) Method(int) (FuncBlock, bool) {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// MethodByName returns the method with that name in the type's
// method set and a boolean indicating if the method was found.
//
// For a non-interface type T or *T, the returned Method's Type and Func
// fields describe a function whose first argument is the receiver.
//
// For an interface type, the returned Method's Type field gives the
// method signature, without a receiver, and the Func field is nil.
func (s *super) MethodByName(string) (FuncBlock, bool) {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// Implements reports whether the type implements the interface type u.
func (s *super) Implements(u TypeBlock) bool {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// addMethod adds a FuncBlock as method.
//
// Returns error if the FuncBlock is already exist or receiver is not the TypeBlock.
func (s *super) addMethod(FuncBlock) error {
	if s.kind == Func {
		panic("Kind cant not be aster.Func!")
	}
	panic("TODO: Coming soon!")
}

// -------------- Only for Kind=Struct ---------------

// NumField returns a struct type's field count.
func (s *super) NumField() int {
	if s.kind != Struct {
		panic("Kind must be aster.Struct!")
	}
	panic("TODO: Coming soon!")
}

// Field returns a struct type's i'th field.
func (s *super) Field(i int) (field *StructField, found bool) {
	if s.kind != Struct {
		panic("Kind must be aster.Struct!")
	}
	panic("TODO: Coming soon!")
}

// FieldByName returns the struct field with the given name
// and a boolean indicating if the field was found.
func (s *super) FieldByName(name string) (field *StructField, found bool) {
	if s.kind != Struct {
		panic("Kind must be aster.Struct!")
	}
	panic("TODO: Coming soon!")
}
