package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/LanceLRQ/zerotier-switcher/src/tools"
	"github.com/charmbracelet/bubbles/list"
)

type PlanetItem struct {
	Planet    *configs.ZerotierPlanetFile
	Id        string
	Name      string
	Desc      string
	IsCurrent bool
}

func (i PlanetItem) FilterValue() string { return i.Name }
func (i PlanetItem) Title() string       { return i.Name }
func (i PlanetItem) Description() string { return i.Desc }

type ActionItem struct {
	Id   string
	Name string
	Desc string
}

func (i ActionItem) FilterValue() string { return "" }
func (i ActionItem) Title() string       { return i.Name }
func (i ActionItem) Description() string { return i.Desc }

func CreatePlanetListView(cfg *configs.ZerotierSwitcherProfile) list.Model {
	l := list.New(RenderPlanetListItem(cfg.Planets), list.NewDefaultDelegate(), 0, 0)
	l.SetShowStatusBar(false)
	l.Title = "Planet List"
	return l
}

func RenderPlanetListItem(planets []configs.ZerotierPlanetFile) []list.Item {
	cHash := tools.GetCurrentPlanetHashFromOS()
	planetListItems := make([]list.Item, len(planets))
	for i := range planets {
		isCurrent := tools.CheckIsCurrentPlanet(planets[i].Data, cHash)
		name := planets[i].Remark
		if isCurrent {
			name += " (current)"
		}
		planetListItems[i] = PlanetItem{
			Planet:    &planets[i],
			Id:        planets[i].Hash,
			Name:      name,
			Desc:      planets[i].RootEndpoint,
			IsCurrent: isCurrent,
		}
	}
	planetListItems = append(planetListItems, []list.Item{
		PlanetItem{Id: "add", Name: "+ Add new", Desc: "select a zerotier planet file"},
		PlanetItem{Id: "backup", Name: "→ Backup", Desc: "Backup config file to current directory"},
		PlanetItem{Id: "import", Name: "← Import", Desc: "See how to import config file"},
	}...)
	return planetListItems
}

func CreateActionListView() list.Model {
	l := list.New(RenderActionListItem(PlanetItem{}, true), list.NewDefaultDelegate(), 40, 30)
	l.SetShowStatusBar(false)
	l.Title = "What do you want?"
	return l
}

func RenderActionListItem(pItem PlanetItem, deleteAble bool) []list.Item {
	actionList := make([]list.Item, 0)
	if !pItem.IsCurrent {
		actionList = append(actionList, ActionItem{Id: "activate", Name: "Activate", Desc: "Activate the planet file"})
	}
	actionList = append(actionList, []list.Item{
		ActionItem{Id: "view", Name: "View info", Desc: "View the info of planet file"},
		ActionItem{Id: "rename", Name: "Rename", Desc: "Rename the planet file"},
		ActionItem{Id: "auto_join", Name: "Auto join", Desc: "Set auto join network id"},
	}...)
	if deleteAble {
		actionList = append(actionList, ActionItem{Id: "delete", Name: "Delete", Desc: "Delete the planet file"})
	}
	return actionList
}
