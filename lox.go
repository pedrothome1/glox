package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	/* AST Printer */
	//expression := &Binary{
	//	Left: &Unary{
	//		Operator: Token{Minus, "-", nil, 1},
	//		Right:    &Literal{123},
	//	},
	//	Operator: Token{Star, "*", nil, 1},
	//	Right:    &Grouping{&Literal{45.67}},
	//}
	//
	//fmt.Println(astPrinter{}.Print(expression))

	/* Original code */
	if len(os.Args) > 2 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		err := runFile(os.Args[1])
		if err != nil {
			// TODO: Report error
			os.Exit(65)
		}
	} else {
		runPrompt()
	}
}

func runFile(fileName string) error {
	bytes, err := os.ReadFile(fileName)
	panicIfError(err)

	return run(string(bytes))
}

func runPrompt() {
	stdin := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !stdin.Scan() {
			break
		}

		panicIfError(stdin.Err())

		err := run(stdin.Text())
		if err != nil {
			fmt.Printf("%v", err)
		}
	}
}

func run(source string) error {
	scanner := NewScanner(source)

	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}

	parser := NewParser(tokens)

	expression, err := parser.Parse()
	if err != nil {
		return err
	}

	astPrinter{}.Print(expression)

	return nil
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
