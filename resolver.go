package main

type resolver struct {
	interpreter *Interpreter
	scopes      mapStack
}

// region statements
func (r *resolver) VisitExpressionStmt(stmt ExpressionStmt) error {
	err := r.resolveExpr(stmt.Expression)
	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) VisitPrintStmt(stmt PrintStmt) error {
	err := r.resolveExpr(stmt.Expression)
	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) VisitVarStmt(stmt VarStmt) error {
	r.declare(stmt.Name)

	if stmt.Initializer != nil {
		err := r.resolveExpr(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	r.define(stmt.Name)

	return nil
}

func (r *resolver) VisitBlockStmt(stmt BlockStmt) error {
	r.beginScope()

	err := r.resolveStmts(stmt.Statements)
	if err != nil {
		return err
	}

	r.endScope()

	return nil
}

func (r *resolver) VisitIfStmt(stmt IfStmt) error {
	err := r.resolveExpr(stmt.Condition)
	if err != nil {
		return err
	}

	err = r.resolveStmt(stmt.ThenBranch)
	if err != nil {
		return err
	}

	if stmt.ElseBranch != nil {
		err = r.resolveStmt(stmt.ElseBranch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *resolver) VisitWhileStmt(stmt WhileStmt) error {
	err := r.resolveExpr(stmt.Condition)
	if err != nil {
		return err
	}

	err = r.resolveStmt(stmt.Body)
	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) VisitFunctionStmt(stmt FunctionStmt) error {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	err := r.resolveFunction(stmt)
	if err != nil {
		return err
	}

	return nil
}

func (r *resolver) VisitReturnStmt(stmt ReturnStmt) error {
	if stmt.Value != nil {
		err := r.resolveExpr(stmt.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// endregion

// region expressions
func (r *resolver) VisitBinaryExpr(expr Binary) (any, error) {
	err := r.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	err = r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *resolver) VisitGroupingExpr(expr Grouping) (any, error) {
	err := r.resolveExpr(expr.Expression)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *resolver) VisitLiteralExpr(expr Literal) (any, error) {
	return nil, nil
}

func (r *resolver) VisitUnaryExpr(expr Unary) (any, error) {
	err := r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *resolver) VisitVariableExpr(expr Variable) (any, error) {
	if len(r.scopes) > 0 && r.scopes.Peek()[expr.Name.Lexeme] == false {
		// TODO: understand better and refactor
		reportParserError(expr.Name, "can't read local variable in its own initializer")
		return nil, nil
	}

	return nil, r.resolveLocal(expr, expr.Name)
}

func (r *resolver) VisitAssignExpr(expr Assign) (any, error) {
	err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	err = r.resolveLocal(expr, expr.Name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *resolver) VisitLogicalExpr(expr Logical) (any, error) {
	err := r.resolveExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	err = r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *resolver) VisitCallExpr(expr Call) (any, error) {
	err := r.resolveExpr(expr.Callee)
	if err != nil {
		return nil, err
	}

	for _, arg := range expr.Arguments {
		err = r.resolveExpr(arg)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// endregion

// region helpers
func (r *resolver) resolveStmts(statements []Stmt) error {
	for _, statement := range statements {
		err := r.resolveStmt(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *resolver) resolveStmt(statement Stmt) error {
	return statement.Accept(r)
}

func (r *resolver) resolveExpr(expression Expr) error {
	_, err := expression.Accept(r)
	return err
}

func (r *resolver) resolveLocal(expression Expr, name Token) error {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			err := r.interpreter.Resolve(expression, len(r.scopes)-1-i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *resolver) beginScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *resolver) endScope() {
	r.scopes.Pop()
}

func (r *resolver) declare(name Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes.Peek()
	scope[name.Lexeme] = false
}

func (r *resolver) define(name Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes.Peek()
	scope[name.Lexeme] = true
}

func (r *resolver) resolveFunction(fn FunctionStmt) error {
	r.beginScope()

	for _, param := range fn.Params {
		r.declare(param)
		r.define(param)
	}

	err := r.resolveStmts(fn.Body)
	if err != nil {
		return err
	}

	r.endScope()

	return nil
}

// endregion

// region helper data structures
type mapStack []map[string]bool

func (s *mapStack) Push(m map[string]bool) {
	*s = append(*s, m)
}

func (s *mapStack) Pop() map[string]bool {
	if len(*s) > 0 {
		v := (*s)[len(*s)-1]
		*s = (*s)[:len(*s)-1]
		return v
	}

	return nil
}

func (s *mapStack) Peek() map[string]bool {
	if len(*s) > 0 {
		v := (*s)[len(*s)-1]
		return v
	}

	return nil
}

// endregion
