package main

import (
	"fmt"
	"strings"
)

type astPrinter struct{}

func (x astPrinter) Print(expr Expr) any {
	return expr.Accept(&x)
}

func (x astPrinter) VisitBinaryExpr(expr Binary) any {
	return x.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (x astPrinter) VisitGroupingExpr(expr Grouping) any {
	return x.parenthesize("group", expr.Expression)
}

func (x astPrinter) VisitLiteralExpr(expr Literal) any {
	if expr.Value == nil {
		return "nil"
	}

	return fmt.Sprintf("%v", expr.Value)
}

func (x astPrinter) VisitUnaryExpr(expr Unary) any {
	return x.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (x astPrinter) parenthesize(name string, exprs ...Expr) string {
	builder := strings.Builder{}

	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.Accept(x).(string))
	}

	builder.WriteString(")")

	fmt.Println(builder.String())

	return builder.String()
}
