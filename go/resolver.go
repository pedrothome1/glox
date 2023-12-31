package main

type classType int

const (
	classTypeNone = iota
	classTypeClass
	classTypeSubclass
)

type functionType int

const (
	funcTypeNone functionType = iota
	funcTypeFunction
	funcTypeMethod
	funcTypeInitializer
)

type Resolver struct {
	interpreter     *Interpreter
	scopes          mapStack
	currentFunction functionType
	currentClass    classType
}

func NewResolver(interpreter *Interpreter) *Resolver {
	return &Resolver{
		interpreter: interpreter,
		scopes:      mapStack{},
	}
}

func (r *Resolver) Resolve(statements []Stmt) error {
	return r.resolveStmts(statements)
}

// region statements
func (r *Resolver) VisitExpressionStmt(stmt *ExpressionStmt) error {
	err := r.resolveExpr(stmt.Expression)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *PrintStmt) error {
	err := r.resolveExpr(stmt.Expression)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitVarStmt(stmt *VarStmt) error {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}

	if stmt.Initializer != nil {
		err = r.resolveExpr(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	r.define(stmt.Name)

	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *BlockStmt) error {
	r.beginScope()

	err := r.resolveStmts(stmt.Statements)
	if err != nil {
		return err
	}

	r.endScope()

	return nil
}

func (r *Resolver) VisitIfStmt(stmt *IfStmt) error {
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

func (r *Resolver) VisitWhileStmt(stmt *WhileStmt) error {
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

func (r *Resolver) VisitFunctionStmt(stmt *FunctionStmt) error {
	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}

	r.define(stmt.Name)

	err = r.resolveFunction(stmt, funcTypeFunction)
	if err != nil {
		return err
	}

	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ReturnStmt) error {
	if r.currentFunction == funcTypeNone {
		return TokenError(stmt.Keyword, "can't return from top-level code")
	}

	if stmt.Value != nil {
		if r.currentFunction == funcTypeInitializer {
			return TokenError(stmt.Keyword, "can't return a value from an initializer")
		}

		err := r.resolveExpr(stmt.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) VisitClassStmt(stmt *ClassStmt) error {
	enclosingClass := r.currentClass
	r.currentClass = classTypeClass

	err := r.declare(stmt.Name)
	if err != nil {
		return err
	}

	r.define(stmt.Name)

	if stmt.Superclass != nil && stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
		return TokenError(stmt.Superclass.Name, "a class can't inherit from itself")
	}

	if stmt.Superclass != nil {
		r.currentClass = classTypeSubclass

		err = r.resolveExpr(stmt.Superclass)
		if err != nil {
			return err
		}
	}

	if stmt.Superclass != nil {
		r.beginScope()
		r.scopes.Peek()["super"] = true
	}

	r.beginScope()
	r.scopes.Peek()["this"] = true

	for _, method := range stmt.Methods {
		declaration := funcTypeMethod
		if method.Name.Lexeme == "init" {
			declaration = funcTypeInitializer
		}

		err = r.resolveFunction(method, declaration)
		if err != nil {
			return err
		}
	}

	r.endScope()

	if stmt.Superclass != nil {
		r.endScope()
	}

	r.currentClass = enclosingClass

	return nil
}

// endregion

// region expressions
func (r *Resolver) VisitBinaryExpr(expr *Binary) (any, error) {
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

func (r *Resolver) VisitGroupingExpr(expr *Grouping) (any, error) {
	err := r.resolveExpr(expr.Expression)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Resolver) VisitLiteralExpr(_ *Literal) (any, error) {
	return nil, nil
}

func (r *Resolver) VisitUnaryExpr(expr *Unary) (any, error) {
	err := r.resolveExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Resolver) VisitVariableExpr(expr *Variable) (any, error) {
	if len(r.scopes) > 0 {
		if value, ok := r.scopes.Peek()[expr.Name.Lexeme]; ok && value == false {
			return nil, TokenError(expr.Name, "can't read local variable in its own initializer")
		}
	}

	return nil, r.resolveLocal(expr, expr.Name)
}

func (r *Resolver) VisitAssignExpr(expr *Assign) (any, error) {
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

func (r *Resolver) VisitLogicalExpr(expr *Logical) (any, error) {
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

func (r *Resolver) VisitCallExpr(expr *Call) (any, error) {
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

func (r *Resolver) VisitGetExpr(expr *Get) (any, error) {
	err := r.resolveExpr(expr.Object)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Resolver) VisitSetExpr(expr *Set) (any, error) {
	err := r.resolveExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	err = r.resolveExpr(expr.Object)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Resolver) VisitThisExpr(expr *ThisExpr) (any, error) {
	if r.currentClass == classTypeNone {
		return nil, TokenError(expr.Keyword, "can't use 'this' outside of a class")
	}

	err := r.resolveLocal(expr, expr.Keyword)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *Resolver) VisitSuperExpr(expr *SuperExpr) (any, error) {
	if r.currentClass == classTypeNone {
		return nil, TokenError(expr.Keyword, "can't use 'super' outside of a class")
	} else if r.currentClass != classTypeSubclass {
		return nil, TokenError(expr.Keyword, "can't use 'super' in a class with no superclass")
	}

	return nil, r.resolveLocal(expr, expr.Keyword)
}

// endregion

// region helpers
func (r *Resolver) resolveStmts(statements []Stmt) error {
	for _, statement := range statements {
		err := r.resolveStmt(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) resolveStmt(statement Stmt) error {
	return statement.Accept(r)
}

func (r *Resolver) resolveExpr(expression Expr) error {
	_, err := expression.Accept(r)
	return err
}

func (r *Resolver) resolveLocal(expression Expr, name Token) error {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			r.interpreter.Resolve(expression, len(r.scopes)-1-i)
		}
	}
	return nil
}

func (r *Resolver) beginScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

func (r *Resolver) declare(name Token) error {
	if len(r.scopes) == 0 {
		return nil
	}

	scope := r.scopes.Peek()

	if _, ok := scope[name.Lexeme]; ok {
		return TokenError(name, "already a variable with this name in this scope")
	}

	scope[name.Lexeme] = false

	return nil
}

func (r *Resolver) define(name Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes.Peek()
	scope[name.Lexeme] = true
}

func (r *Resolver) resolveFunction(fn *FunctionStmt, funcType functionType) error {
	enclosingFunction := r.currentFunction
	r.currentFunction = funcType

	r.beginScope()

	for _, param := range fn.Params {
		err := r.declare(param)
		if err != nil {
			return err
		}

		r.define(param)
	}

	err := r.resolveStmts(fn.Body)
	if err != nil {
		return err
	}

	r.endScope()

	r.currentFunction = enclosingFunction

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
