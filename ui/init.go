package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
)

// ── Init View ─────────────────────────────────────────────────────────────────

type initStep int

const (
	initStepCheck       initStep = iota // 1 - check git config (name/email)
	initStepConfigName                  // 1b - enter name
	initStepConfigEmail                 // 1c - enter email
	initStepBranch                      // 2 - choose main branch
	initStepBranchCustom                // 2b - type custom name
	initStepGitignore                   // 3 - create gitignore?
	initStepGitignorePick               // 3b - choose template
	initStepRemote                      // 4 - connect remote?
	initStepRemoteURL                   // 4b - enter URL
	initStepFirstCommit                 // 5 - first commit?
	initStepRunning                     // running operations
	initStepDone                        // done
	initStepAlreadyRepo                 // already a repo
)

// Summary of completed actions
type initSummary struct {
	configSet    bool
	branch       string
	gitignore    bool
	remoteURL    string
	firstCommit  bool
	initOutput   string
	commitOutput string
	remoteOutput string
	errors       []string
}

type InitModel struct {
	step          initStep
	config        git.GitConfig
	inputName     TextInput
	inputEmail    TextInput
	inputBranch   TextInput
	inputRemote   TextInput
	cursor        int
	branchChoice  string // "main", "master", "custom"
	gitignorePick int    // index in templates list + 1 (0 = none)
	wantRemote    bool
	wantCommit    bool
	summary       initSummary
	loaded        bool
}

func NewInitModel() InitModel {
	return InitModel{
		inputName:   NewTextInput("e.g. John Doe"),
		inputEmail:  NewTextInput("e.g. john@email.com"),
		inputBranch: NewTextInput("e.g. develop"),
		inputRemote: NewTextInput("https://github.com/your-user/your-project.git"),
	}
}

// Messages
type initCheckDoneMsg struct{ config git.GitConfig }
type initRunDoneMsg struct{ summary initSummary }

func checkInitConfig() tea.Cmd {
	return func() tea.Msg {
		cfg := git.GetGlobalConfig()
		return initCheckDoneMsg{config: cfg}
	}
}

func runInit(m InitModel) tea.Cmd {
	return func() tea.Msg {
		s := initSummary{
			branch: m.branchChoice,
		}

		// 1. Set config if needed
		if m.config.Name == "" || m.config.Email == "" {
			// Already set inline during wizard steps — retrieve again
		}
		s.configSet = m.config.Name != "" && m.config.Email != ""

		// 2. git init
		out, err := git.Init(m.branchChoice)
		s.initOutput = out
		if err != nil {
			s.errors = append(s.errors, "git init: "+out)
			return initRunDoneMsg{summary: s}
		}

		// 3. .gitignore
		if m.gitignorePick > 0 {
			idx := m.gitignorePick - 1
			if idx < len(git.GitignoreTemplates) {
				tmpl := git.GitignoreTemplates[idx]
				if err := git.WriteGitignore(tmpl.Content); err != nil {
					s.errors = append(s.errors, ".gitignore: "+err.Error())
				} else {
					s.gitignore = true
				}
			}
		}

		// 4. Remote
		if m.wantRemote && m.inputRemote.Value != "" {
			remoteURL := sanitizeURL(m.inputRemote.Value)
			s.remoteURL = remoteURL
			out, err := git.AddRemote(remoteURL)
			s.remoteOutput = out
			if err != nil {
				s.errors = append(s.errors, "remote: "+out)
				s.remoteURL = ""
			}
		}

		// 5. First commit
		if m.wantCommit {
			out, err := git.InitialCommit()
			s.commitOutput = out
			if err != nil {
				s.errors = append(s.errors, "commit: "+out)
			} else {
				s.firstCommit = true
			}
		}

		return initRunDoneMsg{summary: s}
	}
}

