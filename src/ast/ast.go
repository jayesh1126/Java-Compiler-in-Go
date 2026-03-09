package ast

import "go-java-compiler/src/lexer"

// Node is implemented by all AST nodes.
type Node interface {
	node()
}

// Expression is implemented by all expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Statement is implemented by all statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Program is the root node of every AST the parser produces.
type Program struct {
	Classes []*ClassDecl
}

func (p *Program) node() {}

// ClassDecl represents a class declaration.
type ClassDecl struct {
	Name    string
	Methods []*MethodDecl
}

func (c *ClassDecl) node() {}

// MethodDecl represents a method declaration.
type MethodDecl struct {
	Name       string
	ReturnType string // "void", "int", "String", etc.
	Params     []*ParamDecl
	Body       []Statement
	IsPublic   bool
	IsStatic   bool
}

func (m *MethodDecl) node() {}

// ParamDecl represents a method parameter.
type ParamDecl struct {
	Name string
	Type string
}

// VarDecl represents a variable declaration.
type VarDecl struct {
	Name  string
	Type  string
	Value Expression
}

func (v *VarDecl) node()        {}
func (v *VarDecl) statementNode() {}

// ExpressionStatement wraps an expression as a statement.
type ExpressionStatement struct {
	Expr Expression
}

func (e *ExpressionStatement) node()        {}
func (e *ExpressionStatement) statementNode() {}

// BlockStatement represents { ... }
type BlockStatement struct {
	Statements []Statement
}

func (b *BlockStatement) node()        {}
func (b *BlockStatement) statementNode() {}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Value Expression
}

func (r *ReturnStatement) node()        {}
func (r *ReturnStatement) statementNode() {}

// BinaryExpr represents a binary operation (e.g., a + b).
type BinaryExpr struct {
	Left  Expression
	Op    lexer.TokenType
	Right Expression
}

func (b *BinaryExpr) node()           {}
func (b *BinaryExpr) expressionNode() {}

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Value int64
}

func (i *IntegerLiteral) node()           {}
func (i *IntegerLiteral) expressionNode() {}

// StringLiteral represents a string literal.
type StringLiteral struct {
	Value string
}

func (s *StringLiteral) node()           {}
func (s *StringLiteral) expressionNode() {}

// Identifier represents an identifier (variable name).
type Identifier struct {
	Name string
}

func (i *Identifier) node()           {}
func (i *Identifier) expressionNode() {}

// MethodCall represents a method call (e.g., System.out.println(...)).
type MethodCall struct {
	Object    Expression   // can be identifier or another method call
	Method    string
	Arguments []Expression
}

func (m *MethodCall) node()           {}
func (m *MethodCall) expressionNode() {}
