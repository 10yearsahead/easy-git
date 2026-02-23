package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Palette ──────────────────────────────────────────────────────────────────

var (
	ColorPrimary   = lipgloss.Color("#7C6AF7") // soft purple
	ColorSecondary = lipgloss.Color("#5DD8A4") // mint green
	ColorMuted     = lipgloss.Color("#6B7280") // gray
	ColorDanger    = lipgloss.Color("#F87171") // soft red
	ColorWarning   = lipgloss.Color("#FBBF24") // amber
	ColorSuccess   = lipgloss.Color("#34D399") // green
	ColorText      = lipgloss.Color("#F3F4F6") // near white
	ColorSubtext   = lipgloss.Color("#9CA3AF") // light gray
	ColorBorder    = lipgloss.Color("#374151") // dark gray
	ColorHighlight = lipgloss.Color("#1F2937") // dark bg highlight
)

// ── Base styles ───────────────────────────────────────────────────────────────

var (
	// App container
	AppStyle = lipgloss.NewStyle().
		Padding(1, 2)

	// Title banner
	TitleStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		PaddingBottom(0)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(ColorSubtext).
		PaddingBottom(1)

	// Section headers inside views
	SectionTitleStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true).
		PaddingTop(1).
		PaddingBottom(0)

	// Normal menu item
	MenuItemStyle = lipgloss.NewStyle().
		Foreground(ColorText).
		PaddingLeft(2)

	// Selected menu item
	MenuItemSelectedStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		PaddingLeft(0)

	// Cursor indicator
	CursorStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	// Dimmed / muted text
	MutedStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	// Error / warning message
	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning)

	// Help bar at the bottom
	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		PaddingTop(1)

	// Box / panel
	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1).
		MarginTop(1)

	// Active panel
	PanelActiveStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1).
		MarginTop(1)

	// File in the status list
	FileNameStyle = lipgloss.NewStyle().
		Foreground(ColorText)

	FileTagStagedStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	FileTagModifiedStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)

	FileTagNewStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true)

	FileTagDeletedStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).
		Bold(true)

	// Badge: current branch label
	BranchBadgeStyle = lipgloss.NewStyle().
		Background(ColorPrimary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1).
		MarginRight(1)

	// Checkbox styles
	CheckboxChecked = lipgloss.NewStyle().
		Foreground(ColorSuccess).Bold(true)

	CheckboxUnchecked = lipgloss.NewStyle().
		Foreground(ColorMuted)

	// Text input prompt
	InputPromptStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	InputHintStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// Step indicator
	StepStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true)

	// Divider
	DividerStyle = lipgloss.NewStyle().
		Foreground(ColorBorder)
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// Divider returns a horizontal rule string
func Divider(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return DividerStyle.Render(line)
}

// FileIcon returns an icon based on git status code
func FileIcon(code string) string {
	icons := map[string]string{
		"M": "~",
		"A": "+",
		"D": "-",
		"R": "→",
		"U": "!",
		"?": "+",
	}
	if icon, ok := icons[code]; ok {
		return icon
	}
	return "·"
}

// IsTypeable returns true if the key message is printable text (including paste).
// Rejects control keys like arrows, ctrl combos, function keys, etc.
func IsTypeable(key string) bool {
	// Reject known special keys
	special := map[string]bool{
		"up": true, "down": true, "left": true, "right": true,
		"enter": true, "esc": true, "backspace": true, "delete": true,
		"tab": true, "shift+tab": true, "home": true, "end": true,
		"pgup": true, "pgdown": true, "f1": true, "f2": true,
		"f3": true, "f4": true, "f5": true, "f6": true, "f7": true,
		"f8": true, "f9": true, "f10": true, "f11": true, "f12": true,
		"ctrl+a": true, "ctrl+b": true, "ctrl+c": true, "ctrl+d": true,
		"ctrl+e": true, "ctrl+f": true, "ctrl+g": true, "ctrl+h": true,
		"ctrl+i": true, "ctrl+j": true, "ctrl+k": true, "ctrl+l": true,
		"ctrl+m": true, "ctrl+n": true, "ctrl+o": true, "ctrl+p": true,
		"ctrl+q": true, "ctrl+r": true, "ctrl+s": true, "ctrl+t": true,
		"ctrl+u": true, "ctrl+v": true, "ctrl+w": true, "ctrl+x": true,
		"ctrl+y": true, "ctrl+z": true,
	}
	if special[key] {
		return false
	}
	// Accept anything else (single chars AND multi-char pastes)
	return len(key) >= 1
}

// SanitizeInput cleans pasted text by removing bracketed paste escape sequences
// and other terminal artifacts. Handles all forms:
//   - \x1b[200~text\x1b[201~  (full sequence)
//   - [text]                   (bubbletea strips the escapes, leaves brackets)
//   - [text                    (only leading bracket)
//   - text]                    (only trailing bracket)
func SanitizeInput(s string) string {
	// Remove full bracketed paste escape sequences
	s = strings.ReplaceAll(s, "\x1b[200~", "")
	s = strings.ReplaceAll(s, "\x1b[201~", "")
	s = strings.ReplaceAll(s, "\x1b", "")

	// Strip newlines
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")

	// Only process longer strings (paste) — single chars like space pass through unchanged
	if len(s) > 1 {
		s = strings.TrimSpace(s)

		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			inner := s[1 : len(s)-1]
			if strings.ContainsAny(inner, "./:@") {
				s = inner
			}
		} else if strings.HasPrefix(s, "[") {
			candidate := s[1:]
			if strings.ContainsAny(candidate, "./:@") {
				s = candidate
			}
		} else if strings.HasSuffix(s, "]") {
			candidate := s[:len(s)-1]
			if strings.ContainsAny(candidate, "./:@") {
				s = candidate
			}
		}

		s = strings.TrimSpace(s)
	}

	return s
}

// FileTagStyle returns the appropriate style for a status code
func FileTagStyle(code string) lipgloss.Style {
	switch code {
	case "A", "?":
		return FileTagNewStyle
	case "D":
		return FileTagDeletedStyle
	case "M":
		return FileTagModifiedStyle
	default:
		return FileTagStagedStyle
	}
}
