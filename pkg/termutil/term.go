// +build !windows

// Package termutil provides structures and helper functions to work with
// terminal (state, sizes). Taken from docker-ce source code.
package termutil

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

var (
	// ErrInvalidState is returned if the state of the terminal is invalid.
	ErrInvalidState = errors.New("invalid terminal state")
)

// State represents the state of the terminal.
type State struct {
	termios Termios
}

// IsTerminal returns true if the given file descriptor is a terminal.
func isTerminal(fd uintptr) bool {
	var termios Termios
	return tcget(fd, &termios) == 0
}

// RestoreTerminal restores the terminal connected to the given file descriptor
// to a previous state.
func restoreTerminal(fd uintptr, state *State) error {
	if state == nil {
		return ErrInvalidState
	}

	if err := tcset(fd, &state.termios); err != 0 {
		return err
	}

	return nil
}

// SaveState saves the state of the terminal connected to the given file descriptor.
func saveState(fd uintptr) (*State, error) {
	var oldState State
	if err := tcget(fd, &oldState.termios); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

// DisableEcho applies the specified state to the terminal connected to the file
// descriptor, with echo disabled.
func disableEcho(fd uintptr, state *State) error {
	newState := state.termios
	newState.Lflag &^= unix.ECHO

	if err := tcset(fd, &newState); err != 0 {
		return err
	}

	handleInterrupt(fd, state)

	return nil
}

// setRawTerminalInput puts the terminal connected to the given file descriptor
// into raw mode and returns the previous state. On UNIX, this puts both the
// input and output into raw mode. On Windows, it only puts the input into raw
// mode.
func setRawTerminalInput(fd uintptr) (*State, error) {
	oldState, err := makeRaw(fd)
	if err != nil {
		return nil, err
	}
	// handleInterrupt(fd, oldState)
	return oldState, err
}

// setRawTerminalOutput puts the output of terminal connected to the given file
// descriptor into raw mode. On UNIX, this does nothing and returns nil for the
// state. On Windows, it disables LF -> CRLF translation.
func setRawTerminalOutput(fd uintptr) (*State, error) {
	return nil, nil
}

func handleInterrupt(fd uintptr, state *State) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	go func() {
		for range sigchan {
			// quit cleanly and the new terminal item is on a new line
			fmt.Println()
			signal.Stop(sigchan)
			close(sigchan)
			_ = restoreTerminal(fd, state)
			// os.Exit(1)
		}
	}()
}
