package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Msg is the interface passed between commands and the model, mirroring
// bubbletea's message model: anything can be a message.
type Msg interface{}

// Cmd is a unit of work that eventually produces a Msg (or nil).
type Cmd func() Msg

// App is the core of the program: it receives messages, updates state, and
// renders its view as a string.
type App interface {
	Init() Cmd
	Update(Msg) (App, Cmd)
	View() string
}

// KeyMsg represents a single key press read from the terminal.
type KeyMsg struct {
	Type string
}

func (k KeyMsg) String() string { return k.Type }

type quitMsg struct{}

// Quit tells the program to stop the event loop and return the final model.
var Quit Cmd = func() Msg { return quitMsg{} }

// Tick returns a Cmd that sleeps for d then runs fn with the current time.
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		time.Sleep(d)
		return fn(time.Now())
	}
}

// Batch runs the given commands concurrently, forwarding each non-nil result
// to the running program as it completes.
var active *program

func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		p := active
		if p == nil {
			return nil
		}
		var wg sync.WaitGroup
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			wg.Add(1)
			go func(cmd Cmd) {
				defer wg.Done()
				if msg := cmd(); msg != nil {
					p.msgs <- msg
				}
			}(cmd)
		}
		wg.Wait()
		return nil
	}
}

// program drives an App's event loop.
type program struct {
	model App
	msgs  chan Msg
}

// Run starts the terminal program, blocking until the model quits. It puts the
// terminal into raw mode, hides the cursor, and renders the model's view on
// every message.
func Run(initial App) (App, error) {
	fd := os.Stdin.Fd()

	isTerm, err := isTerminal(fd)
	if err != nil {
		return nil, err
	}

	var oldState rawState
	if isTerm {
		oldState, err = makeRaw(fd)
		if err != nil {
			return nil, fmt.Errorf("raw mode: %w", err)
		}
		defer restoreTerminal(fd, oldState)
	}

	p := &program{model: initial, msgs: make(chan Msg)}
	active = p
	defer func() { active = nil }()

	hideCursor()
	defer showCursor()

	clearScreen()
	p.render()

	if isTerm {
		go p.readKeys()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		if _, ok := <-sigCh; ok {
			p.msgs <- KeyMsg{Type: "ctrl+c"}
		}
	}()

	p.runCmd(initial.Init())

	for {
		select {
		case msg := <-p.msgs:
			switch msg.(type) {
			case quitMsg:
				p.render()
				return p.model, nil
			default:
				next, cmd := p.model.Update(msg)
				p.model = next
				p.render()
				p.runCmd(cmd)
			}
		}
	}
}

func (p *program) runCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	go func() {
		if msg := cmd(); msg != nil {
			p.msgs <- msg
		}
	}()
}

func (p *program) render() {
	clearScreen()
	fmt.Print(p.model.View())
}

// readKeys reads raw bytes from stdin, decoding them into KeyMsg values.
func (p *program) readKeys() {
	buf := make([]byte, 128)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		for i := 0; i < n; i++ {
			b := buf[i]
			switch {
			case b == 0x03: // Ctrl+C
				p.msgs <- KeyMsg{Type: "ctrl+c"}
			case b == 0x1b: // ESC, possibly the start of a sequence
				if i+1 < n && buf[i+1] == '[' {
					i += 2
					for i < n && buf[i] >= 'A' && buf[i] <= '~' {
						i++
					}
				} else {
					p.msgs <- KeyMsg{Type: "esc"}
				}
			default:
				p.msgs <- KeyMsg{Type: string(b)}
			}
		}
	}
}

const (
	hideCursorSeq  = "\033[?25l"
	showCursorSeq  = "\033[?25h"
	clearScreenSeq = "\033[H\033[2J"
)

func hideCursor()  { fmt.Print(hideCursorSeq) }
func showCursor()  { fmt.Print(showCursorSeq) }
func clearScreen() { fmt.Print(clearScreenSeq) }
