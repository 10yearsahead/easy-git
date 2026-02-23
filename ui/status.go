package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
)

// ── Status View ───────────────────────────────────────────────────────────────

type StatusModel struct {
	status git.RepoStatus
	loaded bool
}

type statusLoadedMsg struct{ status git.RepoStatus }

func NewStatusModel() StatusModel {
	return StatusModel{}
}

func LoadStatus() tea.Cmd {
	return func() tea.Msg {
		return statusLoadedMsg{status: git.Status()}
	}
}

func (m StatusModel) Update(msg tea.Msg) (StatusModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case statusLoadedMsg:
		m.status = msg.status
		m.loaded = true
		return m, nil, false

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "backspace":
			return m, nil, true // signal: go back
		}
	}
	return m, nil, false
}

func (m StatusModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(i18n.T("status.title")))
	b.WriteString("\n")

	if !m.loaded {
		b.WriteString(MutedStyle.Render(i18n.T("general.loading")))
		return b.String()
	}

	if !m.status.IsRepo {
		b.WriteString("\n")
		b.WriteString(WarningStyle.Render(i18n.T("status.no_repo")))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("status.back")))
		return b.String()
	}

	// Branch badge
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(i18n.T("status.branch")+" "))
	b.WriteString(BranchBadgeStyle.Render(" 🌿 "+m.status.Branch+" "))
	b.WriteString("\n")

	hasAny := len(m.status.Staged) > 0 || len(m.status.Unstaged) > 0 || len(m.status.Untracked) > 0

	if !hasAny {
		b.WriteString("\n")
		b.WriteString(SuccessStyle.Render(i18n.T("status.clean")))
		b.WriteString("\n")
	} else {
		// Staged files
		if len(m.status.Staged) > 0 {
			b.WriteString(renderFileSection(
				i18n.T("status.staged"),
				m.status.Staged,
				true,
			))
		}

		// Unstaged files
		if len(m.status.Unstaged) > 0 {
			b.WriteString(renderFileSection(
				i18n.T("status.unstaged"),
				m.status.Unstaged,
				false,
			))
		}

		// Untracked files
		if len(m.status.Untracked) > 0 {
			b.WriteString(renderFileSection(
				i18n.T("status.untracked"),
				m.status.Untracked,
				false,
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render(i18n.T("status.back") + "   " + i18n.T("help.esc_back")))

	return b.String()
}

func renderFileSection(title string, files []git.FileStatus, staged bool) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(SectionTitleStyle.Render(title))
	b.WriteString("\n")

	for _, f := range files {
		icon := FileIcon(f.Short)
		tagStyle := FileTagStyle(f.Short)

		tag := tagStyle.Render(fmt.Sprintf("[%s]", icon))
		name := FileNameStyle.Render(f.Name)

		b.WriteString(fmt.Sprintf("  %s %s\n", tag, name))
	}

	return b.String()
}
