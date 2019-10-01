//+build linux

package termutil

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

// RequiresRoot skips tests that require root, unless the test.root flag has
// been set
func RequiresRoot(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("skipping test that requires root")
		return
	}
}

func newTtyForTest(t *testing.T) (*os.File, error) {
	RequiresRoot(t)
	return os.OpenFile("/dev/tty", os.O_RDWR, os.ModeDevice)
}

func newTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "temp")
}

func TestGetWinsize(t *testing.T) {
	cmpWinsize := cmp.AllowUnexported(Winsize{})
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	winSize, err := getWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, winSize != nil)

	newSize := Winsize{Width: 200, Height: 200, x: winSize.x, y: winSize.y}
	err = setWinsize(tty.Fd(), &newSize)
	assert.NilError(t, err)
	winSize, err = getWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.DeepEqual(t, *winSize, newSize, cmpWinsize)
}

func TestSetWinsize(t *testing.T) {
	cmpWinsize := cmp.AllowUnexported(Winsize{})
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	winSize, err := getWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, winSize != nil)
	newSize := Winsize{Width: 200, Height: 200, x: winSize.x, y: winSize.y}
	err = setWinsize(tty.Fd(), &newSize)
	assert.NilError(t, err)
	winSize, err = getWinsize(tty.Fd())
	assert.NilError(t, err)
	assert.DeepEqual(t, *winSize, newSize, cmpWinsize)
}

func TestGetFdInfo(t *testing.T) {
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	term := NewTerminal(tty, tty, tty)
	assert.Equal(t, term.In().Fd(), tty.Fd())
	assert.Equal(t, isTerminal(tty.Fd()), true)
	tmpFile, err := newTempFile()
	assert.NilError(t, err)
	defer handleError(tmpFile.Close())
	term = NewTerminal(tmpFile, tmpFile, nil)
	assert.Equal(t, term.In().Fd(), tmpFile.Fd())
	assert.Equal(t, term.IsTTY(), false)
}

func TestIsTerminal(t *testing.T) {
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	assert.Equal(t, isTerminal(tty.Fd()), true)
	tmpFile, err := newTempFile()
	assert.NilError(t, err)
	defer handleError(tmpFile.Close())
	assert.Equal(t, isTerminal(tmpFile.Fd()), false)
}

func TestSaveState(t *testing.T) {
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	state, err := saveState(tty.Fd())
	assert.NilError(t, err)
	assert.Assert(t, state != nil)
	tty, err = newTtyForTest(t)
	assert.NilError(t, err)
	defer handleError(tty.Close())
	err = restoreTerminal(tty.Fd(), state)
	assert.NilError(t, err)
}

func TestDisableEcho(t *testing.T) {
	tty, err := newTtyForTest(t)
	defer handleError(tty.Close())
	assert.NilError(t, err)
	state, err := setRawTerminalInput(tty.Fd())
	defer handleError(restoreTerminal(tty.Fd(), state))
	assert.NilError(t, err)
	assert.Assert(t, state != nil)
	err = disableEcho(tty.Fd(), state)
	assert.NilError(t, err)
}

func handleError(err error) {
	if err != nil {
		log.Error(err)
	}
}
