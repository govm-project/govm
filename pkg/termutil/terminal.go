package termutil

import (
	"fmt"
	"os"
)

// ErrNoTTY defines the errormessage to return when a terminal is not a tty
var ErrNoTTY = fmt.Errorf("not a TTY")

// Terminal describes a terminal it stores file pointers to the in, out and
// error file handers and the state of the in and output
type Terminal struct {
	in, out, err      *os.File
	inState, outState *State
}

// NewTerminal returns a pointer to a new Terminal struct using the given in,
// out and err files as files for input, output and error handling.
func NewTerminal(in, out, err *os.File) *Terminal {
	if err == nil {
		err = out
	}

	return &Terminal{
		in:  in,
		out: out,
		err: err,
	}
}

// In returns the pointer to the input of the terminal the method is called on
func (t *Terminal) In() *os.File {
	return t.in
}

// Out returns the pointer to the output of the terminal the method is called on
func (t *Terminal) Out() *os.File {
	return t.out
}

// Err returns the pointer to the error of the terminal the method is called on
func (t *Terminal) Err() *os.File {
	return t.err
}

// IsTTY returns true if the terminal the method is called on is a tty
func (t *Terminal) IsTTY() bool {
	return isTerminal(t.in.Fd())
}

// MakeRaw sets the terminal output to raw
func (t *Terminal) MakeRaw() error {
	if os.Getenv("NORAW") != "" {
		return nil
	}

	if !t.IsTTY() {
		return ErrNoTTY
	}

	is, err := setRawTerminalInput(t.in.Fd())
	if err != nil {
		return err
	}

	t.inState = is

	os, err := setRawTerminalOutput(t.out.Fd())
	if err != nil {
		return err
	}

	t.outState = os

	return nil
}

// Restore restores the terminal by setting the input and output state to nil
func (t *Terminal) Restore() error {
	if !t.IsTTY() {
		return ErrNoTTY
	}

	if t.inState != nil {
		if err := restoreTerminal(t.in.Fd(), t.inState); err != nil {
			return err
		}

		t.inState = nil
	}

	if t.outState != nil {
		if err := restoreTerminal(t.out.Fd(), t.outState); err != nil {
			return err
		}

		t.outState = nil
	}

	return nil
}

// GetState returns the current state of the terminal the method is called on
func (t *Terminal) GetState() (*Termios, error) {
	if !t.IsTTY() {
		return nil, ErrNoTTY
	}

	return getTermios(t.in.Fd())
}

// GetWinsize returns the current window size of the terminal the method is
// called on
func (t *Terminal) GetWinsize() (*Winsize, error) {
	if !t.IsTTY() {
		return nil, ErrNoTTY
	}

	return getWinsize(t.in.Fd())
}

// SetWinsize sets the window size of terminal the method is called on
func (t *Terminal) SetWinsize(ws *Winsize) error {
	if !t.IsTTY() {
		return ErrNoTTY
	}

	return setWinsize(t.in.Fd(), ws)
}

// Close closes the terminal it is called on
func (t *Terminal) Close() error {
	if err := t.out.Close(); err != nil {
		return err
	}

	return t.in.Close()
}
