package termutil

import (
	"os"
)

var stdTerm = NewTerminal(os.Stdin, os.Stdout, os.Stderr)

func StdTerminal() *Terminal {
	return stdTerm
}

func In() *os.File {
	return StdTerminal().In()
}

func Out() *os.File {
	return StdTerminal().Out()
}

func Err() *os.File {
	return StdTerminal().Err()
}

func IsTTY() bool {
	return StdTerminal().IsTTY()
}

func MakeRaw() error {
	return StdTerminal().MakeRaw()
}

func Restore() error {
	return StdTerminal().Restore()
}

func GetWinsize() (*Winsize, error) {
	return StdTerminal().GetWinsize()
}

func SetWinsize(ws *Winsize) error {
	return StdTerminal().SetWinsize(ws)
}

func GetState() (*Termios, error) {
	return StdTerminal().GetState()
}
