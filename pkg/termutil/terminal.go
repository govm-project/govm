package termutil

import (
	"fmt"
	"os"
)

var ErrNoTTY = fmt.Errorf("not a TTY")

type Terminal struct {
	in, out, err      *os.File
	inState, outState *State
}

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

func (t *Terminal) In() *os.File {
	return t.in
}

func (t *Terminal) Out() *os.File {
	return t.out
}

func (t *Terminal) Err() *os.File {
	return t.err
}

func (t *Terminal) IsTTY() bool {
	return isTerminal(t.in.Fd())
}

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

func (t *Terminal) GetState() (*Termios, error) {
	if !t.IsTTY() {
		return nil, ErrNoTTY
	}

	return getTermios(t.in.Fd())
}

func (t *Terminal) GetWinsize() (*Winsize, error) {
	if !t.IsTTY() {
		return nil, ErrNoTTY
	}

	return getWinsize(t.in.Fd())
}

func (t *Terminal) SetWinsize(ws *Winsize) error {
	if !t.IsTTY() {
		return ErrNoTTY
	}

	return setWinsize(t.in.Fd(), ws)
}

func (t *Terminal) Close() error {
	if err := t.out.Close(); err != nil {
		return err
	}
	return t.in.Close()
}
