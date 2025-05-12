package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)

type AppViewModel struct {
	screen     string
	planetFile *configs.ZerotierPlanetFile
	planetList list.Model
	actionList list.Model
}

func (m AppViewModel) Init() tea.Cmd {
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "enter" {
			switch m.screen {
			case "list":
				p, ok := m.planetList.SelectedItem().(PlanetItem)
				if ok {
					m.planetFile = p.Planet
					m.screen = "action"
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := planetListStyle.GetFrameSize()
		m.planetList.SetSize(msg.Width-h, msg.Height-v)
		m.actionList.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	switch m.screen {
	case "list":
		m.planetList, cmd = m.planetList.Update(msg)
	case "action":
		m.actionList, cmd = m.actionList.Update(msg)
	}

	return m, cmd
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
		actionList: CreateActionListView(),
	}
	return &m, nil
}
