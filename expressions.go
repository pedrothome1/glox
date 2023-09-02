package main

// ExprVisitor for expressions
type ExprVisitor interface {
	VisitBinaryExpr(expr Binary) (any, error)
	VisitGroupingExpr(expr Grouping) (any, error)
	VisitLiteralExpr(expr Literal) (any, error)
	VisitUnaryExpr(expr Unary) (any, error)
	VisitVariableExpr(expr Variable) (any, error)
	VisitAssignExpr(expr Assign) (any, error)
	VisitLogicalExpr(expr Logical) (any, error)
	VisitCallExpr(expr Call) (any, error)
}

// Expressions
type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (x Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(x)
}

type Grouping struct {
	Expression Expr
}

func (x Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(x)
}

type Literal struct {
	Value any
}

func (x Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(x)
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (x Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(x)
}

type Variable struct {
	Name Token
}

func (x Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(x)
}

type Assign struct {
	Name  Token
	Value Expr
}

func (x Assign) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignExpr(x)
}

type Logical struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (x Logical) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(x)
}

type Call struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
}

func (x Call) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(x)
}
