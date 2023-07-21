package main

import "fmt"

func ReportError(line int, message string) {
	fmt.Printf("[line %d] Error: %s", line, message)
}
