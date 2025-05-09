package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ActionItem struct {
	Id   string
	Name string
	Desc string
}

type ActionListView struct {
	list   list.Model
	Planet *configs.ZerotierPlanetFile
}

func (m ActionListView) Init() tea.Cmd {
	return nil
}

func (m ActionListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := planetListStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ActionListView) View() string {
	return "\n" + m.list.View()
}

func (i ActionItem) FilterValue() string { return "" }
func (i ActionItem) Title() string       { return i.Name }
func (i ActionItem) Description() string { return i.Desc }

func CreateActionListView(planetFile *configs.ZerotierPlanetFile) *ActionListView {
	actionList := []list.Item{
		ActionItem{Id: "activate", Name: "Activate", Desc: "Activate this planet file"},
		ActionItem{Id: "view", Name: "View info", Desc: "View the info of planet file"},
		ActionItem{Id: "delete", Name: "Delete", Desc: "Delete the planet file"},
	}
	l := list.New(actionList, list.NewDefaultDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.Title = "What do you want?"
	return &ActionListView{
		Planet: planetFile,
		list:   l,
	}
}