func (m InitModel) Update(msg tea.Msg) (InitModel, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case initCheckDoneMsg:
		m.config = msg.config
		m.loaded = true

		// Already a repo?
		if git.IsGitRepo() {
			m.step = initStepAlreadyRepo
			return m, nil, false
		}

		// Config ok?
		if m.config.Name != "" && m.config.Email != "" {
			m.step = initStepBranch
		} else {
			m.step = initStepCheck
		}
		return m, nil, false

	case initRunDoneMsg:
		m.summary = msg.summary
		m.step = initStepDone
		return m, nil, false

	case tea.KeyMsg:
		switch m.step {

		// ── Already a repo ────────────────────────────────────────────────
		case initStepAlreadyRepo:
			if msg.String() == "esc" || msg.String() == "enter" || msg.String() == "q" {
				return m, nil, true
			}

		// ── Step 1: Config missing — ask for name ─────────────────────────
		case initStepCheck:
			switch msg.String() {
			case "esc":
				return m, nil, true
			case "enter":
				m.step = initStepConfigName
				m.inputName = NewTextInput("e.g. John Doe")
			}

		case initStepConfigName:
			if msg.String() == "esc" {
				m.step = initStepCheck
				break
			}
			var entered bool
			m.inputName, entered = m.inputName.Update(msg)
			if entered && strings.TrimSpace(m.inputName.Value) != "" {
				m.config.Name = strings.TrimSpace(m.inputName.Value)
				git.SetGlobalConfig(m.config.Name, m.config.Email)
				m.inputEmail = NewTextInput("e.g. john@email.com")
				m.step = initStepConfigEmail
			}

		case initStepConfigEmail:
			if msg.String() == "esc" {
				m.step = initStepConfigName
				break
			}
			var entered bool
			m.inputEmail, entered = m.inputEmail.Update(msg)
			if entered && strings.TrimSpace(m.inputEmail.Value) != "" {
				m.config.Email = strings.TrimSpace(m.inputEmail.Value)
				git.SetGlobalConfig(m.config.Name, m.config.Email)
				m.step = initStepBranch
				m.cursor = 0
			}

		// ── Step 2: Branch ────────────────────────────────────────────────
		case initStepBranch:
			switch msg.String() {
			case "esc":
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
				case 0:
					m.branchChoice = "main"
					m.step = initStepGitignore
					m.cursor = 0
				case 1:
					m.branchChoice = "master"
					m.step = initStepGitignore
					m.cursor = 0
				case 2:
					m.step = initStepBranchCustom
					m.inputBranch = NewTextInput("e.g. develop")
				}
			}

		case initStepBranchCustom:
			if msg.String() == "esc" {
				m.step = initStepBranch
				break
			}
			var entered bool
			m.inputBranch, entered = m.inputBranch.Update(msg)
			if entered && strings.TrimSpace(m.inputBranch.Value) != "" {
				m.branchChoice = strings.TrimSpace(m.inputBranch.Value)
				m.step = initStepGitignore
				m.cursor = 0
			}

		// ── Step 3: Gitignore ─────────────────────────────────────────────
		case initStepGitignore:
			switch msg.String() {
			case "esc":
				m.step = initStepBranch
				m.cursor = 0
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
				if m.cursor == 0 {
					m.step = initStepGitignorePick
					m.cursor = 0
				} else {
					m.gitignorePick = 0
					m.step = initStepRemote
					m.cursor = 0
				}
			}

		case initStepGitignorePick:
			totalItems := len(git.GitignoreTemplates) + 1 // +1 for "none"
			switch msg.String() {
			case "esc":
				m.step = initStepGitignore
				m.cursor = 0
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = totalItems - 1
				}
			case "down", "j":
				if m.cursor < totalItems-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case "enter":
				if m.cursor == totalItems-1 {
					m.gitignorePick = 0 // none
				} else {
					m.gitignorePick = m.cursor + 1 // 1-indexed
				}
				m.step = initStepRemote
				m.cursor = 0
			}

		// ── Step 4: Remote ────────────────────────────────────────────────
		case initStepRemote:
			switch msg.String() {
			case "esc":
				m.step = initStepGitignore
				m.cursor = 0
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
				if m.cursor == 0 {
					m.wantRemote = true
					m.step = initStepRemoteURL
					m.inputRemote = NewTextInput("https://github.com/your-user/your-project.git")
				} else {
					m.wantRemote = false
					m.step = initStepFirstCommit
					m.cursor = 0
				}
			}

		case initStepRemoteURL:
			if msg.String() == "esc" {
				m.step = initStepRemote
				break
			}
			var entered bool
			m.inputRemote, entered = m.inputRemote.Update(msg)
			if entered && strings.TrimSpace(m.inputRemote.Value) != "" {
				m.step = initStepFirstCommit
				m.cursor = 0
			}

		// ── Step 5: First commit ──────────────────────────────────────────
		case initStepFirstCommit:
			switch msg.String() {
			case "esc":
				m.step = initStepRemote
				m.cursor = 0
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
				m.wantCommit = m.cursor == 0
				m.step = initStepRunning
				return m, runInit(m), false
			}

		// ── Running ───────────────────────────────────────────────────────
		case initStepRunning:
			// waiting — no key response

		// ── Done ──────────────────────────────────────────────────────────
		case initStepDone:
			if msg.String() == "enter" || msg.String() == "esc" || msg.String() == "q" {
				return m, nil, true
			}
		}
	}

	return m, nil, false
}

