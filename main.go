package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
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

func (m *Model) render() {
	fmt.Print("\033[2K\033[G")

	speed, unit := scale(m.speed)
	peak, peakUnit := scale(m.peak)

	accent := "\033[38;2;46;248;187m"
	gray := "\033[38;5;240m"
	reset := "\033[0m"
	bold := "\033[1m"

	line := bold + accent + fmt.Sprintf("%5.1f", speed) + reset
	line += gray + " " + unit + reset
	line += " "
	line += accent + sparkline(m.speeds, m.peak, sparkWidth) + reset

	if m.peak > 0 {
		var label string
		if peakUnit != unit {
			label = fmt.Sprintf("  peak %.0f %s", peak, peakUnit)
		} else {
			label = fmt.Sprintf("  peak %.0f", peak)
		}
		line += gray + label + reset
	}

	fmt.Print(line)
}

func mbps(bytes int64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(bytes) * 8 / d.Seconds() / 1e6
}

func scale(speed float64) (float64, string) {
	if speed >= 999.95 {
		return speed / 1000, "Gbps"
	}
	return speed, "Mbps"
}

func main() {
	urls, err := targets(connections)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			fmt.Fprintln(os.Stderr, "No internet connection.")
			os.Exit(1)
		}
		log.Fatal(err)
	}

	m := NewModel(urls)
	m.run()
}
