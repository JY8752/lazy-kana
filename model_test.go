package main

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func press(m model, code rune, text string) model {
	next, _ := m.Update(tea.KeyPressMsg{Code: code, Text: text})
	return next.(model)
}

func TestGojuonCycle(t *testing.T) {
	m := initialModel()
	m = press(m, tea.KeyEnter, "")
	for _, want := range gojuon {
		if got := gojuon[m.kanaIndex]; got != want {
			t.Fatalf("got %c, want %c", got, want)
		}
		oldColor := m.color
		m = press(m, tea.KeySpace, " ")
		if m.color == oldColor {
			t.Fatal("Space must change the color")
		}
	}
	if m.kanaIndex != 0 {
		t.Fatal("gojuon did not wrap to あ")
	}
}

func TestDisplayMode(t *testing.T) {
	m := initialModel()
	if _, dotted := m.kanaDisplay("あ"); dotted {
		t.Fatal("normal card must be the default display mode")
	}
	m = press(m, 'd', "d")
	if !m.dotMode {
		t.Fatal("d must enable dot mode from the menu")
	}

	display, dotted := m.kanaDisplay("あ")
	if !dotted {
		t.Fatal("あ must use the large dot-art glyph on an 80x24 terminal")
	}
	plain := ansi.Strip(display)
	lines := strings.Split(plain, "\n")
	if len(lines) != 14 || strings.Count(plain, "█") < 100 {
		t.Fatalf("dot art is not large enough: %d lines, %d blocks", len(lines), strings.Count(plain, "█"))
	}
	for _, line := range lines {
		if ansi.StringWidth(line) != 36 {
			t.Fatalf("dot-art row width = %d, want 36", ansi.StringWidth(line))
		}
	}
	assertFits(t, m.View().Content, 80, 24)
	for index, char := range gojuon {
		display, dotted = m.kanaDisplay(string(char))
		if !dotted {
			t.Fatalf("%c is missing its dot-art glyph", char)
		}
		lines = strings.Split(ansi.Strip(display), "\n")
		if len(lines) != 14 {
			t.Fatalf("%c has %d rows, want 14", char, len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) != 36 {
				t.Fatalf("%c row width = %d, want 36", char, ansi.StringWidth(line))
			}
		}
		m.kanaIndex = index
		m.screen = ScreenGojuon
		assertFits(t, m.View().Content, 80, 24)
	}
	if len(dotKanaPatterns) != len(gojuon) {
		t.Fatalf("dot font has %d glyphs, want %d", len(dotKanaPatterns), len(gojuon))
	}

	m.width, m.height = 40, 18
	_, dotted = m.kanaDisplay("あ")
	if dotted {
		t.Fatal("small terminals must fall back to the compact kana card")
	}
	m.width, m.height = 44, 22
	if _, dotted = m.kanaDisplay("あ"); dotted {
		t.Fatal("a terminal shorter than the complete dot view must use the compact card")
	}

	m.width, m.height = 80, 24
	if _, dotted = m.kanaDisplay("が"); dotted {
		t.Fatal("unsupported name characters must use the compact kana card")
	}
	m = press(m, tea.KeyEsc, "")
	m = press(m, 'd', "d")
	if m.dotMode {
		t.Fatal("d must switch dot mode off")
	}
}

func TestNameRoster(t *testing.T) {
	m := initialModel()
	m = press(m, '2', "2")
	m = press(m, tea.KeyEnter, "")
	if len(m.nameRoster) != 0 {
		t.Fatal("an empty name must not be added to the roster")
	}
	next, _ := m.Update(tea.PasteMsg{Content: "たか\u3099👨‍👩‍👧q\n\t"})
	m = press(next.(model), tea.KeyBackspace, "")
	m = press(m, tea.KeyEnter, "")
	if len(m.nameRoster) != 1 || m.nameRoster[0] != "たか\u3099👨‍👩‍👧" {
		t.Fatalf("Unicode name was not added intact: %+v", m.nameRoster)
	}
	if m.nameInput != "" {
		t.Fatal("the input must be cleared after adding a name")
	}

	for i := 1; i <= 7; i++ {
		m.appendName(fmt.Sprintf("なまえ%d", i))
		m = press(m, tea.KeyEnter, "")
	}
	view := ansi.Strip(m.View().Content)
	if len(m.nameRoster) != 8 || !strings.Contains(view, "8にん") || !strings.Contains(view, "なまえ7") {
		t.Fatalf("roster did not grow or show its latest entry: %+v", m.nameRoster)
	}
	if strings.Contains(view, "なまえ1") {
		t.Fatal("a long roster must show the latest five names")
	}
	m = press(m, tea.KeyUp, "")
	m = press(m, tea.KeyUp, "")
	m = press(m, tea.KeyUp, "")
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "たか\u3099👨‍👩‍👧") || strings.Contains(view, "なまえ7") {
		t.Fatal("up must scroll back through the roster")
	}
	m = press(m, tea.KeyDown, "")
	if m.nameScroll != 1 {
		t.Fatalf("down must scroll toward newer names: offset %d", m.nameScroll)
	}

	m = press(m, tea.KeyEsc, "")
	m = press(m, '2', "2")
	if len(m.nameRoster) != 8 {
		t.Fatal("the roster must survive returning to the menu")
	}
	m.appendName(strings.Repeat("あ", 10000))
	if len(characters(m.nameInput)) > maxNameLength || len(m.nameInput) > 512 {
		t.Fatal("long paste exceeded the input limit")
	}
}

