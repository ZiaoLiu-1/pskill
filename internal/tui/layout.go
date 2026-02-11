package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout holds the calculated dimensions for a TUI view.
type Layout struct {
	Width, Height int
	LeftW, RightW int
	ContentH      int
	HasDetail     bool
	IsTooSmall    bool
}

// ComputeLayout calculates pane dimensions based on terminal size and context.
func ComputeLayout(w, h int, showDetail bool) Layout {
	l := Layout{
		Width:  w,
		Height: h,
	}

	// Enforce minimum dimensions (60x10)
	if w < 60 || h < 10 {
		l.IsTooSmall = true
		return l
	}

	// Content area height: total - header(1) - tabbar(1) - sep(1) - help(1)
	l.ContentH = h - 4
	if l.ContentH < 1 {
		l.ContentH = 1
	}

	if showDetail {
		l.HasDetail = true
		// 60% list, 40% detail, ensuring min widths
		l.LeftW = (w * 60) / 100
		if l.LeftW < 30 {
			l.LeftW = 30
		}
		l.RightW = w - l.LeftW - 2 // -2 for border/gap
		if l.RightW < 20 {
			// If right pane too small, fallback to single pane
			l.HasDetail = false
			l.LeftW = w - 2
			l.RightW = 0
		}
	} else {
		l.LeftW = w - 2
		l.RightW = 0
	}

	return l
}

func RenderTooSmall(w, h int) string {
	msg := "Terminal too small.\nResize to at least 60x10."
	style := lipgloss.NewStyle().
		Width(w).Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorDanger)
	return style.Render(msg)
}

func Separator(w int) string {
	return dimStyle.Render(strings.Repeat("─", w))
}
