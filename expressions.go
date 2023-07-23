package main

// Visitor
type Visitor interface {
	VisitBinaryExpr(expr Binary) (any, error)
	VisitGroupingExpr(expr Grouping) (any, error)
	VisitLiteralExpr(expr Literal) (any, error)
	VisitUnaryExpr(expr Unary) (any, error)
}

// Expressions
type Expr interface {
	Accept(visitor Visitor) (any, error)
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (x Binary) Accept(visitor Visitor) (any, error) {
	return visitor.VisitBinaryExpr(x)
}

type Grouping struct {
	Expression Expr
}

func (x Grouping) Accept(visitor Visitor) (any, error) {
	return visitor.VisitGroupingExpr(x), nil
}

type Literal struct {
	Value any
}

func (x Literal) Accept(visitor Visitor) (any, error) {
	return visitor.VisitLiteralExpr(x), nil
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (x Unary) Accept(visitor Visitor) (any, error) {
	return visitor.VisitUnaryExpr(x), nil
}
