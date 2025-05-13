package views

import (
	"encoding/hex"
	"fmt"
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/LanceLRQ/zerotier-switcher/src/tools"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)
var filePickerStyle = lipgloss.NewStyle().Margin(2, 2)

type AppViewModel struct {
	screen             string
	config             *configs.ZerotierSwitcherProfile
	planetFile         *configs.ZerotierPlanetFile
	planetList         list.Model
	actionList         list.Model
	filePickerView     filepicker.Model
	filePickerErr      string
	filePickerSelected string
}

func (m AppViewModel) Init() tea.Cmd {
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			switch m.screen {
			case "list":
				return m, tea.Quit
			case "file_picker":
				m.screen = "list"
				return m, nil
			}
		case "backspace":
			switch m.screen {
			case "action":
				m.screen = "list"
				return m, nil
			}
		case "esc":
			switch m.screen {
			case "action", "file_picker":
				m.screen = "list"
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			switch m.screen {
			case "list":
				p, ok := m.planetList.SelectedItem().(PlanetItem)
				if ok {
					if p.Id == "add" {
						m.screen = "file_picker"
						m.filePickerView.CurrentDirectory, _ = os.UserHomeDir()
						return m, m.filePickerView.Init()
					} else {
						m.planetFile = p.Planet
						m.screen = "action"
					}
				}
			case "file_picker":
				if didSelect, path := m.filePickerView.DidSelectFile(msg); didSelect {
					// Get the path of the selected file.
					m.filePickerSelected = path
				}
				world, err := m.parsePlanetFile()
				if err != nil {
					m.filePickerErr = err.Error()
					break
				}
				planet := m.makePlanetFileItem(world)
				m.config.Planets = append(m.config.Planets, *planet)
				err = m.config.WriteAppConfig()
				if err != nil {
					m.filePickerErr = err.Error()
					break
				}
				// rebuild list
				m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
				// Back to list
				m.screen = "list"
				m.filePickerErr = ""
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		h, v := planetListStyle.GetFrameSize()
		_, fv := filePickerStyle.GetFrameSize()
		m.planetList.SetSize(msg.Width-h, msg.Height-v)
		m.actionList.SetSize(msg.Width-h, msg.Height-v)
		m.filePickerView.SetHeight(msg.Height - fv)
	}

	var cmd tea.Cmd
	switch m.screen {
	case "list":
		m.planetList, cmd = m.planetList.Update(msg)
	case "action":
		m.actionList, cmd = m.actionList.Update(msg)
	case "file_picker":
		m.filePickerView, cmd = m.filePickerView.Update(msg)
	}

	return m, cmd
}

func (m AppViewModel) View() string {
	switch m.screen {
	case "list":
		return m.planetList.View()
	case "action":
		return m.actionList.View()
	case "file_picker":
		var s strings.Builder
		if m.filePickerErr != "" {
			s.WriteString(fmt.Sprintf("\n Error: %s", m.filePickerErr))
		} else {
			s.WriteString("\n Please pick a zerotier planet file.")
		}
		s.WriteString("\n\n" + m.filePickerView.View() + "\n")
		return s.String()
	}
	return ""
}

func (m AppViewModel) parsePlanetFile() (*tools.World, error) {
	data, err := os.ReadFile(m.filePickerSelected)
	if err != nil {
		return nil, err
	}
	world, err := tools.ParseWorld(data)
	if err != nil {
		return nil, err
	}
	return world, nil
}
func (m AppViewModel) makePlanetFileItem(world *tools.World) *configs.ZerotierPlanetFile {
	var root tools.Root
	var ep tools.InetAddress
	if len(world.Roots) > 0 {
		root = world.Roots[0]
		if len(root.StableEndpoints) > 0 {
			ep = root.StableEndpoints[0]
		}
	}
	return &configs.ZerotierPlanetFile{
		Hash:         hex.EncodeToString(world.Signature[:32]),
		Remark:       ep.String(),
		Data:         world.ToBase64(),
		CreateTime:   world.Timestamp,
		WorldId:      world.ID,
		WorldType:    world.Type,
		RootIdentity: root.Identity.String(),
		RootEndpoint: ep.String(),
	}
}

func CreateAppView(cfg *configs.ZerotierSwitcherProfile) (*AppViewModel, error) {
	m := AppViewModel{
		screen:         "list",
		config:         cfg,
		planetList:     CreatePlanetListView(cfg),
		actionList:     CreateActionListView(),
		filePickerView: filepicker.New(),
	}
	return &m, nil
}
