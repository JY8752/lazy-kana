package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var palette = []string{"#FF75B5", "#FFB45E", "#F8E56B", "#8DEB9B", "#72DDF7", "#BDA0FF"}

func colorful(text string, offset int) string {
	var b strings.Builder
	for i, char := range characters(text) {
		b.WriteString(lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color(palette[(i+offset)%len(palette)])).Render(char))
	}
	return b.String()
}

func (m model) accent() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(palette[m.color]))
}

func (m model) panelWidth() int { return max(1, min(60, m.width-4)) }
func (m model) laneWidth() int  { return max(1, m.panelWidth()-4) }

func (m model) View() tea.View {
	var content string
	if m.width < 40 || m.height < 18 {
		content = "かなあそび\n40列 × 18行 以上に\nひろげてね\nEsc: もどる"
		if m.screen == ScreenMenu {
			content += "  q: おわり"
		}
	} else if m.screen == ScreenMenu {
		content = m.menuView()
	} else {
		title := menuItems[int(m.screen)-1]
		body, hint := m.playView()
		innerWidth := m.panelWidth() - 2
		body = lipgloss.Place(innerWidth, 11, lipgloss.Center, lipgloss.Center, body)
		panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(palette[m.color])).
			Render(lipgloss.JoinVertical(lipgloss.Center, colorful(title, m.color), "", body))
		content = lipgloss.JoinVertical(lipgloss.Center, panel, "", hint, "Esc: メニューへ")
	}
	// Clip by terminal cells, never by bytes (including during window resize).
	content = fit(content, m.width, m.height)
	v := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}

func (m model) menuView() string {
	var lines []string
	lines = append(lines, colorful("🌈 かなあそび", m.color), "", "おして、ながめて、あそぼう。", "")
	for i, item := range menuItems {
		label := fmt.Sprintf("  %d  %s", i+1, item)
		style := lipgloss.NewStyle().Width(30)
		if i == m.selected {
			label = fmt.Sprintf("▶ %d  %s", i+1, item)
			style = style.Bold(true).Foreground(lipgloss.Color("#20172E")).Background(lipgloss.Color("#F8E56B"))
		} else {
			style = style.Foreground(lipgloss.Color(palette[i]))
		}
		lines = append(lines, style.Render(label))
	}
	mode := "ふつう"
	if m.dotMode {
		mode = "ドット"
	}
	lines = append(lines, "", "↑ ↓ 選択   Enter 決定", "d: もじ="+mode+"   q: おわり")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BDA0FF")).Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Center, lines...))
}

func (m model) playView() (string, string) {
	switch m.screen {
	case ScreenGojuon:
		char := string(gojuon[m.kanaIndex])
		display, dotted := m.kanaDisplay(char)
		parts := []string{colorful("★  こんにちは！  ★", m.color), "", display, "", fmt.Sprintf("%d / %d", m.kanaIndex+1, len(gojuon))}
		if dotted {
			// Give the oversized glyph the vertical space normally used by the greeting.
			parts = []string{display, "", fmt.Sprintf("%d / %d", m.kanaIndex+1, len(gojuon))}
		}
		body := lipgloss.JoinVertical(lipgloss.Center, parts...)
		return body, "Space: つぎの もじ"
	case ScreenName:
		return m.nameRosterView(), "Enter: ついか   ↑↓: めいぼ   Backspace: けす"
	case ScreenMissing, ScreenMistake:
		return m.questionView()
	case ScreenTrain:
		return lipgloss.JoinVertical(lipgloss.Center, colorful("つぎは、"+gojuonRows[m.trainRow]+" えき！", m.color), "", m.trainView(), "",
			m.accent().Render(strings.Repeat("─", m.laneWidth()))), "Space: つぎの ぎょう"
	}
	return "", ""
}

func (m model) nameRosterView() string {
	value := m.nameInput
	if value == "" {
		value = "（なまえを いれてね）"
	}
	width := m.panelWidth() - 10
	if ansi.StringWidth(value) > width-2 {
		value = "…" + ansi.Cut(value, ansi.StringWidth(value)-(width-3), ansi.StringWidth(value))
	}
	input := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width).
		BorderForeground(lipgloss.Color("#72DDF7")).Render(m.accent().Render(value + " ▏"))

	countText := fmt.Sprintf("✨ めいぼ  %dにん ✨", len(m.nameRoster))
	maxStart := max(0, len(m.nameRoster)-5)
	start := min(m.nameScroll, maxStart)
	if len(m.nameRoster) > 5 {
		countText = fmt.Sprintf("✨ めいぼ %dにん  %d〜%d ✨", len(m.nameRoster), start+1, min(start+5, len(m.nameRoster)))
	}
	count := colorful(countText, m.color)
	rows := make([]string, 0, 5)
	for i := start; i < min(start+5, len(m.nameRoster)); i++ {
		name := ansi.Truncate(m.nameRoster[i], width-8, "…")
		row := fmt.Sprintf("%2d  %s", i+1, name)
		rows = append(rows, lipgloss.NewStyle().Width(width).Foreground(
			lipgloss.Color(palette[(m.color+i)%len(palette)])).Render(row))
	}
	if len(rows) == 0 {
		rows = append(rows, lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Foreground(lipgloss.Color("#8A958F")).Render("まだ だれも いないよ"))
	}

	parts := []string{"なまえを いれて Enter！", input, "", count}
	parts = append(parts, rows...)
	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}

