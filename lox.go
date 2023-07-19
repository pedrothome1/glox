package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
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
	var line string

	for {
		fmt.Print("> ")

		_, err := fmt.Scanln(&line)
		if err == io.EOF || line == "" {
			break
		}

		panicIfError(err)

		err = run(line)
		if err != nil {
			// TODO: Report error and move on
		}
	}
}

func run(source string) error {
	scanner := NewScanner(source)

	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}

	for _, token := range tokens {
		fmt.Println(token)
	}

	return nil
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
