package views

import (
	"github.com/charmbracelet/bubbles/list"
)

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
		ActionItem{Id: "activate", Name: "Activate", Desc: "Activate this planet file"},
		ActionItem{Id: "view", Name: "View info", Desc: "View the info of planet file"},
		ActionItem{Id: "delete", Name: "Delete", Desc: "Delete the planet file"},
	}
	l := list.New(actionList, list.NewDefaultDelegate(), 40, 30)
	l.SetShowStatusBar(false)
	l.Title = "What do you want?"
	return l
}
