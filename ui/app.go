package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Root App ──────────────────────────────────────────────────────────────────

type App struct {
	scene      Scene
	menu       MenuModel
	init       InitModel
	status     StatusModel
	commit     CommitModel
	pushPull   PushPullModel
	branches   BranchesModel
	config     ConfigModel
	width      int
	height     int
}

func NewApp() App {
	return App{
		scene:    SceneMenu,
		menu:     NewMenuModel(),
		status:   NewStatusModel(),
		commit:   NewCommitModel(),
		pushPull: NewPushPullModel(),
		branches: NewBranchesModel(),
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	}

	switch a.scene {
	case SceneMenu:
		var cmd tea.Cmd
		var nextScene Scene
		a.menu, cmd, nextScene = a.menu.Update(msg)
		if nextScene != SceneMenu {
			a.scene = nextScene
			// Initialize the target scene
			switch nextScene {
			case SceneInit:
				a.init = NewInitModel()
				return a, checkInitConfig()
			case SceneStatus:
				a.status = NewStatusModel()
				return a, LoadStatus()
			case SceneCommit:
				a.commit = NewCommitModel()
				return a, LoadCommitFiles()
			case ScenePushPull:
				a.pushPull = NewPushPullModel()
			case SceneBranches:
				a.branches = NewBranchesModel()
				return a, LoadBranches()
			case SceneConfig:
				a.config = NewConfigModel()
				return a, loadConfig()
			}
		}
		return a, cmd

	case SceneInit:
		var cmd tea.Cmd
		var goBack bool
		a.init, cmd, goBack = a.init.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd

	case SceneStatus:
		var cmd tea.Cmd
		var goBack bool
		a.status, cmd, goBack = a.status.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd

	case SceneCommit:
		var cmd tea.Cmd
		var goBack bool
		a.commit, cmd, goBack = a.commit.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd

	case ScenePushPull:
		var cmd tea.Cmd
		var goBack bool
		a.pushPull, cmd, goBack = a.pushPull.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd

	case SceneBranches:
		var cmd tea.Cmd
		var goBack bool
		a.branches, cmd, goBack = a.branches.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd

	case SceneConfig:
		var cmd tea.Cmd
		var goBack bool
		a.config, cmd, goBack = a.config.Update(msg)
		if goBack {
			a.scene = SceneMenu
			a.menu = NewMenuModel()
		}
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	var content string

	switch a.scene {
	case SceneMenu:
		content = a.menu.View()
	case SceneInit:
		content = a.init.View()
	case SceneStatus:
		content = a.status.View()
	case SceneCommit:
		content = a.commit.View()
	case ScenePushPull:
		content = a.pushPull.View()
	case SceneBranches:
		content = a.branches.View()
	case SceneConfig:
		content = a.config.View()
	}

	// Wrap in app container with consistent width
	width := a.width
	if width == 0 {
		width = 80
	}

	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 3).
		Render(content)
}
