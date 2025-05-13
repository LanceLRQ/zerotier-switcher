package views

import (
	"encoding/hex"
	"fmt"
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/LanceLRQ/zerotier-switcher/src/tools"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)
var filePickerStyle = lipgloss.NewStyle().Margin(2, 2)
var filePickerErrorStyle = lipgloss.NewStyle().Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15"))

const MaxRemarkLength = 64

type AppViewModel struct {
	screen             string
	config             *configs.ZerotierSwitcherProfile
	planetFile         *configs.ZerotierPlanetFile
	planetList         list.Model
	actionList         list.Model
	filePickerView     filepicker.Model
	errorMessage       string
	filePickerSelected string
	remarkInput        textinput.Model
}

func (m AppViewModel) Init() tea.Cmd {
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.screen == "file_picker" {
			m.errorMessage = ""
		}
		switch msg.String() {
		case "backspace":
			switch m.screen {
			case "action":
				m.screen = "list"
				return m, nil
			}
		case "esc":
			switch m.screen {
			case "list":
				return m, tea.Quit
			case "action", "file_picker", "rename":
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
						m.actionList.Title = m.getActionPageTitle()
						m.screen = "action"
					}
				}
			case "file_picker":
				m.filePickerView, cmd = m.filePickerView.Update(msg)
				if didSelect, path := m.filePickerView.DidSelectFile(msg); didSelect {
					// Get the path of the selected file.
					m.filePickerSelected = path
				}
				world, err := m.parsePlanetFile()
				if err != nil {
					m.errorMessage = fmt.Sprintf("Not a valid planet file: %s", err.Error())
					break
				}
				planet := m.makePlanetFileItem(world)
				m.config.Planets = append(m.config.Planets, *planet)
				err = m.config.WriteAppConfig()
				if err != nil {
					m.errorMessage = fmt.Sprintf("Save profile error: %s", err.Error())
					break
				}
				// rebuild list
				m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
				// Back to list
				m.screen = "list"
				m.errorMessage = ""
				return m, cmd
			case "action":
				aItem, ok := m.actionList.SelectedItem().(ActionItem)
				if ok {
					switch aItem.Id {
					case "rename":
						m.screen = "rename"
						m.remarkInput.SetValue(m.planetFile.Remark)
						m.remarkInput.SetCursor(0)
						return m, textinput.Blink
					}
				}
			case "rename":
				newVal := m.remarkInput.Value()
				if newVal == "" {
					newVal = m.planetFile.RootEndpoint
				}
				m.planetFile.Remark = newVal
				err := m.savePlanetChange()
				if err != nil {
					m.errorMessage = fmt.Sprintf("Save profile error: %s", err.Error())
					break
				}
				m.actionList.Title = m.getActionPageTitle()
				// rebuild list
				m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
				m.screen = "action"
				m.errorMessage = ""
				return m, cmd
			}
		}
	case tea.WindowSizeMsg:
		h, v := planetListStyle.GetFrameSize()
		_, fv := filePickerStyle.GetFrameSize()
		m.planetList.SetSize(msg.Width-h, msg.Height-v)
		m.actionList.SetSize(msg.Width-h, msg.Height-v)
		m.filePickerView.SetHeight(msg.Height - fv)
	}

	switch m.screen {
	case "list":
		m.planetList, cmd = m.planetList.Update(msg)
	case "action":
		m.actionList, cmd = m.actionList.Update(msg)
	case "file_picker":
		m.filePickerView, cmd = m.filePickerView.Update(msg)
	case "rename":
		m.remarkInput, cmd = m.remarkInput.Update(msg)
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
		s.WriteString("\n Please pick a zerotier planet file.")
		s.WriteString("\n\n" + m.filePickerView.View() + "\n")
		if m.errorMessage != "" {
			s.WriteString(filePickerErrorStyle.Render(m.errorMessage))
		}
		return s.String()
	case "rename":
		return fmt.Sprintf(
			"Write a remark for the planet file:\n\n%s\n(%d/%d)\n\n%s",
			m.remarkInput.View(),
			len(m.remarkInput.Value()),
			MaxRemarkLength,
			"(esc to back)",
		) + "\n"
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

func (m AppViewModel) savePlanetChange() error {
	for i := 0; i < len(m.config.Planets); i++ {
		item := m.config.Planets[i]
		if item.Hash == m.planetFile.Hash {
			m.config.Planets[i] = item
		}
	}
	return m.config.WriteAppConfig()
}
func (m AppViewModel) getActionPageTitle() string {
	rTitle := m.planetFile.Remark
	if len(rTitle) > 16 {
		rTitle = rTitle[:16] + "..."
	}
	return fmt.Sprintf("Planet file: %s (%v)", rTitle, m.planetFile.RootEndpoint)
}

func CreateAppView(cfg *configs.ZerotierSwitcherProfile) (*AppViewModel, error) {
	m := AppViewModel{
		screen:         "list",
		config:         cfg,
		planetList:     CreatePlanetListView(cfg),
		actionList:     CreateActionListView(),
		filePickerView: filepicker.New(),
		remarkInput:    CreateRemarkInput("remark text", MaxRemarkLength),
	}
	return &m, nil
}
