package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/10yearsahead/easy-git/git"
	"github.com/10yearsahead/easy-git/i18n"
)

// ── Config View ───────────────────────────────────────────────────────────────

type configScene int

const (
	configSceneMenu configScene = iota
	configSceneSaveName
	configSceneSaveEmail
	configSceneDone
)

type ConfigModel struct {
	scene      configScene
	cursor     int
	nameInput  TextInput
	emailInput TextInput
	config     git.GitConfig
	result     string
	isErr      bool
	loaded     bool
}

type configLoadedMsg struct{ config git.GitConfig }
type configDoneMsg struct {
	result string
	isErr  bool
}

func NewConfigModel() ConfigModel {
	return ConfigModel{
		nameInput:  NewTextInput("e.g. John Doe"),
		emailInput: NewTextInput("e.g. john@email.com"),
	}
}

func loadConfig() tea.Cmd {
	return func() tea.Msg {
		return configLoadedMsg{config: git.GetGlobalConfig()}
	}
}

func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case configLoadedMsg:
		m.config = msg.config
		m.loaded = true
		return m, nil, false

	case configDoneMsg:
		m.result = msg.result
		m.isErr = msg.isErr
		m.scene = configSceneDone
		return m, nil, false

	case tea.KeyMsg:
		switch m.scene {

		case configSceneMenu:
			switch msg.String() {
			case "esc", "q":
				return m, nil, true
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = 4
				}
			case "down", "j":
				if m.cursor < 4 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			case "enter":
				switch m.cursor {
				case 0: // Edit name
					m.scene = configSceneSaveName
					m.nameInput = NewTextInput("e.g. John Doe")
					if m.config.Name != "" {
						m.nameInput = m.nameInput.SetValue(m.config.Name)
					}
				case 1: // Edit email
					m.scene = configSceneSaveEmail
					m.emailInput = NewTextInput("e.g. john@email.com")
					if m.config.Email != "" {
						m.emailInput = m.emailInput.SetValue(m.config.Email)
					}
				case 2: // Save token (enable credential helper)
					err := git.SaveCredentialHelper()
					if err != nil {
						return m, func() tea.Msg {
							return configDoneMsg{result: err.Error(), isErr: true}
						}, false
					}
					return m, func() tea.Msg {
						return configDoneMsg{
							result: "✅  Credentials will be saved automatically after the next push!\n\n  Just push once and enter your username + token one last time.\n  You won't need to type them again.",
							isErr:  false,
						}
					}, false
				case 3: // Clear credentials (token)
					git.ClearCredentials()
					return m, func() tea.Msg {
						return configDoneMsg{
							result: "✅  Credentials removed!\n\n  GitHub will ask for your username and token again on the next push.",
							isErr:  false,
						}
					}, false
				case 4: // Clear everything
					git.ClearAllConfig()
					return m, func() tea.Msg {
						return configDoneMsg{
							result: "✅  All cleared!\n\n  Name, email and credentials have been removed.\n  Configure them again whenever you're ready.",
							isErr:  false,
						}
					}, false
				}
			}

		case configSceneSaveName:
			if msg.String() == "esc" {
				m.scene = configSceneMenu
				break
			}
			var entered bool
			m.nameInput, entered = m.nameInput.Update(msg)
			if entered {
				name := strings.TrimSpace(m.nameInput.Value)
				if name != "" {
					if err := git.SetGlobalConfig(name, m.config.Email); err != nil {
						return m, func() tea.Msg {
							return configDoneMsg{result: err.Error(), isErr: true}
						}, false
					}
					m.config.Name = name
					return m, func() tea.Msg {
						return configDoneMsg{result: "✅  Name saved: " + name, isErr: false}
					}, false
				}
			}

		case configSceneSaveEmail:
			if msg.String() == "esc" {
				m.scene = configSceneMenu
				break
			}
			var entered bool
			m.emailInput, entered = m.emailInput.Update(msg)
			if entered {
				email := strings.TrimSpace(m.emailInput.Value)
				if email != "" {
					if err := git.SetGlobalConfig(m.config.Name, email); err != nil {
						return m, func() tea.Msg {
							return configDoneMsg{result: err.Error(), isErr: true}
						}, false
					}
					m.config.Email = email
					return m, func() tea.Msg {
						return configDoneMsg{result: "✅  Email saved: " + email, isErr: false}
					}, false
				}
			}

		case configSceneDone:
			switch msg.String() {
			case "enter", "esc", "q":
				// Reload config and go back to menu
				m.scene = configSceneMenu
				m.cursor = 0
				m.loaded = false
				return m, loadConfig(), false
			}
		}
	}
	return m, nil, false
}

