package main

import "time"

type funClock struct{}

func (f *funClock) Arity() int {
	return 0
}

func (f *funClock) Call(_ *Interpreter, _ []any) (any, error) {
	return time.Now().Unix(), nil
}

func (f *funClock) String() string {
	return "<native fn>"
}
