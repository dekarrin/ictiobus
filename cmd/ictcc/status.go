package main

import (
	"fmt"
	"os"

	"github.com/dekarrin/ictiobus/syntaxerr"
)

const (
	warnPrefix  = "WARN: "
	errorPrefix = "ERROR: "
)

const (
	// ExitSuccess is the exit code for a successful run.
	ExitSuccess = iota

	// ExitErrNoFiles is the code returned as exit status when no files are
	// provided to the invocation.
	ExitErrNoFiles

	// ExitErrInvalidFlags is used if the combination of flags specified is
	// invalid.
	ExitErrInvalidFlags

	// ExitErrSyntax is the code returned as exit status when a syntax error
	// occurs.
	ExitErrSyntax

	// ExitErrParser is the code returned as exit status when there is an error
	// generating the parser.
	ExitErrParser

	// ExitErrGeneration is the code returned as exit status when there is an
	// error creating the generated files.
	ExitErrGeneration

	// ExitErrFatalWarn is the code returned as exit status when a warning that
	// is specified to be treated as fatal is encountered.
	ExitErrFatalWarn

	// ExitErrOther is a generic error code for any other error.
	ExitErrOther
)

var (
	exitStatus = ExitSuccess
)

// errNoFiles sets the exit status to ExitErrNoFiles and prints the given error
// message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errNoFiles(msg string) {
	exitErr(ExitErrNoFiles, msg)
}

// errInvalidFlags sets the exit status to ExitErrInvalidFlags and prints the
// given error message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errInvalidFlags(msg string) {
	exitErr(ExitErrInvalidFlags, msg)
}

// errSyntax sets the exit status to ExitErrSyntax and prints an error message
// given by the syntax error to stderr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errSyntax(filename string, synErr *syntaxerr.Error) {
	if filename == "" {
		filename = "<INPUT>"
	}
	fmt.Fprintf(os.Stderr, "%s\n", synErr.MessageForFile(filename))
	exitStatus = ExitErrSyntax
}

// errParser sets the exit status to ExitErrParser and prints the given error
// message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errParser(msg string) {
	exitErr(ExitErrParser, msg)
}

// errGeneration sets the exit status to ExitErrGeneration and prints the given
// error message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errGeneration(msg string) {
	exitErr(ExitErrGeneration, msg)
}

// errFatalWarn sets the exit status to ExitErrFatalWarn and prints the given
// error message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errFatalWarn(msg string) {
	exitErr(ExitErrFatalWarn, msg)
}

// errOther sets the exit status to ExitErrOther and prints the given error
// message to stderr by calling exitErr.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func errOther(msg string) {
	exitErr(ExitErrOther, msg)
}

// exitErr sets the exit status and prints "ERROR: " followed by the given
// error message to stderr. Automatically ends printed message with a newline.
//
// Caller is responsible for exiting main immediately after this function
// returns.
func exitErr(statusCode int, msg string) {
	fmt.Fprintf(os.Stderr, errorPrefix+"%s\n", msg)
	exitStatus = statusCode
}

// basic function to check if panic is happening and recover it while also
// preserving possibly-set exit code. Immediately call this as defered as first
// statement in main.
func preservePanicOrExitWithStatus() {
	if panicErr := recover(); panicErr != nil {
		// we are panicking, make sure we dont lose the panic just because
		// we checked
		panic("unrecoverable panic occured")
	} else {
		os.Exit(exitStatus)
	}
}
