package ui

import (
	"fmt"
	"strings"

	"github.com/10yearsahead/easy-git/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

// Scene represents which screen is currently shown
type Scene int

const (
	SceneMenu Scene = iota
	SceneInit
	SceneStatus
	SceneCommit
	ScenePushPull
	SceneBranches
	SceneConfig
)

// ── Main Menu ─────────────────────────────────────────────────────────────────

type MenuItem struct {
	Key   string
	Scene Scene
}

var menuItems = []MenuItem{
	{"menu.init", SceneInit},
	{"menu.status", SceneStatus},
	{"menu.commit", SceneCommit},
	{"menu.push_pull", ScenePushPull},
	{"menu.branches", SceneBranches},
	{"menu.config", SceneConfig},
}

type MenuModel struct {
	cursor int
}

func NewMenuModel() MenuModel {
	return MenuModel{cursor: 0}
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd, Scene) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			total := len(menuItems) + 1 // +quit
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = total - 1
			}
		case "down", "j":
			total := len(menuItems) + 1
			if m.cursor < total-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter", " ":
			switch m.cursor {
			case len(menuItems): // quit
				return m, tea.Quit, SceneMenu
			default:
				return m, nil, menuItems[m.cursor].Scene
			}
		case "q":
			return m, tea.Quit, SceneMenu
		}
	}
	return m, nil, SceneMenu
}

func (m MenuModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(i18n.T("menu.title")))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(i18n.T("menu.subtitle")))
	b.WriteString("\n\n")

	// Menu items
	allItems := make([]string, 0, len(menuItems)+2)
	for _, item := range menuItems {
		allItems = append(allItems, i18n.T(item.Key))
	}
	allItems = append(allItems, i18n.T("menu.quit"))

	for i, label := range allItems {
		if i == m.cursor {
			cursor := CursorStyle.Render("❯")
			item := MenuItemSelectedStyle.Render(label)
			b.WriteString(fmt.Sprintf(" %s %s\n", cursor, item))
		} else {
			item := MenuItemStyle.Render(label)
			b.WriteString(fmt.Sprintf("   %s\n", item))
		}
	}

	// Help
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render(i18n.T("menu.navigate")))

	return b.String()
}
