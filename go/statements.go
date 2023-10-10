package main

// StmtVisitor for statements
type StmtVisitor interface {
	VisitExpressionStmt(stmt *ExpressionStmt) error
	VisitPrintStmt(stmt *PrintStmt) error
	VisitVarStmt(stmt *VarStmt) error
	VisitBlockStmt(stmt *BlockStmt) error
	VisitIfStmt(stmt *IfStmt) error
	VisitWhileStmt(stmt *WhileStmt) error
	VisitFunctionStmt(stmt *FunctionStmt) error
	VisitReturnStmt(stmt *ReturnStmt) error
	VisitClassStmt(stmt *ClassStmt) error
}

type Stmt interface {
	Accept(visitor StmtVisitor) error
}

type ExpressionStmt struct {
	Expression Expr
}

func (x *ExpressionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitExpressionStmt(x)
}

type PrintStmt struct {
	Expression Expr
}

func (x *PrintStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitPrintStmt(x)
}

type VarStmt struct {
	Name        Token
	Initializer Expr
}

func (x *VarStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitVarStmt(x)
}

type BlockStmt struct {
	Statements []Stmt
}

func (x *BlockStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitBlockStmt(x)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (x *IfStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitIfStmt(x)
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (x *WhileStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitWhileStmt(x)
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

func (x *FunctionStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitFunctionStmt(x)
}

type ReturnStmt struct {
	Keyword Token
	Value   Expr
}

func (x *ReturnStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitReturnStmt(x)
}

type ClassStmt struct {
	Name       Token
	Superclass *Variable
	Methods    []*FunctionStmt
}

func (x *ClassStmt) Accept(visitor StmtVisitor) error {
	return visitor.VisitClassStmt(x)
}
