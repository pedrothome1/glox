package main

import (
	"errors"
	"fmt"
	"strconv"
)

type Interpreter struct {
	globals     *Environment
	environment *Environment
	locals      map[Expr]int
}

func (x *Interpreter) Init() *Interpreter {
	x.globals = &Environment{
		values: map[string]any{},
	}

	// native functions
	x.globals.Define("clock", &funClock{})

	x.environment = x.globals
	x.locals = make(map[Expr]int)

	return x
}

func (x *Interpreter) Interpret(statements []Stmt) error {
	var err error

	for _, stmt := range statements {
		if err = x.execute(stmt); err != nil {
			var rErr RuntimeError
			if errors.As(err, &rErr) {
				runtimeError(rErr)
			}

			return err
		}
	}

	return err
}

func (x *Interpreter) Resolve(expr Expr, depth int) {
	x.locals[expr] = depth
}

// region Expression visitor methods
func (x *Interpreter) VisitBinaryExpr(expr *Binary) (any, error) {
	left, err := x.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := x.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Greater:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) > right.(float64), nil
	case GreaterEqual:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) >= right.(float64), nil
	case Less:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) < right.(float64), nil
	case LessEqual:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) <= right.(float64), nil
	case Minus:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) - right.(float64), nil
	case Plus:
		if leftVal, ok := left.(float64); ok {
			if rightVal, ok := right.(float64); ok {
				return leftVal + rightVal, nil
			}
		}

		if leftVal, ok := left.(string); ok {
			if rightVal, ok := right.(string); ok {
				return leftVal + rightVal, nil
			}
		}

		return nil, RuntimeError{"operands must be two numbers or two strings", expr.Operator}
	case Slash:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) / right.(float64), nil
	case Star:
		if err := x.checkNumberOperands(expr.Operator, left, right); err != nil {
			return nil, err
		}

		return left.(float64) * right.(float64), nil
	case BangEqual:
		return !x.isEqual(left, right), nil
	case EqualEqual:
		return x.isEqual(left, right), nil
	}

	return nil, nil
}

func (x *Interpreter) VisitGroupingExpr(expr *Grouping) (any, error) {
	return x.evaluate(expr.Expression)
}

func (x *Interpreter) VisitLiteralExpr(expr *Literal) (any, error) {
	return expr.Value, nil
}

func (x *Interpreter) VisitUnaryExpr(expr *Unary) (any, error) {
	right, err := x.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case Bang:
		return !x.isTruthy(right), nil
	case Minus:
		if err := x.checkNumberOperand(expr.Operator, right); err != nil {
			return nil, err
		}

		return -(right.(float64)), nil
	}

	return nil, nil
}

func (x *Interpreter) VisitVariableExpr(expr *Variable) (any, error) {
	return x.lookupVariable(expr.Name, expr)
}

func (x *Interpreter) VisitAssignExpr(expr *Assign) (any, error) {
	value, err := x.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	if distance, ok := x.locals[expr]; ok {
		x.environment.AssignAt(distance, expr.Name, value)
	} else {
		err = x.globals.Assign(expr.Name, value)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func (x *Interpreter) VisitLogicalExpr(expr *Logical) (any, error) {
	left, err := x.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type == Or {
		if x.isTruthy(left) {
			return left, nil
		}
	} else {
		if !x.isTruthy(left) {
			return left, nil
		}
	}

	return x.evaluate(expr.Right)
}

func (x *Interpreter) VisitCallExpr(expr *Call) (any, error) {
	callee, err := x.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}

	var arguments []any
	for _, arg := range expr.Arguments {
		argument, err := x.evaluate(arg)
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, argument)
	}

	fn, ok := callee.(Callable)
	if !ok {
		return nil, RuntimeError{
			Message: "can only call functions and classes",
			Token:   expr.Paren,
		}
	}

	return fn.Call(x, arguments)
}

func (x *Interpreter) VisitGetExpr(expr *Get) (any, error) {
	object, err := x.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	if instance, ok := object.(*InstanceImpl); ok {
		return instance.Get(expr.Name)
	}

	return nil, RuntimeError{"only instances have properties", expr.Name}
}

func (x *Interpreter) VisitSetExpr(expr *Set) (any, error) {
	object, err := x.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	var instance *InstanceImpl
	if v, ok := object.(*InstanceImpl); !ok {
		return nil, RuntimeError{"only instances have fields", expr.Name}
	} else {
		instance = v
	}

	value, err := x.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	instance.Set(expr.Name, value)

	return value, nil
}

func (x *Interpreter) VisitThisExpr(expr *ThisExpr) (any, error) {
	return x.lookupVariable(expr.Keyword, expr)
}

func (x *Interpreter) VisitSuperExpr(expr *SuperExpr) (any, error) {
	distance := x.locals[expr]

	super, _ := x.environment.GetAt(distance, "super")
	superclass := super.(*ClassImpl)

	obj, _ := x.environment.GetAt(distance-1, "this")
	object := obj.(*InstanceImpl)

	method := superclass.FindMethod(expr.Method.Lexeme)

	if method == nil {
		return nil, RuntimeError{"undefined property '" + expr.Method.Lexeme + "'", expr.Method}
	}

	return method.Bind(object), nil
}

// endregion

