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
	l := list.New(RenderPlanetListItem(cfg.Planets), list.NewDefaultDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.Title = "Planet List"
	return l
}

func RenderPlanetListItem(planets []configs.ZerotierPlanetFile) []list.Item {
	planetListItems := make([]list.Item, len(planets)+1)
	for i := range planets {
		planetListItems[i] = PlanetItem{
			Planet: &planets[i],
			Id:     planets[i].Hash,
			Name:   planets[i].Remark,
			Desc:   planets[i].RootEndpoint,
		}
	}
	planetListItems[len(planets)] = PlanetItem{
		Id:   "add",
		Name: "+ Add new",
		Desc: "select a zerotier planet file",
	}
	return planetListItems
}

type ActionItem struct {
	Id   string
	Name string
	Desc string
}

func (i ActionItem) FilterValue() string { return "" }
func (i ActionItem) Title() string       { return i.Name }
func (i ActionItem) Description() string { return i.Desc }

func CreateActionListView() list.Model {
	actionList := []list.Item{
		ActionItem{Id: "activate", Name: "Activate", Desc: "Activate the planet file"},
		ActionItem{Id: "rename", Name: "Rename", Desc: "Rename the planet file"},
		ActionItem{Id: "view", Name: "View info", Desc: "View the info of planet file"},
		ActionItem{Id: "delete", Name: "Delete", Desc: "Delete the planet file"},
	}
	l := list.New(actionList, list.NewDefaultDelegate(), 40, 30)
	l.SetShowStatusBar(false)
	l.Title = "What do you want?"
	return l
}
