package main

type Scanner interface {
	ScanTokens() ([]Token, error)
}

type Token interface {
	String() string
}

func NewScanner(source string) Scanner {
	return nil
}
