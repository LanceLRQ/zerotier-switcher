package views

import (
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
)

type PlanetItem struct {
	configs.ZerotierPlanetFile
}

func (i PlanetItem) FilterValue() string { return i.Remark }
func (i PlanetItem) Title() string       { return i.Remark }
func (i PlanetItem) Description() string { return i.RootEndpoint }
