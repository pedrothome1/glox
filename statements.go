package main

// StmtVisitor for statements
type StmtVisitor interface {
	VisitExpressionStmt(stmt ExpressionStmt) error
	VisitPrintStmt(stmt PrintStmt) error
}

type Stmt interface {
	Accept(visitor StmtVisitor) error
}

type ExpressionStmt struct {
	Expression Expr
}

func (x ExpressionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitExpressionStmt(x)
}

type PrintStmt struct {
	Expression Expr
}

func (x PrintStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitPrintStmt(x)
}
