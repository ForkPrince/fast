package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	connections = 5
	duration    = 10 * time.Second
	sparkWidth  = 20
)

const tickInterval = time.Second / 10

type Model struct {
	targets []string
	bytes   *atomic.Int64
	ctx     context.Context
	cancel  context.CancelFunc
	start   time.Time
	speed   float64
	speeds  []float64
	peak    float64
	done    bool
}

func NewModel(targets []string) Model {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	return Model{
		targets: targets,
		bytes:   &atomic.Int64{},
		ctx:     ctx,
		cancel:  cancel,
		start:   time.Now(),
	}
}

func (m *Model) run() {
	for _, url := range m.targets {
		go download(m.ctx, url, m.bytes)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	for {
		select {
		case <-m.ctx.Done():
			m.done = true
			m.render()
			fmt.Println()
			return
		case <-ticker.C:
			elapsed := time.Since(m.start)
			m.speed = mbps(m.bytes.Load(), elapsed)
			m.speeds = append(m.speeds, m.speed)
			if m.speed > m.peak {
				m.peak = m.speed
			}
			m.render()
		case <-sigCh:
			m.cancel()
			return
		}
	}
}

func mbps(bytes int64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(bytes) * 8 / d.Seconds() / 1e6
}
