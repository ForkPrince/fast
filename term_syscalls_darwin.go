//go:build darwin

package main

import "syscall"

const (
	tcgetscmd = syscall.TIOCGETA
	tcsetscmd = syscall.TIOCSETA
)