func (m InitModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(i18n.T("init.title")))
	b.WriteString("\n")

	if !m.loaded {
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render(i18n.T("general.loading")))
		return b.String()
	}

	switch m.step {

	// ── Already a repo ────────────────────────────────────────────────────────
	case initStepAlreadyRepo:
		b.WriteString("\n")
		b.WriteString(WarningStyle.Render(i18n.T("init.already_repo")))
		b.WriteString("\n\n")
		b.WriteString(MutedStyle.Render("This folder already has an active git repository."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.esc_back_menu")))

	// ── Step 1: Config warning ────────────────────────────────────────────────
	case initStepCheck:
		b.WriteString(StepStyle.Render(i18n.T("init.step1")))
		b.WriteString("\n\n")
		b.WriteString(WarningStyle.Render(i18n.T("init.config_missing")))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("  " + i18n.T("init.config_why")))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_configure_cancel")))

	// ── Step 1b: Name ─────────────────────────────────────────────────────────
	case initStepConfigName:
		b.WriteString(StepStyle.Render(i18n.T("init.step1")))
		b.WriteString("\n\n")
		b.WriteString(m.inputName.ViewWithLabel(i18n.T("init.name_prompt"), ""))

	// ── Step 1c: Email ────────────────────────────────────────────────────────
	case initStepConfigEmail:
		b.WriteString(StepStyle.Render(i18n.T("init.step1")))
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("✅  " + i18n.T("init.name_prompt") + " " + m.config.Name))
		b.WriteString("\n\n")
		b.WriteString(m.inputEmail.ViewWithLabel(i18n.T("init.email_prompt"), ""))

	// ── Step 2: Branch ────────────────────────────────────────────────────────
	case initStepBranch:
		b.WriteString(StepStyle.Render(i18n.T("init.step2")))
		b.WriteString("\n\n")
		b.WriteString(InputPromptStyle.Render(i18n.T("init.branch_prompt")))
		b.WriteString("\n")
		b.WriteString(InputHintStyle.Render("  " + i18n.T("init.branch_hint")))
		b.WriteString("\n\n")

		opts := []string{
			i18n.T("init.branch_main"),
			i18n.T("init.branch_master"),
			i18n.T("init.branch_custom"),
		}
		for i, opt := range opts {
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
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_confirm_cancel")))

	// ── Step 2b: Custom branch ────────────────────────────────────────────────
	case initStepBranchCustom:
		b.WriteString(StepStyle.Render(i18n.T("init.step2")))
		b.WriteString("\n\n")
		b.WriteString(m.inputBranch.ViewWithLabel(i18n.T("init.branch_custom_type"), ""))

	// ── Step 3: Gitignore? ────────────────────────────────────────────────────
	case initStepGitignore:
		b.WriteString(StepStyle.Render(i18n.T("init.step3")))
		b.WriteString("\n\n")
		b.WriteString(InputPromptStyle.Render(i18n.T("init.gitignore_title")))
		b.WriteString("\n")
		b.WriteString(InputHintStyle.Render("  " + i18n.T("init.gitignore_why")))
		b.WriteString("\n\n")

		if git.GitignoreExists() {
			b.WriteString(WarningStyle.Render("  " + i18n.T("init.gitignore_exists")))
			b.WriteString("\n\n")
		}

		opts := []string{i18n.T("init.yes"), i18n.T("init.no")}
		for i, opt := range opts {
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

	// ── Step 3b: Choose template ──────────────────────────────────────────────
	case initStepGitignorePick:
		b.WriteString(StepStyle.Render(i18n.T("init.step3")))
		b.WriteString("\n\n")
		b.WriteString(InputPromptStyle.Render(i18n.T("init.gitignore_choose")))
		b.WriteString("\n\n")

		for i, tmpl := range git.GitignoreTemplates {
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
				b.WriteString(MenuItemSelectedStyle.Render(tmpl.Label))
			} else {
				b.WriteString("   ")
				b.WriteString(MenuItemStyle.Render(tmpl.Label))
			}
			b.WriteString("\n")
		}
		// "None" option
		noneIdx := len(git.GitignoreTemplates)
		if m.cursor == noneIdx {
			b.WriteString(CursorStyle.Render(" ❯ "))
			b.WriteString(MutedStyle.Render(i18n.T("init.gitignore_none")))
		} else {
			b.WriteString("   ")
			b.WriteString(MutedStyle.Render(i18n.T("init.gitignore_none")))
		}
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_select_back")))

	// ── Step 4: Remote? ───────────────────────────────────────────────────────
	case initStepRemote:
		b.WriteString(StepStyle.Render(i18n.T("init.step4")))
		b.WriteString("\n\n")
		b.WriteString(InputPromptStyle.Render(i18n.T("init.remote_title")))
		b.WriteString("\n")
		b.WriteString(InputHintStyle.Render("  " + i18n.T("init.remote_why")))
		b.WriteString("\n\n")

		opts := []string{i18n.T("init.remote_yes"), i18n.T("init.remote_skip")}
		for i, opt := range opts {
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

	// ── Step 4b: Remote URL ───────────────────────────────────────────────────
	case initStepRemoteURL:
		b.WriteString(StepStyle.Render(i18n.T("init.step4")))
		b.WriteString("\n\n")
		b.WriteString(m.inputRemote.ViewWithLabel(i18n.T("init.remote_url"), i18n.T("init.remote_hint")))

	// ── Step 5: First commit? ─────────────────────────────────────────────────
	case initStepFirstCommit:
		b.WriteString(StepStyle.Render(i18n.T("init.step5")))
		b.WriteString("\n\n")
		b.WriteString(InputPromptStyle.Render(i18n.T("init.first_commit_title")))
		b.WriteString("\n")
		b.WriteString(InputHintStyle.Render("  " + i18n.T("init.first_commit_why")))
		b.WriteString("\n\n")

		opts := []string{
			i18n.T("init.first_commit_yes"),
			i18n.T("init.first_commit_no"),
		}
		for i, opt := range opts {
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
				if i == 0 {
					b.WriteString(SuccessStyle.Render(opt))
				} else {
					b.WriteString(MutedStyle.Render(opt))
				}
			} else {
				b.WriteString("   ")
				b.WriteString(MenuItemStyle.Render(opt))
			}
			b.WriteString("\n")
		}

		// Preview of what will be executed
		b.WriteString("\n")
		b.WriteString(SectionTitleStyle.Render("  Summary of what will be executed:"))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  · git init -b %s", m.branchChoice)))
		b.WriteString("\n")
		if m.gitignorePick > 0 {
			idx := m.gitignorePick - 1
			if idx < len(git.GitignoreTemplates) {
				b.WriteString(MutedStyle.Render("  · create .gitignore for " + git.GitignoreTemplates[idx].Label))
				b.WriteString("\n")
			}
		}
		if m.wantRemote && m.inputRemote.Value != "" {
			b.WriteString(MutedStyle.Render("  · git remote add origin " + m.inputRemote.Value))
			b.WriteString("\n")
		}
		if m.cursor == 0 {
			b.WriteString(MutedStyle.Render(`  · git add -A && git commit -m "🎉 Initial commit"`))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_confirm_back")))

	// ── Running ───────────────────────────────────────────────────────────────
	case initStepRunning:
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("⏳  Initializing repository..."))

	// ── Done ──────────────────────────────────────────────────────────────────
	case initStepDone:
		s := m.summary
		b.WriteString("\n")
		b.WriteString(SuccessStyle.Render(i18n.T("init.done_title")))
		b.WriteString("\n\n")
		b.WriteString(SectionTitleStyle.Render("  " + i18n.T("init.done_summary")))
		b.WriteString("\n\n")

		b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_init")))
		b.WriteString("\n")

		if s.configSet {
			b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_config")))
			b.WriteString(MutedStyle.Render(fmt.Sprintf("  (%s <%s>)", m.config.Name, m.config.Email)))
			b.WriteString("\n")
		}

		b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_branch")))
		b.WriteString(FileNameStyle.Render("  " + s.branch))
		b.WriteString("\n")

		if s.gitignore {
			b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_gitignore")))
			b.WriteString("\n")
		}

		if s.remoteURL != "" {
			b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_remote")))
			b.WriteString(FileNameStyle.Render("  " + s.remoteURL))
			b.WriteString("\n")
		}

		if s.firstCommit {
			b.WriteString(SuccessStyle.Render("  " + i18n.T("init.done_commit")))
			b.WriteString("\n")
		}

		// Errors
		if len(s.errors) > 0 {
			b.WriteString("\n")
			b.WriteString(ErrorStyle.Render("  ⚠️  Some errors occurred:"))
			b.WriteString("\n")
			for _, e := range s.errors {
				b.WriteString(ErrorStyle.Render("  · " + e))
				b.WriteString("\n")
			}
		}

		// Next steps
		b.WriteString("\n")
		b.WriteString(SectionTitleStyle.Render("  " + i18n.T("init.done_next")))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("  " + i18n.T("init.done_next1")))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("  " + i18n.T("init.done_next2")))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_back_menu")))
	}

	return b.String()
}

// sanitizeURL strips terminal escape sequences and whitespace from a pasted URL.
// Handles bracketed paste mode artifacts like '[https://...' becoming 'https://...'
func sanitizeURL(url string) string {
	// Remove escape sequences
	url = strings.ReplaceAll(url, "\x1b[200~", "")
	url = strings.ReplaceAll(url, "\x1b[201~", "")
	url = strings.ReplaceAll(url, "\x1b", "")
	// Remove stray brackets that terminals sometimes inject at the start
	url = strings.TrimLeft(url, "[")
	// Remove whitespace and newlines
	url = strings.ReplaceAll(url, "\n", "")
	url = strings.ReplaceAll(url, "\r", "")
	url = strings.TrimSpace(url)
	return url
}

// renderInput renders a labeled text input box
func renderInput(label, value, placeholder string) string {
	var b strings.Builder
	b.WriteString(InputPromptStyle.Render(label))
	b.WriteString("\n")

	display := value
	if display == "" {
		display = MutedStyle.Render(placeholder)
	} else {
		display = FileNameStyle.Render(value) + CursorStyle.Render("▌")
	}
	b.WriteString(PanelStyle.Render(" " + display))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render(i18n.T("help.enter_confirm_back")))
	return b.String()
}
