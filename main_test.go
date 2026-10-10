package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func testPicker() picker {
	return *newPicker([]list.Item{
		orgItem{Org{ID: "1", Name: "one"}},
		orgItem{Org{ID: "2", Name: "two"}},
	}, "test")
}

// key builds a key press the way the terminal would, so String() derives it
// from Code+Mod exactly as it does at runtime.
func key(code rune, mod tea.KeyMod, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Mod: mod, Text: text}
}

var (
	keyEsc    = key(tea.KeyEscape, 0, "")
	keyEnter  = key(tea.KeyEnter, 0, "")
	keyCtrlC  = key('c', tea.ModCtrl, "")
	keyLetter = func(r rune, text string) tea.KeyPressMsg { return key(r, 0, text) }
)

func TestEscAsksBeforeClosing(t *testing.T) {
	m := testPicker()
	next, cmd := m.handleKey(keyEsc)
	got := next.(picker)
	if !got.confirming {
		t.Fatal("esc should raise the quit prompt, not quit outright")
	}
	if cmd != nil {
		t.Fatal("esc must not quit before confirmation")
	}
	if got.selected != nil {
		t.Fatal("nothing should be selected while confirming")
	}
}

func TestConfirmYesQuits(t *testing.T) {
	m := testPicker()
	after, _ := m.handleKey(keyEsc)

	_, cmd := after.(picker).handleKey(keyLetter('y', "y"))
	if cmd == nil {
		t.Fatal("y should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestConfirmNoGoesBack(t *testing.T) {
	for _, k := range []tea.KeyPressMsg{
		keyLetter('n', "n"),
		keyEsc,
		keyLetter('q', "q"),
	} {
		m := testPicker()
		after, _ := m.handleKey(keyEsc)
		next, cmd := after.(picker).handleKey(k)
		if cmd != nil {
			t.Fatalf("%q should not quit", k.String())
		}
		if next.(picker).confirming {
			t.Fatalf("%q should dismiss the prompt", k.String())
		}
	}
}

func TestConfirmSwallowsOtherKeys(t *testing.T) {
	m := testPicker()
	after, _ := m.handleKey(keyEsc)
	next, _ := after.(picker).handleKey(keyLetter('x', "x"))
	if !next.(picker).confirming {
		t.Fatal("unrelated key must not dismiss the prompt")
	}
}

func TestEnterSelects(t *testing.T) {
	next, cmd := testPicker().handleKey(keyEnter)
	if cmd == nil {
		t.Fatal("enter should quit with a selection")
	}
	if next.(picker).selected == nil {
		t.Fatal("enter should select the highlighted item")
	}
	if next.(picker).confirming {
		t.Fatal("enter should not raise the quit prompt")
	}
}

func TestCtrlCAlwaysQuits(t *testing.T) {
	_, cmd := testPicker().handleKey(keyCtrlC)
	if cmd == nil {
		t.Fatal("ctrl+c must always quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestCtrlCQuitsEvenWhileFiltering(t *testing.T) {
	m := testPicker()
	m.list.SetFilterState(list.Filtering)
	_, cmd := m.handleKey(keyCtrlC)
	if cmd == nil {
		t.Fatal("ctrl+c must escape a stuck filter input")
	}
}

// "v" is the list's own quit binding and used to exit the app silently.
func TestVDoesNotQuit(t *testing.T) {
	_, cmd := testPicker().handleKey(keyLetter('v', "v"))
	if cmd != nil {
		t.Fatal("v must not quit")
	}
}

// esc should peel one layer at a time rather than jumping straight to quit.
func TestEscClearsFilterBeforeAsking(t *testing.T) {
	m := testPicker()
	m.list.SetFilterText("two")
	m.list.SetFilterState(list.FilterApplied)

	next, _ := m.handleKey(keyEsc)
	got := next.(picker)
	if got.confirming {
		t.Fatal("first esc should clear the filter, not ask to quit")
	}
	if got.list.FilterState() != list.Unfiltered {
		t.Fatalf("filter should be cleared, still %v", got.list.FilterState())
	}

	next2, _ := got.handleKey(keyEsc)
	if !next2.(picker).confirming {
		t.Fatal("second esc should ask to quit")
	}
}

// While typing, esc belongs to the filter input, not to us.
func TestEscDuringFilteringIsPassedThrough(t *testing.T) {
	m := testPicker()
	m.list.SetFilterState(list.Filtering)
	next, _ := m.handleKey(keyEsc)
	if next.(picker).confirming {
		t.Fatal("esc while filtering must not open the quit prompt")
	}
}

func TestFooterMentionsEveryKey(t *testing.T) {
	for _, k := range []string{"↑/↓", "enter", "/", "esc"} {
		if !strings.Contains(footer, k) {
			t.Errorf("footer missing %q: %q", k, footer)
		}
	}
	if strings.Contains(footer, "v ") {
		t.Error("footer should not advertise the list's v binding")
	}
}

func TestQuitPromptOffersEveryAnswer(t *testing.T) {
	for _, k := range []string{"y", "n", "esc"} {
		if !strings.Contains(quitPrompt, k) {
			t.Errorf("quit prompt missing %q: %q", k, quitPrompt)
		}
	}
}

func TestStatusLineSwapsForPrompt(t *testing.T) {
	m := testPicker()
	if m.statusLine() != footer {
		t.Fatal("idle status line should be the footer")
	}
	after, _ := m.handleKey(keyEsc)
	if after.(picker).statusLine() != quitPrompt {
		t.Fatal("confirming status line should be the quit prompt")
	}
}
