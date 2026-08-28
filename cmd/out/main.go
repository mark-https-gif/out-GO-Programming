package main

import (
	"fmt"
	"os"

	"github.com/out-lang/out/internal/env"
	"github.com/out-lang/out/internal/eval"
	"github.com/out-lang/out/internal/lexer"
	"github.com/out-lang/out/internal/object"
	"github.com/out-lang/out/internal/parser"
)

const VERSION = "v0.4.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("OUT " + VERSION)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		case "run":
			if len(os.Args) < 3 {
				fmt.Println("Usage: out run <filename.out>")
				os.Exit(1)
			}
			runFile(os.Args[2])
		default:
			runFile(os.Args[1])
		}
	} else {
		runRepl()
	}
}

func printHelp() {
	fmt.Println(`OUT Language ` + VERSION + `

Usage:
  out                     Start REPL
  out run <file.out>      Run a file
  out <file.out>          Run a file (shorthand)
  out --version           Show version
  out --help              Show this help

Example:
  out run hello.out`)
}

func runFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", msg)
		}
		os.Exit(1)
	}

	e := env.New()
	result := eval.Eval(program, e)

	if result != nil {
		if errObj, ok := result.(*object.Error); ok {
			fmt.Fprintf(os.Stderr, "Runtime error: %s\n", errObj.Message)
			os.Exit(1)
		}
	}
}

func runRepl() {
	fmt.Println("OUT Language " + VERSION)
	fmt.Println("Type 'exit' or Ctrl+D to quit")
	fmt.Println()

	e := env.New()

	for {
		fmt.Print(">> ")
		var input string
		n, err := fmt.Scanln(&input)
		if n == 0 || err != nil {
			fmt.Println()
			break
		}
		if input == "exit" || input == "quit" {
			break
		}

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				fmt.Printf("  %s\n", msg)
			}
			continue
		}

		result := eval.Eval(program, e)
		if result != nil {
			resultStr := result.Inspect()
			if resultStr != "" {
				fmt.Printf("= %s\n", resultStr)
			}
		}
	}
}
