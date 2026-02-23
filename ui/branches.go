package ui

import (
	"fmt"
	"strings"

	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

// ── Branches View ─────────────────────────────────────────────────────────────

type branchScene int

const (
	branchSceneMenu branchScene = iota
	branchSceneSwitch
	branchSceneCreate
	branchSceneDeleteConfirm
	branchSceneDone
)

type BranchesModel struct {
	scene         branchScene
	branches      []string
	currentBranch string
	cursor        int
	branchInput   TextInput
	deleteTarget  string
	result        string
	isErr         bool
	loaded        bool
}

type branchesLoadedMsg struct {
	branches []string
	current  string
}

type branchDoneMsg struct {
	output string
	isErr  bool
}

func NewBranchesModel() BranchesModel {
	return BranchesModel{
		branchInput: NewTextInput("e.g. my-feature"),
	}
}

func LoadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, _ := git.Branches()
		current := git.CurrentBranch()
		return branchesLoadedMsg{branches: branches, current: current}
	}
}

func doSwitchBranch(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := git.SwitchBranch(name)
		return branchDoneMsg{output: out, isErr: err != nil}
	}
}

func doCreateBranch(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := git.CreateBranch(name)
		return branchDoneMsg{output: out, isErr: err != nil}
	}
}

func doDeleteBranch(name string) tea.Cmd {
	return func() tea.Msg {
		out, err := git.DeleteBranch(name)
		return branchDoneMsg{output: out, isErr: err != nil}
	}
}

func (m BranchesModel) Update(msg tea.Msg) (BranchesModel, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case branchesLoadedMsg:
		m.branches = msg.branches
		m.currentBranch = msg.current
		m.loaded = true
		return m, nil, false

	case branchDoneMsg:
		m.result = msg.output
		m.isErr = msg.isErr
		m.scene = branchSceneDone
		return m, nil, false

	case tea.KeyMsg:
		switch m.scene {
		case branchSceneMenu:
			switch msg.String() {
			case "esc", "q":
				return m, nil, true
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = 2
				}
			case "down", "j":
				if m.cursor < 2 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case "enter":
				switch m.cursor {
				case 0: // Switch
					m.scene = branchSceneSwitch
					m.cursor = 0
				case 1: // Create
					m.scene = branchSceneCreate
					m.branchInput = NewTextInput("e.g. my-feature")
				case 2: // Delete
					m.scene = branchSceneDeleteConfirm
					m.deleteTarget = ""
					m.cursor = 0
				}
			}

		case branchSceneSwitch:
			switch msg.String() {
			case "esc":
				m.scene = branchSceneMenu
				m.cursor = 0
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "enter":
				if m.cursor < len(m.branches) {
					return m, doSwitchBranch(m.branches[m.cursor]), false
				}
			}

		case branchSceneCreate:
			if msg.String() == "esc" {
				m.scene = branchSceneMenu
				break
			}
			var entered bool
			m.branchInput, entered = m.branchInput.Update(msg)
			if entered && strings.TrimSpace(m.branchInput.Value) != "" {
				return m, doCreateBranch(m.branchInput.Value), false
			}

		case branchSceneDeleteConfirm:
			switch msg.String() {
			case "esc":
				m.scene = branchSceneMenu
				m.cursor = 0
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "enter":
				if m.cursor < len(m.branches) {
					target := m.branches[m.cursor]
					if target != m.currentBranch {
						return m, doDeleteBranch(target), false
					}
				}
			}

		case branchSceneDone:
			switch msg.String() {
			case "enter", "esc", "q":
				// Reload branches and go back to menu
				m.scene = branchSceneMenu
				m.cursor = 0
				m.loaded = false
				return m, LoadBranches(), false
			}
		}
	}
	return m, nil, false
}

func (m BranchesModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(i18n.T("branches.title")))
	b.WriteString("\n")

	if !m.loaded {
		b.WriteString(MutedStyle.Render(i18n.T("general.loading")))
		return b.String()
	}

	// Show current branch
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(i18n.T("branches.current") + " "))
	b.WriteString(BranchBadgeStyle.Render(" 🌿 " + m.currentBranch + " "))
	b.WriteString("\n")

	switch m.scene {
	case branchSceneMenu:
		b.WriteString("\n")
		options := []string{
			i18n.T("branches.switch"),
			i18n.T("branches.create"),
			i18n.T("branches.delete"),
		}
		for i, opt := range options {
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
				b.WriteString(MenuItemSelectedStyle.Render(opt))
			} else {
				b.WriteString("   ")
				b.WriteString(MenuItemStyle.Render(opt))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_confirm_back")))

	case branchSceneSwitch:
		b.WriteString(SectionTitleStyle.Render(i18n.T("branches.switch")))
		b.WriteString("\n\n")
		for i, br := range m.branches {
			current := br == m.currentBranch
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
			} else {
				b.WriteString("   ")
			}
			if current {
				b.WriteString(BranchBadgeStyle.Render(br))
			} else {
				b.WriteString(FileNameStyle.Render(br))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_switch_back")))

	case branchSceneCreate:
		b.WriteString(SectionTitleStyle.Render(i18n.T("branches.create")))
		b.WriteString("\n\n")
		b.WriteString(m.branchInput.ViewWithLabel(i18n.T("branches.name"), ""))

	case branchSceneDeleteConfirm:
		b.WriteString(SectionTitleStyle.Render(i18n.T("branches.delete")))
		b.WriteString("\n\n")
		b.WriteString(WarningStyle.Render("  " + i18n.T("branches.choose_delete")))
		b.WriteString("\n\n")
		for i, br := range m.branches {
			isCurrent := br == m.currentBranch
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
			} else {
				b.WriteString("   ")
			}
			if isCurrent {
				b.WriteString(MutedStyle.Render(br + " (" + i18n.T("branches.cannot_delete") + ")"))
			} else {
				b.WriteString(ErrorStyle.Render(br))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_delete_back")))

	case branchSceneDone:
		b.WriteString("\n")
		if m.isErr {
			b.WriteString(ErrorStyle.Render(i18n.T("branches.error")))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render(m.result))
		} else {
			b.WriteString(SuccessStyle.Render(i18n.T("branches.success")))
			if m.result != "" {
				b.WriteString("\n\n")
				b.WriteString(MutedStyle.Render(m.result))
			}
		}
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_continue")))
	}

	_ = fmt.Sprintf // avoid unused import
	return b.String()
}
