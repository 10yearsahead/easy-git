package ui

import (
	"fmt"
	"strings"

	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
	tea "github.com/charmbracelet/bubbletea"
)

// ── Commit View ───────────────────────────────────────────────────────────────

type commitStep int

const (
	commitStepFiles commitStep = iota
	commitStepMessage
	commitStepConfirm
	commitStepDone
)

type CommitModel struct {
	step        commitStep
	files       []git.FileStatus
	selected    []bool
	cursor      int
	msgInput    TextInput
	result      string
	resultIsErr bool
	loaded      bool
}

type commitFilesLoadedMsg struct{ files []git.FileStatus }
type commitDoneMsg struct {
	output string
	isErr  bool
}

func NewCommitModel() CommitModel {
	return CommitModel{
		msgInput: NewTextInput("e.g. Add login button"),
	}
}

func LoadCommitFiles() tea.Cmd {
	return func() tea.Msg {
		s := git.Status()
		seen := map[string]bool{}
		var files []git.FileStatus

		for _, f := range s.Staged {
			seen[f.Name] = true
			files = append(files, f)
		}
		for _, f := range s.Unstaged {
			if !seen[f.Name] {
				seen[f.Name] = true
				files = append(files, f)
			}
		}
		for _, f := range s.Untracked {
			if !seen[f.Name] {
				seen[f.Name] = true
				files = append(files, f)
			}
		}
		return commitFilesLoadedMsg{files: files}
	}
}

func doCommit(files []git.FileStatus, selected []bool, message string) tea.Cmd {
	return func() tea.Msg {
		var toAdd []string
		for i, f := range files {
			if selected[i] {
				toAdd = append(toAdd, f.Name)
			}
		}
		if len(toAdd) > 0 {
			if err := git.Add(toAdd); err != nil {
				return commitDoneMsg{output: err.Error(), isErr: true}
			}
		}
		out, err := git.Commit(message)
		return commitDoneMsg{output: out, isErr: err != nil}
	}
}

func (m CommitModel) Update(msg tea.Msg) (CommitModel, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case commitFilesLoadedMsg:
		m.files = msg.files
		m.selected = make([]bool, len(msg.files))
		for i, f := range m.files {
			if f.Status == "staged" {
				m.selected[i] = true
			}
		}
		m.loaded = true
		return m, nil, false

	case commitDoneMsg:
		m.result = msg.output
		m.resultIsErr = msg.isErr
		m.step = commitStepDone
		return m, nil, false

	case tea.KeyMsg:
		switch m.step {

		case commitStepFiles:
			switch msg.String() {
			case "esc", "q":
				return m, nil, true
			case "up", "k":
				if len(m.files) == 0 {
					break
				}
				total := len(m.files) + 1
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = total - 1
				}
			case "down", "j":
				if len(m.files) == 0 {
					break
				}
				total := len(m.files) + 1
				if m.cursor < total-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case " ":
				if m.cursor == 0 {
					anyUnselected := false
					for _, s := range m.selected {
						if !s {
							anyUnselected = true
							break
						}
					}
					for i := range m.selected {
						m.selected[i] = anyUnselected
					}
				} else {
					idx := m.cursor - 1
					m.selected[idx] = !m.selected[idx]
				}
			case "enter":
				anySelected := false
				for _, s := range m.selected {
					if s {
						anySelected = true
						break
					}
				}
				if anySelected {
					m.step = commitStepMessage
				}
			}

		case commitStepMessage:
			if msg.String() == "esc" {
				m.step = commitStepFiles
				break
			}
			var entered bool
			m.msgInput, entered = m.msgInput.Update(msg)
			if entered {
				if strings.TrimSpace(m.msgInput.Value) == "" {
					break
				}
				m.step = commitStepConfirm
				m.cursor = 0
			}

		case commitStepConfirm:
			switch msg.String() {
			case "esc":
				m.step = commitStepMessage
			case "up", "k":
				m.cursor = 0
			case "down", "j":
				m.cursor = 1
			case "enter":
				if m.cursor == 0 {
					return m, doCommit(m.files, m.selected, m.msgInput.Value), false
				}
				return m, nil, true
			}

		case commitStepDone:
			switch msg.String() {
			case "enter", "esc", "q":
				return m, nil, true
			}
		}
	}
	return m, nil, false
}

