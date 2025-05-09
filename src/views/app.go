package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)

type AppViewModel struct {
	screen     string
	planetList *PlanetListView
	actionList *ActionListView
}

func (m AppViewModel) Init() tea.Cmd {
	if m.screen == "list" {
		return m.planetList.Init()
	} else if m.screen == "action" {
		return m.actionList.Init()
	}
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.screen == "list" {
		return m.planetList.Update(msg)
	} else if m.screen == "action" {
		return m.actionList.Update(msg)
	}
	return m, tea.Quit
}

func (m AppViewModel) View() string {
	// Return a string representation of the UI
	if m.screen == "list" {
		return m.planetList.View()
	} else if m.screen == "action" {
		return m.actionList.View()
	}
	return ""
}

func CreateAppView(cfg *configs.ZerotierSwitcherProfile) (*AppViewModel, error) {
	m := AppViewModel{
		screen:     "list",
		planetList: CreatePlanetListView(cfg),
		actionList: CreateActionListView(nil),
	}
	m.planetList.onSelect = func(file *configs.ZerotierPlanetFile) error {
		m.screen = "action"
		m.actionList.Planet = file
		return nil
	}
	return &m, nil
}
