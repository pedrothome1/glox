package main

//type astPrinter struct{}
//
//func (x astPrinter) Print(expr Expr) any {
//	result, _ := expr.Accept(&x)
//
//	fmt.Println(result.(string))
//
//	return result
//}
//
//func (x astPrinter) VisitBinaryExpr(expr Binary) (any, error) {
//	return x.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right), nil
//}
//
//func (x astPrinter) VisitGroupingExpr(expr Grouping) (any, error) {
//	return x.parenthesize("group", expr.Expression), nil
//}
//
//func (x astPrinter) VisitLiteralExpr(expr Literal) (any, error) {
//	if expr.Value == nil {
//		return "nil", nil
//	}
//
//	return fmt.Sprintf("%v", expr.Value), nil
//}
//
//func (x astPrinter) VisitUnaryExpr(expr Unary) (any, error) {
//	return x.parenthesize(expr.Operator.Lexeme, expr.Right), nil
//}
//
//func (x astPrinter) parenthesize(name string, exprs ...Expr) string {
//	builder := strings.Builder{}
//
//	builder.WriteString("(")
//	builder.WriteString(name)
//
//	for _, expr := range exprs {
//		value, _ := expr.Accept(x)
//
//		builder.WriteString(" ")
//		builder.WriteString(value.(string))
//	}
//
//	builder.WriteString(")")
//
//	return builder.String()
//}
