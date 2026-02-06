// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func TestWithOrdinals(t *testing.T) {
	t.Run("prefixes string options", func(t *testing.T) {
		opts := withOrdinals([]huh.Option[string]{
			huh.NewOption("alpha", "a"),
			huh.NewOption("beta", "b"),
			huh.NewOption("gamma", "c"),
		})
		want := []string{"1. alpha", "2. beta", "3. gamma"}
		for i, opt := range opts {
			if opt.Key != want[i] {
				t.Errorf("option %d Key = %q, want %q", i, opt.Key, want[i])
			}
		}
	})

	t.Run("prefixes uint options", func(t *testing.T) {
		opts := withOrdinals([]huh.Option[uint]{
			huh.NewOption("First", uint(1)),
			huh.NewOption("Second", uint(2)),
		})
		if opts[0].Key != "1. First" {
			t.Errorf("option 0 Key = %q, want %q", opts[0].Key, "1. First")
		}
		if opts[1].Key != "2. Second" {
			t.Errorf("option 1 Key = %q, want %q", opts[1].Key, "2. Second")
		}
	})

	t.Run("double-digit ordinals", func(t *testing.T) {
		opts := make([]huh.Option[string], 12)
		for i := range opts {
			opts[i] = huh.NewOption("item", "v")
		}
		opts = withOrdinals(opts)
		if opts[9].Key != "10. item" {
			t.Errorf("option 9 Key = %q, want %q", opts[9].Key, "10. item")
		}
		if opts[11].Key != "12. item" {
			t.Errorf("option 11 Key = %q, want %q", opts[11].Key, "12. item")
		}
	})
}

func TestSelectOrdinal(t *testing.T) {
	tests := []struct {
		name   string
		msg    tea.KeyMsg
		wantN  int
		wantOk bool
	}{
		{
			name:   "key 1",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}},
			wantN:  1,
			wantOk: true,
		},
		{
			name:   "key 5",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}},
			wantN:  5,
			wantOk: true,
		},
		{
			name:   "key 9",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}},
			wantN:  9,
			wantOk: true,
		},
		{
			name:   "key 0 is not an ordinal",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}},
			wantN:  0,
			wantOk: false,
		},
		{
			name:   "letter key",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantN:  0,
			wantOk: false,
		},
		{
			name:   "enter key",
			msg:    tea.KeyMsg{Type: tea.KeyEnter},
			wantN:  0,
			wantOk: false,
		},
		{
			name:   "empty runes",
			msg:    tea.KeyMsg{Type: tea.KeyRunes, Runes: nil},
			wantN:  0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := selectOrdinal(tt.msg)
			if n != tt.wantN || ok != tt.wantOk {
				t.Errorf("selectOrdinal() = (%d, %v), want (%d, %v)",
					n, ok, tt.wantN, tt.wantOk)
			}
		})
	}
}

func TestIsSelectField(t *testing.T) {
	t.Run("select field returns true", func(t *testing.T) {
		val := "a"
		sel := huh.NewSelect[string]().
			Options(
				huh.NewOption("Alpha", "a"),
				huh.NewOption("Beta", "b"),
				huh.NewOption("Gamma", "c"),
			).
			Value(&val)
		form := huh.NewForm(huh.NewGroup(sel))
		form.Init()

		if !isSelectField(form) {
			t.Error("expected isSelectField=true for a Select field")
		}
	})

	t.Run("input field returns false", func(t *testing.T) {
		val := ""
		inp := huh.NewInput().Title("Name").Value(&val)
		form := huh.NewForm(huh.NewGroup(inp))
		form.Init()

		if isSelectField(form) {
			t.Error("expected isSelectField=false for an Input field")
		}
	})
}

func TestSelectOptionCount(t *testing.T) {
	val := "a"
	sel := huh.NewSelect[string]().
		Options(
			huh.NewOption("Alpha", "a"),
			huh.NewOption("Beta", "b"),
			huh.NewOption("Gamma", "c"),
		).
		Value(&val)
	form := huh.NewForm(huh.NewGroup(sel))
	form.Init()

	field := form.GetFocusedField()
	count := selectOptionCount(field)
	if count != 3 {
		t.Errorf("selectOptionCount = %d, want 3", count)
	}
}

func TestSelectOptionCountForInput(t *testing.T) {
	val := ""
	inp := huh.NewInput().Title("Name").Value(&val)
	form := huh.NewForm(huh.NewGroup(inp))
	form.Init()

	field := form.GetFocusedField()
	count := selectOptionCount(field)
	if count != -1 {
		t.Errorf("selectOptionCount = %d, want -1 for Input", count)
	}
}

func TestJumpSelectToOrdinal(t *testing.T) {
	t.Run("jumps to correct option", func(t *testing.T) {
		val := "a"
		sel := huh.NewSelect[string]().
			Options(
				huh.NewOption("Alpha", "a"),
				huh.NewOption("Beta", "b"),
				huh.NewOption("Gamma", "c"),
			).
			Value(&val)

		form := huh.NewForm(huh.NewGroup(sel))
		form.Init()

		m := &Model{form: form}
		m.jumpSelectToOrdinal(2) // should jump to "Beta"

		if val != "b" {
			t.Errorf("expected val=%q after jumping to ordinal 2, got %q", "b", val)
		}
	})

	t.Run("ordinal 1 selects first option", func(t *testing.T) {
		val := "c"
		sel := huh.NewSelect[string]().
			Options(
				huh.NewOption("Alpha", "a"),
				huh.NewOption("Beta", "b"),
				huh.NewOption("Gamma", "c"),
			).
			Value(&val)

		form := huh.NewForm(huh.NewGroup(sel))
		form.Init()

		m := &Model{form: form}
		m.jumpSelectToOrdinal(1)

		if val != "a" {
			t.Errorf("expected val=%q after jumping to ordinal 1, got %q", "a", val)
		}
	})

	t.Run("ordinal exceeding option count is ignored", func(t *testing.T) {
		val := "a"
		sel := huh.NewSelect[string]().
			Options(
				huh.NewOption("Alpha", "a"),
				huh.NewOption("Beta", "b"),
			).
			Value(&val)

		form := huh.NewForm(huh.NewGroup(sel))
		form.Init()

		m := &Model{form: form}
		m.jumpSelectToOrdinal(5) // exceeds 2 options

		if val != "a" {
			t.Errorf("expected val=%q (unchanged) when ordinal exceeds count, got %q", "a", val)
		}
	})

	t.Run("works with uint select", func(t *testing.T) {
		val := uint(10)
		sel := huh.NewSelect[uint]().
			Options(
				huh.NewOption("First", uint(10)),
				huh.NewOption("Second", uint(20)),
				huh.NewOption("Third", uint(30)),
			).
			Value(&val)

		form := huh.NewForm(huh.NewGroup(sel))
		form.Init()

		m := &Model{form: form}
		m.jumpSelectToOrdinal(3)

		if val != 30 {
			t.Errorf("expected val=%d after jumping to ordinal 3, got %d", 30, val)
		}
	})
}