func TestQuestionRevealAndNext(t *testing.T) {
	for _, screen := range []Screen{ScreenMissing, ScreenMistake} {
		m := initialModel()
		m.openScreen(screen)
		for i := 0; i < 100; i++ {
			row := []rune(gojuonRows[m.questionRow])
			correct := row[m.hiddenIndex]
			if m.revealed || correct == m.replacement {
				t.Fatal("new question has an invalid replacement or is already revealed")
			}
			before, _ := m.questionView()
			if screen == ScreenMissing && !strings.Contains(ansi.Strip(before), "□") {
				t.Fatal("missing kana must be hidden")
			}
			m = press(m, tea.KeySpace, " ")
			answer, _ := m.questionView()
			if !m.revealed || !strings.Contains(ansi.Strip(answer), string(correct)) || !strings.Contains(ansi.Strip(answer), "↑") {
				t.Fatal("answer must reveal the correct kana and its position")
			}
			m = press(m, tea.KeySpace, " ")
		}
	}
}

func TestKeyboardAndViewport(t *testing.T) {
	for screen := ScreenMenu; screen <= ScreenTrain; screen++ {
		m := initialModel()
		m.openScreen(screen)
		// Repeated function keys, modifiers, mouse events and release events
		// must not navigate, exit or corrupt the model.
		for i := 0; i < 100; i++ {
			for _, msg := range []tea.Msg{tea.KeyPressMsg{Code: tea.KeyF1}, tea.KeyReleaseMsg{Code: tea.KeyEnter}, tea.MouseClickMsg{}} {
				next, cmd := m.Update(msg)
				m = next.(model)
				if m.screen != screen || cmd != nil {
					t.Fatal("unrelated input changed the screen or ran a command")
				}
			}
		}
		for _, size := range [][2]int{{80, 24}, {40, 18}, {120, 40}, {20, 8}, {1, 1}, {0, 0}} {
			next, _ := m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			m = next.(model)
			assertFits(t, m.View().Content, size[0], size[1])
		}
		m = press(m, tea.KeyEsc, "")
		if m.screen != ScreenMenu {
			t.Fatal("Esc must always return to menu")
		}
	}
	m := initialModel()
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("q must quit from the menu")
	}
	m.openScreen(ScreenTrain)
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd != nil {
		t.Fatal("q must not quit during play")
	}
}

func TestTrainAnimationAndClipping(t *testing.T) {
	m := initialModel()
	m.openScreen(ScreenTrain)
	start := m.trainX
	next, cmd := m.Update(tickMsg{})
	m = next.(model)
	if m.trainX != start-1 || cmd == nil {
		t.Fatal("tick must move the train left and schedule another tick")
	}
	for x := -m.trainWidth(); x <= m.laneWidth(); x++ {
		m.trainX = x
		assertFits(t, m.trainView(), m.laneWidth(), 4)
	}
	for i := range gojuonRows {
		if m.trainRow != i {
			t.Fatal("train row out of order")
		}
		m = press(m, tea.KeySpace, " ")
	}
	if m.trainRow != 0 {
		t.Fatal("train rows must wrap")
	}
}

func assertFits(t *testing.T, view string, width, height int) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("rendering broke UTF-8")
	}
	if view == "" {
		return
	}
	lines := strings.Split(view, "\n")
	if len(lines) > height {
		t.Fatalf("view has %d lines, height %d", len(lines), height)
	}
	for _, line := range lines {
		if ansi.StringWidth(line) > width {
			t.Fatalf("view exceeds width %d: %q", width, ansi.Strip(line))
		}
	}
}
