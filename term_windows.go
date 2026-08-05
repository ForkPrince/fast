//go:build windows

package main

type rawState = struct{}

// isTerminal always reports false on Windows; raw mode is unsupported.
func isTerminal(fd uintptr) (bool, error) { return false, nil }

// makeRaw is a no-op on Windows.
func makeRaw(fd uintptr) (rawState, error) { return struct{}{}, nil }

// restoreTerminal is a no-op on Windows.
func restoreTerminal(fd uintptr, state rawState) {}