func (m ConfigModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚙️   Settings"))
	b.WriteString("\n")

	if !m.loaded {
		b.WriteString(MutedStyle.Render("Loading..."))
		return b.String()
	}

	switch m.scene {
	case configSceneMenu:
		// Show current config
		b.WriteString("\n")
		b.WriteString(SectionTitleStyle.Render("  Current configuration:"))
		b.WriteString("\n")

		if m.config.Name != "" {
			b.WriteString(MutedStyle.Render("  Name   "))
			b.WriteString(FileNameStyle.Render(m.config.Name))
		} else {
			b.WriteString(MutedStyle.Render("  Name   "))
			b.WriteString(WarningStyle.Render("not configured"))
		}
		b.WriteString("\n")

		if m.config.Email != "" {
			b.WriteString(MutedStyle.Render("  Email  "))
			b.WriteString(FileNameStyle.Render(m.config.Email))
		} else {
			b.WriteString(MutedStyle.Render("  Email  "))
			b.WriteString(WarningStyle.Render("not configured"))
		}
		b.WriteString("\n")

		helper := git.GetCredentialHelper()
		hasCreds := git.HasStoredCredentials()
		b.WriteString(MutedStyle.Render("  Token  "))
		if hasCreds {
			b.WriteString(SuccessStyle.Render("saved ✓"))
		} else if helper != "" {
			b.WriteString(WarningStyle.Render(i18n.T("config.helper_active")))
		} else {
			b.WriteString(WarningStyle.Render("not saved"))
		}
		b.WriteString("\n\n")

		// Menu options
		options := []struct {
			label string
			desc  string
		}{
			{"✏️   Edit name", ""},
			{"✏️   Edit email", ""},
			{"💾  Save token automatically", "never type it again"},
			{"🗑️   Remove saved token", "next push will ask again"},
			{"⚠️   Clear everything", "removes name, email and token"},
		}

		for i, opt := range options {
			if m.cursor == i {
				b.WriteString(CursorStyle.Render(" ❯ "))
				b.WriteString(MenuItemSelectedStyle.Render(opt.label))
				if opt.desc != "" {
					b.WriteString(MutedStyle.Render("  — " + opt.desc))
				}
			} else {
				b.WriteString("   ")
				if i == 4 { // clear everything — highlight in red
					b.WriteString(ErrorStyle.Render(opt.label))
				} else {
					b.WriteString(MenuItemStyle.Render(opt.label))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.navigate_confirm_back")))

	case configSceneSaveName:
		b.WriteString(StepStyle.Render("Edit name"))
		b.WriteString("\n\n")
		b.WriteString(m.nameInput.ViewWithLabel("Your name:", ""))

	case configSceneSaveEmail:
		b.WriteString(StepStyle.Render("Edit email"))
		b.WriteString("\n\n")
		b.WriteString(m.emailInput.ViewWithLabel("Your email:", ""))

	case configSceneDone:
		b.WriteString("\n")
		if m.isErr {
			b.WriteString(ErrorStyle.Render("❌  Error:"))
			b.WriteString("\n\n")
			b.WriteString(MutedStyle.Render("  " + m.result))
		} else {
			b.WriteString(SuccessStyle.Render(m.result))
		}
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render(i18n.T("help.enter_continue")))
	}

	return b.String()
}
