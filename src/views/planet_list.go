package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type PlanetItem struct {
	Planet *configs.ZerotierPlanetFile
	Id     string
	Name   string
	Desc   string
}

type PlanetListView struct {
	list     list.Model
	onSelect func(*configs.ZerotierPlanetFile) error
}

func (m PlanetListView) Init() tea.Cmd {
	return nil
}

func (m PlanetListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.onSelect != nil {
				sel, ok := m.list.SelectedItem().(PlanetItem)
				if ok {
					_ = m.onSelect(sel.Planet)
				}
			}
			return m, tea.Suspend
		}
	case tea.WindowSizeMsg:
		h, v := planetListStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m PlanetListView) View() string {
	return "\n" + m.list.View()
}

func (i PlanetItem) FilterValue() string { return i.Name }
func (i PlanetItem) Title() string       { return i.Name }
func (i PlanetItem) Description() string { return i.Desc }

func CreatePlanetListView(cfg *configs.ZerotierSwitcherProfile) *PlanetListView {
	planetListItems := make([]list.Item, len(cfg.Planets)+1)
	for i := range cfg.Planets {
		planetListItems[i] = PlanetItem{
			Planet: &cfg.Planets[i],
			Id:     cfg.Planets[i].Hash,
			Name:   cfg.Planets[i].Remark,
			Desc:   cfg.Planets[i].RootEndpoint,
		}
	}
	planetListItems[len(cfg.Planets)] = PlanetItem{
		Id:   "add",
		Name: "+ Add new",
		Desc: "select a zerotier planet file",
	}
	l := list.New(planetListItems, list.NewDefaultDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.Title = "Planet List"
	return &PlanetListView{
		list: l,
	}
}
