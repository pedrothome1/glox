package main

// Visitor
type Visitor interface {
	VisitBinaryExpr(expr Binary) any
	VisitGroupingExpr(expr Grouping) any
	VisitLiteralExpr(expr Literal) any
	VisitUnaryExpr(expr Unary) any
}

type visitorImpl struct{}

func (x visitorImpl) VisitBinaryExpr(expr Binary) any {
	panic("implement me")
}

func (x visitorImpl) VisitGroupingExpr(expr Grouping) any {
	panic("implement me")
}

func (x visitorImpl) VisitLiteralExpr(expr Literal) any {
	panic("implement me")
}

func (x visitorImpl) VisitUnaryExpr(expr Unary) any {
	panic("implement me")
}

// Expressions
type Expr interface {
	Accept(visitor Visitor) any
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (x Binary) Accept(visitor Visitor) any {
	return visitor.VisitBinaryExpr(x)
}

type Grouping struct {
	Expression Expr
}

func (x Grouping) Accept(visitor Visitor) any {
	return visitor.VisitGroupingExpr(x)
}

type Literal struct {
	Value any
}

func (x Literal) Accept(visitor Visitor) any {
	return visitor.VisitLiteralExpr(x)
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (x Unary) Accept(visitor Visitor) any {
	return visitor.VisitUnaryExpr(x)
}
