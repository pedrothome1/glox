package main

type Callable interface {
	Arity() int
	String() string
	Call(interpreter *Interpreter, arguments []any) (any, error)
}

type function struct {
	declaration *FunctionStmt
	closure     *Environment
}

func (f *function) Arity() int {
	return len(f.declaration.Params)
}

func (f *function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}

func (f *function) Call(interpreter *Interpreter, arguments []any) (retVal any, err error) {
	defer func() {
		if v, ok := recover().(functionReturn); ok {
			retVal = v.value
		}
	}()

	environment := &Environment{
		values:    map[string]any{},
		enclosing: f.closure,
	}

	for i := 0; i < len(f.declaration.Params); i++ {
		environment.Define(f.declaration.Params[i].Lexeme, arguments[i])
	}

	err = interpreter.executeBlock(f.declaration.Body, environment)
	if err != nil {
		return nil, err
	}

	return retVal, nil
}

type functionReturn struct {
	value any
}
