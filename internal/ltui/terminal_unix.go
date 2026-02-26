//go:build darwin || linux || freebsd

package ltui

import (
	"os"

	"golang.org/x/sys/unix"
)

func makeRaw(fd *os.File) (*unix.Termios, error) {
	oldState, err := unix.IoctlGetTermios(int(fd.Fd()), unix.TIOCGETA)
	if err != nil {
		return nil, err
	}

	newState := *oldState
	newState.Lflag &^= unix.ECHO | unix.ICANON

	if err := unix.IoctlSetTermios(int(fd.Fd()), unix.TIOCSETA, &newState); err != nil {
		return nil, err
	}

	return oldState, nil
}

func restoreTerminal(fd *os.File, state *unix.Termios) error {
	if state == nil {
		return nil
	}
	return unix.IoctlSetTermios(int(fd.Fd()), unix.TIOCSETA, state)
}
