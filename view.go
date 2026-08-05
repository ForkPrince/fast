package main

import (
	"fmt"
	"strings"
)

const (
	sparkWidth = 20

	downloadColor = "46;248;187"
	uploadColor   = "189;82;255"

	downloadLabel = "↓"
	uploadLabel   = "↑"
)

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func fg(color, s string) string {
	return "\033[38;2;" + color + "m" + s + "\033[0m"
}

func gray(s string) string {
	return "\033[38;5;240m" + s + "\033[0m"
}

func renderRow(label string, currentSpeed float64, speeds []float64, peak float64, color string) string {
	var s strings.Builder
	s.WriteString(fg(color, label))
	s.WriteString(" ")
	speed, unit := scale(currentSpeed)
	s.WriteString(bold(fmt.Sprintf("%5.1f", speed)))
	s.WriteString(gray(" " + unit))
	s.WriteString(" ")
	s.WriteString(fg(color, sparkline(speeds, peak, sparkWidth)))
	if peak > 0 {
		peakVal, peakUnit := scale(peak)
		label := fmt.Sprintf("  peak %.0f", peakVal)
		if peakUnit != unit {
			label += " " + peakUnit
		}
		s.WriteString(gray(label))
	}
	return s.String()
}

func renderSummary(m Model) string {
	sep := gray(" • ")
	ping := gray("—")
	if m.ping > 0 {
		ping = bold(fmt.Sprintf("%d", m.ping.Milliseconds())) + gray(" ms")
	}
	return summarySpeed(downloadLabel, m.download.speed, downloadColor) + sep +
		summarySpeed(uploadLabel, m.upload.speed, uploadColor) + sep + ping
}

func summarySpeed(label string, speed float64, color string) string {
	value, unit := scale(speed)
	return fg(color, label) + " " +
		bold(fmt.Sprintf("%.1f", value)) +
		gray(" "+unit)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var content string
	switch m.phase {
	case phaseLoading, phaseDownloading:
		content = renderRow(downloadLabel, m.download.speed, m.download.samples, m.download.peak, downloadColor)
	case phaseUploading, phaseMeasuringPing:
		content = renderRow(uploadLabel, m.upload.speed, m.upload.samples, m.upload.peak, uploadColor)
	case phaseDone:
		content = renderSummary(m)
	}

	// Replicates lipgloss baseStyle Padding(1, 2); the done phase gets an
	// extra bottom line via PaddingBottom(2).
	content = "\n" + "  " + content + "  " + "\n"
	if m.phase == phaseDone {
		content += "\n"
	}
	return content
}

func scale(speed float64) (float64, string) {
	if speed >= 999.95 {
		return speed / 1000, "Gbps"
	}
	return speed, "Mbps"
}
