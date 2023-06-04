/*
FMI, the FISHIMath Interpreter, calculates the value of mathematical expressions
written in the FISHIMath language.

Usage:

	fmi
	fmi FILE ...
	fmi -c FM_CODE [-c FM_CODE ...] [FILE ...]

The fmi command is an example client of an Ictiobus frontend. FISHIMath
statements passed to fmi are converted to an AST, which is then evaluated. The
result is then printed to stdout.

All input is expected to be UTF-8 encoded text containing FISHIMath statements.
If fmi is not passed any arguments, it will enter interactive REPL mode, where
every statement entered into stdin is immediately evaluated and the result
printed to stdout. This will continue until the user enters the special command
"\q" by itself or sends the EOT character with ^D. Errors in code entered during
REPL mode will not cause fmi to exit with non-zero status.

If a filename is passed as an argument to fmi, instead of entering REPL mode,
fmi executes the FISHIMath code in the file and prints the result of the final
statement to stdout. If multiple files are given as arguments, each one is
executed as a separate file and the results of each are printed in the order
they are given. If an error occurs while executing a file, the error is printed
to stderr and fmi moves on to the next file given, if any. One file having an
error in it will not stop the remaining files from being executed, but it will
cause the exit status of fmi to be non-zero.

FISHIMath code may also be given via arguments to one or more -c/--code flags.
If they are, fmi will evaluate it and print the result of the final statement to
stdout. Multiple -c flags can be given. If there are are multiple, they are all
considered part of the same FISHIMath program and their statements will be
evaluated in the order they are given. If any -c has statements that result in
an error, fmi immediately halts execution and returns a non-zero exit code. Code
given via -c flags is executed before any files specified as arguments to fmi
are executed. A statement must be fully within a single -c argument; it cannot
be split "across" -c arguments.

FISHIMath is a simple mathematical expressions language with eccentric tokens
meant to illustrate how to build a custom language. Valid FM code consists of
one or more statements, each ending with the "statement shark" symbol "<o^><".
A statement is either a variable assignment expression or a mathematical sum
that uses the binary arithmetic operators "+", "-", "*", and "/". Grouping is
supported with the "fishtail" symbol ">{" and "fishhead" symbol "'}" which take
the place of the traditional parentheses. Variables are supported and can be
assigned to by using the "tentacle" operator "=o" and placing the variable on
the left-hand side and the expression to assign to it on the right side.
Variables may hold any type of value supported by FM. If a variable is read
before it is assigned, it will contain the integer 0. Both float and int type
constants and variables are supported, but variables do not have explicit
typing. Internally, fmi represents floating point numbers as IEEE-754 single
precision numbers (it uses the Go `float32` type).

Flags:

	-c, --code FM_CODE
		Execute the FISHIMath statements in FM_EXPR. Can be given multiple times
		and if so only the result of the last statement across all -c args is
		shown. Code in all -c flags will execute before any files specified.
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dekarrin/ictfishimath_ast/fm"
	"github.com/spf13/pflag"
)

var (
	flagCode = pflag.StringArrayP("code", "c", nil, "Execute the given FM statements.")
)

const (
	ExitSuccess = iota
	ExitErr
)

var exitStatus int

func main() {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// we are panicking, make sure we dont lose the panic just because
			// we checked
			panic("unrecoverable panic occured")
		} else {
			os.Exit(exitStatus)
		}
	}()

	pflag.Parse()

	// sanity check cli args
	for i, codeArg := range *flagCode {
		if strings.TrimSpace(codeArg) == "" {
			handleError(fmt.Errorf("-c/--code argument #%d is blank", i+1))
		}
	}

	inputFiles := pflag.Args()

	replMode := len(inputFiles) < 1 && len(*flagCode) < 1

	// get the FISHIMath interpreter
	fmi := fm.Interpreter{}

	// if entering repl mode, do that
	if replMode {
		err := readEvalPrintLoop(fmi)
		if err != nil {
			handleError(err)
		}
		return
	}

	// otherwise, read all the -c commands in order first
	if len(*flagCode) > 0 {
		for i := range *flagCode {
			codeArg := (*flagCode)[i]

			err := fmi.Eval(codeArg)
			if err != nil {
				handleError(err)
				return
			}
		}

		fmt.Printf("%s\n", fmi.LastResult.String())
	}

	// next, if there are any files, read those in as well
	for i := range inputFiles {
		fmi.Clear()

		fName := inputFiles[i]
		file, err := os.Open(fName)
		if err != nil {
			handleError(err)

			// deliberately not exiting here so we go through all files
			continue
		}
		defer file.Close()

		readFrom := bufio.NewReader(file)
		fmi.File = fName
		err = fmi.EvalReader(readFrom)
		if err != nil {
			handleError(err)

			// deliberately not exiting here so we go through all files
			continue
		}

		fmt.Printf("%s\n", fmi.LastResult.String())
	}
}

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR:\n%v\n", err)
	exitStatus = ExitErr
}

func readEvalPrintLoop(fmi fm.Interpreter) error {
	fmt.Printf("FISHIMath REPL\n")
	fmt.Printf("(end lines with <o^>< to evaluate; type \\q to quit)\n")
	fmt.Printf("---------------------------------------------------\n")

	running := true
	stdin := bufio.NewReader(os.Stdin)
	var linesToSend string
	var prompt = "==> "
	for running {
		fmt.Print(prompt)
		input, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}

		inputTrimmed := strings.TrimSpace(input)

		if inputTrimmed == "\\q" {
			running = false
			continue
		}

		// adding full input, not inputTrimmed to linesToSend
		linesToSend += input

		// if we didn't end with a statement shark <o^>< nothing else to do.
		if !strings.HasSuffix(inputTrimmed, "<o^><") {
			prompt = "shark?> "
			continue
		}

		// if we got here, user ended the input with a statement shark; send it
		// to the interpreter:

		err = fmi.Eval(linesToSend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		} else {
			fmt.Printf("%s\n", fmi.LastResult)
		}
		linesToSend = ""
		prompt = "==> "
	}

	return nil
}
