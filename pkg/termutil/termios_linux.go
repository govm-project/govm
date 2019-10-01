package termutil

import (
	"golang.org/x/sys/unix"
)

const (
	getTermiosOp = unix.TCGETS
	setTermiosOp = unix.TCSETS
)

// Termios is the Unix API for terminal I/O.
type Termios unix.Termios

// makeRaw put the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
func makeRaw(fd uintptr) (*State, error) {
	termios, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	var oldState State
	oldState.termios = *termios

	termios.Iflag &^= (unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= (unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)
	termios.Cflag &^= (unix.CSIZE | unix.PARENB)
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	if err := setTermios(fd, termios); err != nil {
		return nil, err
	}

	return &oldState, nil
}

func getTermios(fd uintptr) (*Termios, error) {
	tios, err := unix.IoctlGetTermios(int(fd), getTermiosOp)
	return (*Termios)(tios), err
}

func setTermios(fd uintptr, t *Termios) error {
	return unix.IoctlSetTermios(int(fd), setTermiosOp, (*unix.Termios)(t))
}
