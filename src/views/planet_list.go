package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/charmbracelet/bubbles/list"
)

type PlanetItem struct {
	Planet *configs.ZerotierPlanetFile
	Id     string
	Name   string
	Desc   string
}

func (i PlanetItem) FilterValue() string { return i.Name }
func (i PlanetItem) Title() string       { return i.Name }
func (i PlanetItem) Description() string { return i.Desc }

func CreatePlanetListView(cfg *configs.ZerotierSwitcherProfile) list.Model {
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
	return l
}
