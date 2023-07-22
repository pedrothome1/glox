package main

type Expr interface{}

type binary struct {
	left     Expr
	operator Token
	right    Expr
}

type grouping struct {
	expression Expr
}

type literal struct {
	value any
}

type unary struct {
	operator Token
	right    Expr
}
