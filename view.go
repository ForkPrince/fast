package main

import "fmt"

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

func scale(speed float64) (float64, string) {
	if speed >= 999.95 {
		return speed / 1000, "Gbps"
	}
	return speed, "Mbps"
}
