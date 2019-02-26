// +build !windows

package termutil

import (
	"golang.org/x/sys/unix"
)

// Winsize represents the size of the terminal window.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

// getWinsize returns the window size based on the specified file descriptor.
func getWinsize(fd uintptr) (*Winsize, error) {
	uws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	ws := &Winsize{Height: uws.Row, Width: uws.Col, x: uws.Xpixel, y: uws.Ypixel}
	return ws, err
}

// setWinsize tries to set the specified window size for the specified file descriptor.
func setWinsize(fd uintptr, ws *Winsize) error {
	uws := &unix.Winsize{Row: ws.Height, Col: ws.Width, Xpixel: ws.x, Ypixel: ws.y}
	return unix.IoctlSetWinsize(int(fd), unix.TIOCSWINSZ, uws)
}
