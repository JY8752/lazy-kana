package main

import (
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

func TestNameUnicodeAndCompletion(t *testing.T) {
	m := initialModel()
	m = press(m, '2', "2")
	m = press(m, tea.KeyEnter, "")
	if m.namePlaying {
		t.Fatal("empty name must stay in the input screen")
	}
	next, _ := m.Update(tea.PasteMsg{Content: "たか\u3099👨‍👩‍👧q\n\t"})
	m = press(next.(model), tea.KeyBackspace, "")
	m = press(m, tea.KeyEnter, "")
	want := []string{"た", "か\u3099", "👨‍👩‍👧"}
	for i, char := range want {
		if !m.namePlaying || len(m.name) != 3 || m.name[m.nameIndex] != char {
			t.Fatalf("name character %d: %+v", i, m.name)
		}
		m = press(m, tea.KeySpace, " ")
	}
	if m.nameIndex != len(m.name) || !strings.Contains(ansi.Strip(m.View().Content), "よめたね") {
		t.Fatal("full name celebration missing")
	}
	m = press(m, tea.KeySpace, " ")
	if m.nameIndex != 0 {
		t.Fatal("name did not restart")
	}
	m = press(m, tea.KeyEsc, "")
	m = press(m, '2', "2")
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