func (m model) kanaDisplay(char string) (string, bool) {
	runes := []rune(char)
	var pattern [14]uint32
	ok := false
	if len(runes) == 1 {
		pattern, ok = dotKanaPatterns[runes[0]]
	}
	// Keep the compact card when a small terminal cannot show every dot.
	if !m.dotMode || len(runes) != 1 || !ok || m.width < 44 || m.height < 23 {
		return m.kanaCard(char), false
	}

	lines := make([]string, len(pattern))
	for row, dots := range pattern {
		var line strings.Builder
		line.Grow(dotKanaWidth * 2)
		for column := dotKanaWidth - 1; column >= 0; column-- {
			if dots&(1<<column) != 0 {
				line.WriteString("██")
			} else {
				line.WriteString("  ")
			}
		}
		lines[row] = lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color(palette[(m.color+row/2)%len(palette)])).
			Render(line.String())
	}
	return strings.Join(lines, "\n"), true
}

func (m model) kanaCard(char string) string {
	return lipgloss.NewStyle().Width(min(28, m.panelWidth()-8)).Height(5).
		Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(palette[m.color])).Render(m.accent().Render(char))
}

func (m model) questionView() (string, string) {
	row := []rune(gojuonRows[m.questionRow])
	cells := make([]string, len(row))
	for i, r := range row {
		char := string(r)
		if i == m.hiddenIndex && !m.revealed {
			if m.screen == ScreenMissing {
				char = "□"
			} else {
				char = string(m.replacement)
			}
		}
		style := m.accent()
		if m.revealed {
			style = style.Foreground(lipgloss.Color(palette[(m.color+i)%len(palette)]))
		}
		cells[i] = style.Width(4).Align(lipgloss.Center).Render(char)
	}
	line := strings.Join(cells, "")
	question := "なにが ない？"
	if m.screen == ScreenMistake {
		question = "どこが おかしい？"
	}
	if !m.revealed {
		return lipgloss.JoinVertical(lipgloss.Center, line, "", question), "Space: こたえを みる"
	}
	markers := make([]string, len(row))
	for i := range markers {
		mark := ""
		if i == m.hiddenIndex {
			mark = "↑"
		}
		markers[i] = m.accent().Width(4).Align(lipgloss.Center).Render(mark)
	}
	answer := "✨ " + string(row[m.hiddenIndex]) + "！ ✨"
	if m.screen == ScreenMistake {
		answer = "ここ！ " + string(m.replacement) + " → " + string(row[m.hiddenIndex]) + " ✨"
	}
	return lipgloss.JoinVertical(lipgloss.Center, line, strings.Join(markers, ""), "", colorful(answer, m.color)), "Space: つぎの もんだい"
}

func (m model) trainLines() []string {
	chars := characters(gojuonRows[m.trainRow])
	cargo := " " + strings.Join(chars, " ") + " "
	roof := strings.Repeat("─", ansi.StringWidth(cargo))
	return []string{
		"      ┌" + roof + "┐",
		"🚂====│" + cargo + "│==",
		"      └" + roof + "┘",
		"        ◉" + strings.Repeat(" ", ansi.StringWidth(cargo)-4) + "◉",
	}
}

func (m model) trainWidth() int { return ansi.StringWidth(m.trainLines()[1]) }

func (m model) trainView() string {
	lines := m.trainLines()
	for i, line := range lines {
		line = colorful(line, m.color+i)
		if m.trainX >= 0 {
			line = strings.Repeat(" ", m.trainX) + line
			line = ansi.Cut(line, 0, m.laneWidth())
		} else {
			line = ansi.Cut(line, -m.trainX, -m.trainX+m.laneWidth())
		}
		lines[i] = lipgloss.NewStyle().Width(m.laneWidth()).Render(line)
	}
	return strings.Join(lines, "\n")
}

func fit(s string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	lines = lines[:min(len(lines), height)]
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "")
	}
	return strings.Join(lines, "\n")
}
