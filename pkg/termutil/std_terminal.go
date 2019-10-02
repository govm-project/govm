package termutil

import (
	"os"
)

// StdTerminal returns a pointer to a new terminal
func StdTerminal() *Terminal {
	return NewTerminal(os.Stdin, os.Stdout, os.Stderr)
}

// In returns a file pointer to stdin of a new terminal
func In() *os.File {
	return StdTerminal().In()
}

// Out returns a file pointer to stdout of a new terminal
func Out() *os.File {
	return StdTerminal().Out()
}

// Err returns a file pointer to stderr of a new terminal
func Err() *os.File {
	return StdTerminal().Err()
}

// IsTTY returns true if the new terminal is a tty
func IsTTY() bool {
	return StdTerminal().IsTTY()
}

// MakeRaw creates a new terminal with raw output
func MakeRaw() error {
	return StdTerminal().MakeRaw()
}

// Restore returns a new terminal with no in or out state
func Restore() error {
	return StdTerminal().Restore()
}

// GetWinsize returns a new terminals window size
func GetWinsize() (*Winsize, error) {
	return StdTerminal().GetWinsize()
}

// SetWinsize sets the window size of a new terminal
func SetWinsize(ws *Winsize) error {
	return StdTerminal().SetWinsize(ws)
}

// GetState returns the state of a new terminal
func GetState() (*Termios, error) {
	return StdTerminal().GetState()
}