func (m CommitModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(i18n.T("commit.title")))
	b.WriteString("\n")

	if !m.loaded && m.step == commitStepFiles {
		b.WriteString(MutedStyle.Render(i18n.T("general.loading")))
		return b.String()
	}

	switch m.step {
	case commitStepFiles:
		b.WriteString(StepStyle.Render(i18n.T("commit.step1")))
		b.WriteString("\n\n")

		if len(m.files) == 0 {
			b.WriteString(SuccessStyle.Render("✨  Nothing to commit!"))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render("  Your repository is clean — no files were modified,\n  added or removed since the last commit."))
			b.WriteString("\n\n")
			b.WriteString(HelpStyle.Render(i18n.T("help.esc_back")))
			break
		}

		if m.cursor == 0 {
			b.WriteString(CursorStyle.Render(" ❯ "))
		} else {
			b.WriteString("   ")
		}
		allSelected := true
		for _, s := range m.selected {
			if !s {
				allSelected = false
				break
			}
		}
		cb := CheckboxUnchecked.Render("○")
		if allSelected && len(m.selected) > 0 {
			cb = CheckboxChecked.Render("●")
		}
		b.WriteString(cb + " " + MutedStyle.Render(i18n.T("commit.select_all")))
		b.WriteString("\n")

		for i, f := range m.files {
			row := i + 1
			if m.cursor == row {
				b.WriteString(CursorStyle.Render(" ❯ "))
			} else {
				b.WriteString("   ")
			}
			cb := CheckboxUnchecked.Render("○")
			if m.selected[i] {
				cb = CheckboxChecked.Render("●")
			}
			icon := FileTagStyle(f.Short).Render(fmt.Sprintf("[%s]", FileIcon(f.Short)))

			extra := ""
			if f.Status == "staged" {
				extra = " " + FileTagStagedStyle.Render(i18n.T("help.staged"))
			}

			b.WriteString(fmt.Sprintf("%s %s %s%s\n", cb, icon, FileNameStyle.Render(f.Name), extra))
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_space_back")))

	case commitStepMessage:
		b.WriteString(StepStyle.Render(i18n.T("commit.step2")))
		b.WriteString("\n\n")
		b.WriteString(m.msgInput.ViewWithLabel(
			i18n.T("commit.msg_prompt"),
			i18n.T("commit.msg_hint"),
		))

	case commitStepConfirm:
		b.WriteString(StepStyle.Render(i18n.T("commit.step3")))
		b.WriteString("\n\n")

		count := 0
		for _, s := range m.selected {
			if s {
				count++
			}
		}

		msgDisplay := m.msgInput.Value
		if strings.TrimSpace(msgDisplay) == "" {
			msgDisplay = MutedStyle.Render("(no message)")
		} else {
			msgDisplay = FileNameStyle.Render("\"" + msgDisplay + "\"")
		}
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  %d file(s)  ·  ", count)))
		b.WriteString(msgDisplay)
		b.WriteString("\n\n")

		options := []string{i18n.T("commit.confirm"), i18n.T("commit.back")}
		for i, opt := range options {
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
				if i == 0 {
					b.WriteString(SuccessStyle.Render(opt))
				} else {
					b.WriteString(ErrorStyle.Render(opt))
				}
			} else {
				b.WriteString("   ")
				b.WriteString(MutedStyle.Render(opt))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_choose")))

	case commitStepDone:
		b.WriteString("\n")
		if m.resultIsErr {
			b.WriteString(ErrorStyle.Render("❌ " + i18n.T("pushpull.error")))
			b.WriteString("\n")
			b.WriteString(MutedStyle.Render(m.result))
		} else {
			b.WriteString(SuccessStyle.Render(i18n.T("commit.success")))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render(m.result))
		}
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_back_menu")))
	}

	return b.String()
}
