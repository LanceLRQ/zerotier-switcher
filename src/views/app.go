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
	planetList list.Model
}

func (m AppViewModel) Init() tea.Cmd {
	// Return initial command(s) or nil
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	if m.screen == "list" {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			h, v := planetListStyle.GetFrameSize()
			m.planetList.SetSize(msg.Width-h, msg.Height-v)
		}
		var cmd tea.Cmd
		m.planetList, cmd = m.planetList.Update(msg)
		return m, cmd
	}
	return m, tea.Quit
}

func (m AppViewModel) View() string {
	// Return a string representation of the UI
	if m.screen == "list" {
		return "\n" + m.planetList.View()
	}
	return ""
}

func CreateAppView(cfg *configs.ZerotierSwitcherProfile) (*AppViewModel, error) {
	planetListItems := make([]list.Item, len(cfg.Planets)+1)
	for i := range cfg.Planets {
		planetListItems[i] = PlanetItem{cfg.Planets[i]}
	}
	planetListItems[len(cfg.Planets)] = PlanetItem{
		configs.ZerotierPlanetFile{
			Remark:       "+ Add new",
			RootEndpoint: "select a zerotier planet file",
		},
	}
	m := AppViewModel{
		screen:     "list",
		planetList: list.New(planetListItems, list.NewDefaultDelegate(), 0, 0),
	}
	m.planetList.Title = "Planet List"
	return &m, nil
}