// region Statement visitor methods
func (x *Interpreter) VisitExpressionStmt(stmt *ExpressionStmt) error {
	_, err := x.evaluate(stmt.Expression)

	return err
}

func (x *Interpreter) VisitPrintStmt(stmt *PrintStmt) error {
	value, err := x.evaluate(stmt.Expression)
	if err != nil {
		return err
	}

	fmt.Println(x.stringify(value))

	return nil
}

func (x *Interpreter) VisitVarStmt(stmt *VarStmt) error {
	var value any
	var err error

	if stmt.Initializer != nil {
		value, err = x.evaluate(stmt.Initializer)
		if err != nil {
			return err
		}
	}

	x.environment.Define(stmt.Name.Lexeme, value)

	return nil
}

func (x *Interpreter) VisitBlockStmt(stmt *BlockStmt) error {
	return x.executeBlock(stmt.Statements, &Environment{enclosing: x.environment})
}

func (x *Interpreter) VisitIfStmt(stmt *IfStmt) error {
	condition, err := x.evaluate(stmt.Condition)
	if err != nil {
		return err
	}

	if x.isTruthy(condition) {
		err = x.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		err = x.execute(stmt.ElseBranch)
	}
	if err != nil {
		return err
	}

	return nil
}

func (x *Interpreter) VisitWhileStmt(stmt *WhileStmt) error {
	for {
		condition, err := x.evaluate(stmt.Condition)
		if err != nil {
			return err
		}

		if !x.isTruthy(condition) {
			break
		}

		err = x.execute(stmt.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *Interpreter) VisitFunctionStmt(stmt *FunctionStmt) error {
	fn := &FunctionImpl{stmt, x.environment, false}

	x.environment.Define(stmt.Name.Lexeme, fn)

	return nil
}

func (x *Interpreter) VisitReturnStmt(stmt *ReturnStmt) error {
	var value any
	var err error

	if stmt.Value != nil {
		value, err = x.evaluate(stmt.Value)
		if err != nil {
			return err
		}
	}

	panic(FunctionReturn{value})
}

func (x *Interpreter) VisitClassStmt(stmt *ClassStmt) error {
	var superclassImpl *ClassImpl

	if stmt.Superclass != nil {
		superclass, err := x.evaluate(stmt.Superclass)
		if err != nil {
			return err
		}

		var ok bool
		if superclassImpl, ok = superclass.(*ClassImpl); !ok {
			return RuntimeError{"superclass must be a class", stmt.Superclass.Name}
		}
	}

	x.environment.Define(stmt.Name.Lexeme, nil)

	if stmt.Superclass != nil {
		x.environment = &Environment{
			values:    make(map[string]any),
			enclosing: x.environment,
		}
		x.environment.Define("super", superclassImpl)
	}

	methods := make(map[string]*FunctionImpl)
	for _, method := range stmt.Methods {
		function := &FunctionImpl{
			declaration:   method,
			closure:       x.environment,
			isInitializer: method.Name.Lexeme == "init",
		}
		methods[method.Name.Lexeme] = function
	}

	klass := &ClassImpl{name: stmt.Name.Lexeme, methods: methods, superclass: superclassImpl}

	if superclassImpl != nil {
		x.environment = x.environment.enclosing
	}

	err := x.environment.Assign(stmt.Name, klass)
	if err != nil {
		return err
	}

	return nil
}

// endregion

// region private helpers
func (x *Interpreter) executeBlock(statements []Stmt, environment *Environment) error {
	previous := x.environment

	defer func() {
		x.environment = previous
	}()

	x.environment = environment

	for _, statement := range statements {
		err := x.execute(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *Interpreter) evaluate(expr Expr) (any, error) {
	return expr.Accept(x)
}

func (x *Interpreter) execute(stmt Stmt) error {
	return stmt.Accept(x)
}

func (x *Interpreter) lookupVariable(name Token, expr Expr) (any, error) {
	if distance, ok := x.locals[expr]; ok {
		value, _ := x.environment.GetAt(distance, name.Lexeme)

		return value, nil
	}

	return x.globals.Get(name)
}

func (x *Interpreter) stringify(value any) string {
	if value == nil {
		return "nil"
	}

	if v, ok := value.(float64); ok {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}

	if v, ok := value.(fmt.Stringer); ok {
		return v.String()
	}

	return fmt.Sprintf("%v", value)
}

func (x *Interpreter) isEqual(a any, b any) bool {
	return a == b
}

func (x *Interpreter) isTruthy(value any) bool {
	if value == nil {
		return false
	}

	if val, ok := value.(bool); ok {
		return val
	}

	return true
}

func (x *Interpreter) checkNumberOperand(operator Token, operand any) error {
	if _, ok := operand.(float64); ok {
		return nil
	}

	return RuntimeError{"operand must be a number", operator}
}

func (x *Interpreter) checkNumberOperands(operator Token, left any, right any) error {
	if _, ok := left.(float64); ok {
		if _, ok := right.(float64); ok {
			return nil
		}
	}

	return RuntimeError{"operands must be numbers", operator}
}

// endregion

// region Errors
func runtimeError(err RuntimeError) {
	fmt.Printf("%s\n[line %d]\n", err.Message, err.Token.Line)
}

type RuntimeError struct {
	Message string
	Token   Token
}

func (x RuntimeError) Error() string {
	return x.Message
}

// endregion
