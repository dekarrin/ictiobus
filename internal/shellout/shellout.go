// Package shellout contains functions for executing shell commands. It mostly
// just wraps the os/exec package, but provides a few convenience functions and
// makes calling it slightly simpler.
package shellout

import (
	"os"
	"os/exec"
)

// Shell is a command execution shell on a system. It maintains a working
// directory and an environment.
type Shell struct {

	// WorkingDir is the current working directory of the shell.
	Dir string

	// Env is the current environment of the shell.
	Env []string
}

// Exec immediately executes the command in the shell's working directory with
// the shell's environment. It returns the combined output of stdout and stderr
// as a string once the command has finished executing.
func (s *Shell) Exec(cmd string, args ...string) (string, error) {
	return Exec(s.Dir, s.Env, cmd, args...)
}

// Exec executes the command in the working directory wd with the given
// environment. It returns the combined output of stdout and stderr as a string
// once the command has finished executing.
func Exec(wd string, env []string, cmd string, args ...string) (string, error) {
	execCmd := exec.Command(cmd, args...)
	execCmd.Env = env
	execCmd.Dir = wd
	out, err := execCmd.CombinedOutput()
	return string(out), err
}

// ExecIn executes the command in the working directory wd with the same
// environment as the current process. It returns the combined output of stdout
// and stderr as a string once the command has finished executing.
func ExecIn(wd string, cmd string, args ...string) (string, error) {
	return Exec(wd, os.Environ(), cmd, args...)
}

// ExecFG executes the command in the working directory wd with the given
// environment and with stdout and stderr forwarded to the calling program's
// stdout and stderr. It returns the error status of the command once it has
// finished executing.
func ExecFG(wd string, env []string, cmd string, args ...string) error {
	execCmd := exec.Command(cmd, args...)
	execCmd.Env = env
	execCmd.Dir = wd
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
