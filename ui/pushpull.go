package ui

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
)

// ── Push/Pull View ────────────────────────────────────────────────────────────

type pushPullState int

const (
	ppStateMenu pushPullState = iota
	ppStateLoading
	ppStateDone
)

type PushPullModel struct {
	cursor int
	state  pushPullState
	isPush bool
	result string
	isErr  bool
}

type pushPullDoneMsg struct {
	output string
	isErr  bool
}

func NewPushPullModel() PushPullModel {
	return PushPullModel{}
}

// execPush suspends the TUI and runs git push interactively in the real terminal
func execPush() tea.Cmd {
	branch := git.CurrentBranch()
	c := exec.Command("git", "push", "origin", branch)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return pushPullDoneMsg{output: err.Error(), isErr: true}
		}
		return pushPullDoneMsg{output: "Push completed successfully!", isErr: false}
	})
}

// execPull suspends the TUI and runs git pull interactively in the real terminal
func execPull() tea.Cmd {
	c := exec.Command("git", "pull")
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return pushPullDoneMsg{output: err.Error(), isErr: true}
		}
		return pushPullDoneMsg{output: "Pull completed successfully!", isErr: false}
	})
}

func (m PushPullModel) Update(msg tea.Msg) (PushPullModel, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case pushPullDoneMsg:
		m.result = msg.output
		m.isErr = msg.isErr
		m.state = ppStateDone
		return m, nil, false

	case tea.KeyMsg:
		switch m.state {
		case ppStateMenu:
			switch msg.String() {
			case "esc", "q":
				return m, nil, true
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = 1
				}
			case "down", "j":
				if m.cursor < 1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case "enter":
				m.isPush = m.cursor == 0
				m.state = ppStateLoading
				if m.isPush {
					return m, execPush(), false
				}
				return m, execPull(), false
			}
		case ppStateDone:
			switch msg.String() {
			case "enter", "esc", "q":
				return m, nil, true
			}
		}
	}
	return m, nil, false
}

func (m PushPullModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(i18n.T("pushpull.title")))
	b.WriteString("\n\n")

	switch m.state {
	case ppStateMenu:
		options := []string{
			i18n.T("pushpull.push"),
			i18n.T("pushpull.pull"),
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

	case ppStateLoading:
		if m.isPush {
			b.WriteString(MutedStyle.Render("⏳  " + i18n.T("pushpull.pushing")))
		} else {
			b.WriteString(MutedStyle.Render("⏳  " + i18n.T("pushpull.pulling")))
		}
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("The terminal will ask for your GitHub username and token..."))

	case ppStateDone:
		b.WriteString("\n")
		if m.isErr {
			b.WriteString(ErrorStyle.Render(i18n.T("pushpull.error")))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render(m.result))
			b.WriteString("\n\n")
			b.WriteString(InputHintStyle.Render("💡 Tip: use your GitHub token as the password, not your account password."))
		} else {
			b.WriteString(SuccessStyle.Render(i18n.T("pushpull.success")))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render(m.result))
		}
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_back_menu")))
	}

	return b.String()
}
