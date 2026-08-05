//go:build !windows

package main

import (
	"syscall"
	"unsafe"
)

type rawState = *syscall.Termios

// isTerminal reports whether fd refers to a terminal.
func isTerminal(fd uintptr) (bool, error) {
	var state syscall.Termios
	if err := ioctl(fd, tcgetscmd, &state); err != nil {
		if err == syscall.ENOTTY {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// makeRaw puts the terminal into raw mode and returns the previous state.
func makeRaw(fd uintptr) (rawState, error) {
	var old syscall.Termios
	if err := ioctl(fd, tcgetscmd, &old); err != nil {
		return nil, err
	}

	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON |
		syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := ioctl(fd, tcsetscmd, &raw); err != nil {
		return nil, err
	}
	return &old, nil
}

// restoreTerminal puts the terminal back into its previous state.
func restoreTerminal(fd uintptr, state rawState) {
	if state == nil {
		return
	}
	_ = ioctl(fd, tcsetscmd, state)
}

func ioctl(fd uintptr, request uintptr, arg any) error {
	var ptr uintptr
	switch a := arg.(type) {
	case *syscall.Termios:
		ptr = uintptr(unsafe.Pointer(a))
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, ptr)
	if errno != 0 {
		return errno
	}
	return nil
}
