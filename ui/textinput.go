package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/10yearsahead/easy-git/i18n"
)

// TextInput is a reusable single-line text input with cursor movement support.
// Supports: typing, backspace, delete, left/right arrows, home/end, paste.
type TextInput struct {
	Value       string
	Placeholder string
	cursor      int // byte position of cursor in Value
}

func NewTextInput(placeholder string) TextInput {
	return TextInput{Placeholder: placeholder}
}

func (t TextInput) SetValue(v string) TextInput {
	t.Value = v
	t.cursor = len(v)
	return t
}

// Update handles key messages. Returns updated TextInput and whether Enter was pressed.
func (t TextInput) Update(msg tea.KeyMsg) (TextInput, bool) {
	switch msg.String() {
	case "enter":
		return t, true

	case "backspace":
		if t.cursor > 0 {
			runes := []rune(t.Value)
			pos := t.runePos()
			if pos > 0 {
				runes = append(runes[:pos-1], runes[pos:]...)
				t.Value = string(runes)
				t.cursor = len(string(runes[:pos-1]))
			}
		}

	case "delete":
		runes := []rune(t.Value)
		pos := t.runePos()
		if pos < len(runes) {
			runes = append(runes[:pos], runes[pos+1:]...)
			t.Value = string(runes)
		}

	case "left":
		if t.cursor > 0 {
			runes := []rune(t.Value[:t.cursor])
			if len(runes) > 0 {
				t.cursor = len(string(runes[:len(runes)-1]))
			}
		}

	case "right":
		if t.cursor < len(t.Value) {
			runes := []rune(t.Value[t.cursor:])
			if len(runes) > 0 {
				t.cursor += len(string(runes[:1]))
			}
		}

	case "home", "ctrl+a":
		t.cursor = 0

	case "end", "ctrl+e":
		t.cursor = len(t.Value)

	case "ctrl+u":
		t.Value = ""
		t.cursor = 0

	case "ctrl+w":
		if t.cursor > 0 {
			before := t.Value[:t.cursor]
			trimmed := strings.TrimRight(before, " ")
			lastSpace := strings.LastIndex(trimmed, " ")
			if lastSpace == -1 {
				t.Value = t.Value[t.cursor:]
				t.cursor = 0
			} else {
				newBefore := before[:lastSpace+1]
				t.Value = newBefore + t.Value[t.cursor:]
				t.cursor = len(newBefore)
			}
		}

	default:
		if IsTypeable(msg.String()) {
			insert := SanitizeInput(msg.String())
			if insert != "" {
				t.Value = t.Value[:t.cursor] + insert + t.Value[t.cursor:]
				t.cursor += len(insert)
			}
		}
	}
	return t, false
}

// runePos returns the rune index corresponding to the current byte cursor
func (t TextInput) runePos() int {
	return len([]rune(t.Value[:t.cursor]))
}

// View renders the input box
func (t TextInput) View() string {
	var b strings.Builder

	if t.Value == "" {
		b.WriteString(PanelStyle.Render(" " + MutedStyle.Render(t.Placeholder)))
		return b.String()
	}

	before := t.Value[:t.cursor]
	after := t.Value[t.cursor:]

	rendered := FileNameStyle.Render(before) +
		CursorStyle.Render("▌") +
		FileNameStyle.Render(after)

	b.WriteString(PanelStyle.Render(" " + rendered))
	return b.String()
}

// ViewWithLabel renders label + input + hint
func (t TextInput) ViewWithLabel(label, hint string) string {
	var b strings.Builder
	b.WriteString(InputPromptStyle.Render(label))
	b.WriteString("\n")
	b.WriteString(t.View())
	if hint != "" {
		b.WriteString("\n")
		b.WriteString(InputHintStyle.Render(hint))
	}
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render(i18n.T("help.cursor_confirm_back")))
	return b.String()
}
