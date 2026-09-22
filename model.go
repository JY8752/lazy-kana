package main

import (
	"math/rand/v2"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/rivo/uniseg"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenGojuon
	ScreenName
	ScreenMissing
	ScreenMistake
	ScreenTrain
)

const maxNameLength = 32

var menuItems = []string{"あいうえお", "おなまえ", "むし食い", "まちがい探し", "ひらがな列車 🚂"}

type tickMsg time.Time

type model struct {
	screen        Screen
	width, height int
	selected      int
	kanaIndex     int
	color         int
	ticks         int
	nameInput     string
	name          []string
	nameIndex     int
	namePlaying   bool
	questionRow   int
	hiddenIndex   int
	replacement   rune
	revealed      bool
	trainRow      int
	trainX        int
}

func initialModel() model {
	return model{width: 80, height: 24}
}

func nextTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd { return nextTick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
		m.trainX = min(m.trainX, m.laneWidth())
	case tickMsg:
		// One timer chain for the whole app: changing screens never adds timers.
		m.ticks = (m.ticks + 1) % 8
		if m.ticks == 0 {
			m.nextColor()
		}
		if m.screen == ScreenTrain {
			m.trainX--
			if m.trainX < -m.trainWidth() {
				m.trainX = m.laneWidth()
			}
		}
		return m, nextTick()
	case tea.PasteMsg:
		if m.screen == ScreenName && !m.namePlaying {
			m.appendName(msg.Content)
		}
	case tea.KeyPressMsg:
		key := msg.String()
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.screen = ScreenMenu
			return m, nil
		}
		if m.screen == ScreenMenu {
			switch key {
			case "q":
				return m, tea.Quit
			case "up":
				m.selected = (m.selected + len(menuItems) - 1) % len(menuItems)
			case "down":
				m.selected = (m.selected + 1) % len(menuItems)
			case "enter":
				m.openScreen(Screen(m.selected + 1))
			case "1", "2", "3", "4", "5":
				m.selected = int(key[0] - '1')
				m.openScreen(Screen(m.selected + 1))
			}
			return m, nil
		}
		if m.screen == ScreenName && !m.namePlaying {
			switch key {
			case "enter":
				m.name = characters(strings.TrimSpace(m.nameInput))
				m.namePlaying = len(m.name) > 0
				m.nameIndex = 0
			case "backspace", "delete":
				chars := characters(m.nameInput)
				if len(chars) > 0 {
					m.nameInput = strings.Join(chars[:len(chars)-1], "")
				}
			default:
				// Text is empty for function keys and modified shortcuts.
				m.appendName(msg.Text)
			}
			return m, nil
		}
		if key == "space" || key == " " {
			m.nextColor()
			switch m.screen {
			case ScreenGojuon:
				m.kanaIndex = (m.kanaIndex + 1) % len(gojuon)
			case ScreenName:
				if m.nameIndex < len(m.name) {
					m.nameIndex++
				} else {
					m.nameIndex = 0
				}
			case ScreenMissing, ScreenMistake:
				if m.revealed {
					m.newQuestion()
				} else {
					m.revealed = true
				}
			case ScreenTrain:
				m.trainRow = (m.trainRow + 1) % len(gojuonRows)
				m.trainX = m.laneWidth() - 2
			}
		}
	}
	return m, nil
}

func (m *model) openScreen(screen Screen) {
	m.screen = screen
	switch screen {
	case ScreenGojuon:
		m.kanaIndex = 0
	case ScreenName:
		m.namePlaying = false
		m.nameIndex = 0
	case ScreenMissing, ScreenMistake:
		m.newQuestion()
	case ScreenTrain:
		m.trainRow = 0
		m.trainX = m.laneWidth() - 2
	}
}

func (m *model) nextColor() { m.color = (m.color + 1) % len(palette) }

func (m *model) newQuestion() {
	m.questionRow = rand.IntN(len(gojuonRows))
	row := []rune(gojuonRows[m.questionRow])
	m.hiddenIndex = rand.IntN(len(row))
	// Choose a different kana without retries, even if keys are mashed.
	correct := row[m.hiddenIndex]
	index := slices.Index(gojuon, correct)
	m.replacement = gojuon[(index+1+rand.IntN(len(gojuon)-1))%len(gojuon)]
	m.revealed = false
}

// Grapheme clusters keep combining marks and emoji sequences together.
func characters(s string) []string {
	var result []string
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		result = append(result, g.Str())
	}
	return result
}

func (m *model) appendName(text string) {
	// Bound both input bytes and character count; discard terminal controls.
	var b strings.Builder
	for _, r := range text {
		if b.Len()+len(m.nameInput)+utf8.RuneLen(r) > 512 {
			break
		}
		if unicode.IsPrint(r) || r == '\u200d' {
			b.WriteRune(r)
		}
	}
	chars := characters(m.nameInput + b.String())
	m.nameInput = strings.Join(chars[:min(len(chars), maxNameLength)], "")
}
