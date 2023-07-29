package main

// StmtVisitor for statements
type StmtVisitor interface {
	VisitExpressionStmt(stmt ExpressionStmt) (any, error)
	VisitPrintStmt(stmt PrintStmt) (any, error)
}

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type ExpressionStmt struct {
	Expression Expr
}

func (x ExpressionStmt) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(x)
}

type PrintStmt struct {
	Expression Expr
}

func (x PrintStmt) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(x)
}



