package main

type Callable interface {
	Arity() int
	String() string
	Call(interpreter *Interpreter, arguments []any) (any, error)
}

// region FunctionImpl
type FunctionImpl struct {
	declaration   *FunctionStmt
	closure       *Environment
	isInitializer bool
}

func (f *FunctionImpl) Arity() int {
	return len(f.declaration.Params)
}

func (f *FunctionImpl) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}

func (f *FunctionImpl) Call(interpreter *Interpreter, arguments []any) (retVal any, err error) {
	defer func() {
		if v, ok := recover().(FunctionReturn); ok {
			if f.isInitializer {
				retVal, _ = f.closure.GetAt(0, "this")
			} else {
				retVal = v.value
			}
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

	if f.isInitializer {
		instance, _ := f.closure.GetAt(0, "this")
		return instance, nil
	}

	return retVal, nil
}

func (f *FunctionImpl) Bind(instance *InstanceImpl) *FunctionImpl {
	environment := &Environment{
		values:    make(map[string]any),
		enclosing: f.closure,
	}
	environment.Define("this", instance)

	return &FunctionImpl{
		declaration:   f.declaration,
		closure:       environment,
		isInitializer: f.isInitializer,
	}
}

type FunctionReturn struct {
	value any
}

// endregion

// region Class
type ClassImpl struct {
	name    string
	methods map[string]*FunctionImpl
}

func (c *ClassImpl) Arity() int {
	initializer := c.FindMethod("init")
	if initializer == nil {
		return 0
	}

	return initializer.Arity()
}

func (c *ClassImpl) Call(interpreter *Interpreter, args []any) (any, error) {
	instance := &InstanceImpl{klass: c}

	initializer := c.FindMethod("init")
	if initializer != nil {
		_, err := initializer.Bind(instance).Call(interpreter, args)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func (c *ClassImpl) String() string {
	return c.name
}

func (c *ClassImpl) FindMethod(name string) *FunctionImpl {
	if method, ok := c.methods[name]; ok {
		return method
	}

	return nil
}

type InstanceImpl struct {
	klass  *ClassImpl
	fields map[string]any
}

func (x *InstanceImpl) Get(name Token) (any, error) {
	if x.fields == nil {
		x.fields = make(map[string]any)
	}

	if value, ok := x.fields[name.Lexeme]; ok {
		return value, nil
	}

	if method := x.klass.FindMethod(name.Lexeme); method != nil {
		return method.Bind(x), nil
	}

	return nil, RuntimeError{"undefined property '" + name.Lexeme + "'", name}
}

func (x *InstanceImpl) Set(name Token, value any) {
	if x.fields == nil {
		x.fields = make(map[string]any)
	}

	x.fields[name.Lexeme] = value
}

func (x *InstanceImpl) String() string {
	return x.klass.name + " instance"
}

// endregion
