//go:build linux || freebsd || netbsd || openbsd || dragonfly || solaris

package main

import "syscall"

const (
	tcgetscmd = syscall.TCGETS
	tcsetscmd = syscall.TCSETS
)
