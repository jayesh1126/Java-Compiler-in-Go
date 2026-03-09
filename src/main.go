package main

import (
	"fmt"
	"go-java-compiler/src/compiler"
	"go-java-compiler/src/lexer"
	"go-java-compiler/src/parser"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Fatal error: %v\n", r)
		}
	}()

	if len(os.Args) < 2 {
		fmt.Println("Usage: java-compiler <input.java>")
		return
	}

	inputFile := os.Args[1]

	// Read the java source file
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Printf("Compiling %s...\n", inputFile)

	// Lexical analysis
	l := lexer.New(string(source))
	
	// Count tokens
	tokenCount := 0
	for {
		tok := l.NextToken()
		tokenCount++
		if tok.Type == lexer.EOF {
			break
		}
		if tokenCount <= 10 {
			fmt.Printf("  Token %d: %s = %q\n", tokenCount, tok.Type, tok.Literal)
		}
	}
	fmt.Printf("Total tokens: %d\n", tokenCount)

	// Re-create lexer for parsing
	l = lexer.New(string(source))

	// Parsing
	p := parser.New(l)
	program := p.ParseProgram()
	
	fmt.Printf("Parsed program with %d classes\n", len(program.Classes))

	if program == nil || len(program.Classes) == 0 {
		fmt.Println("Error: Failed to parse Java file - no classes found")
		return
	}

	className := program.Classes[0].Name
	fmt.Printf("Found class: %s\n", className)
	
	// Debug: print method info
	for _, m := range program.Classes[0].Methods {
		fmt.Printf("  Method: %s, isStatic=%v, returnType=%s\n", m.Name, m.IsStatic, m.ReturnType)
	}

	// Compilation to bytecode
	comp := compiler.NewCompiler(program)
	bytecode, err := comp.Compile()
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	// Write the .class file
	outputFile := className + ".class"
	err = os.WriteFile(outputFile, bytecode, 0644)
	if err != nil {
		fmt.Printf("Error writing class file: %v\n", err)
		return
	}

	fmt.Printf("✓ Successfully compiled to %s (%d bytes)\n", outputFile, len(bytecode))
	fmt.Printf("You can run it with: java %s\n", className)
}
