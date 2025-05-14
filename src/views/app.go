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
	"path/filepath"
	"strings"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)
var filePickerStyle = lipgloss.NewStyle().Margin(4, 2)
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
	confirmCursor      int
}

func (m AppViewModel) Init() tea.Cmd {
	return nil
}

func (m AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.errorMessage != "" {
			m.errorMessage = ""
		}
		switch msg.String() {
		case "down", "w", "j":
			if m.screen == "delete_confirm" {
				m.confirmCursor++
				if m.confirmCursor >= 2 {
					m.confirmCursor = 0
				}
			}
		case "up", "s", "k":
			if m.screen == "delete_confirm" {
				m.confirmCursor--
				if m.confirmCursor < 0 {
					m.confirmCursor = 1
				}
			}

		case "backspace":
			switch m.screen {
			case "action":
				m.screen = "list"
			case "view_planet", "delete_confirm":
				m.screen = "action"
			}
		case "esc":
			switch m.screen {
			case "list":
				return m, tea.Quit
			case "action", "file_picker", "rename":
				m.screen = "list"
			case "view_planet", "delete_confirm":
				m.screen = "action"
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
						return m, m.filePickerView.Init()
					} else {
						m.planetFile = p.Planet
						m.actionList.Title = m.getActionPageTitle()
						m.screen = "action"
					}
				}
			case "file_picker":
				m.filePickerView, cmd = m.filePickerView.Update(msg)
				if didSelect, sPath := m.filePickerView.DidSelectFile(msg); didSelect {
					// Get the path of the selected file.
					s, err := os.Stat(sPath)
					if err != nil || s.IsDir() {
						break
					}
					m.filePickerSelected = sPath

					world, err := m.parsePlanetFile()
					if err != nil {
						m.errorMessage = fmt.Sprintf("Not a valid planet file: %s", err.Error())
						break
					}
					planet := m.makePlanetFileItem(world, filepath.Base(sPath))
					if m.isPlanetFileExists(planet) {
						m.errorMessage = fmt.Sprintf("Planet file (%s) exists", planet.RootIdentity[:8])
						break
					}
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
				}
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
					case "view":
						m.screen = "view_planet"
						return m, nil
					case "delete":
						if len(m.config.Planets) <= 1 {
							m.errorMessage = fmt.Sprintf("The last planet file cannot be delete. ")
							break
						}
						m.confirmCursor = 1
						m.screen = "delete_confirm"
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
			case "delete_confirm":
				if m.confirmCursor == 0 {
					err := m.removePlanet()
					if err != nil {
						m.errorMessage = fmt.Sprintf("Save profile error: %s", err.Error())
						break
					}
					// rebuild list
					m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
					m.screen = "list"
					m.planetFile = nil
					m.errorMessage = ""
				} else {
					m.screen = "action"
					m.errorMessage = ""
				}
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
	var s strings.Builder

	switch m.screen {
	case "list":
		s.WriteString(m.planetList.View())
	case "action":
		s.WriteString(m.actionList.View())
	case "file_picker":
		s.WriteString("\n Please pick a zerotier planet file.")
		s.WriteString("\n\n" + m.filePickerView.View() + "\n")

	case "rename":
		s.WriteString(fmt.Sprintf(
			"Write a remark for the planet file:\n\n%s\n(%d/%d)\n\n%s",
			m.remarkInput.View(),
			len(m.remarkInput.Value()),
			MaxRemarkLength,
			"(esc to back)",
		) + "\n")
	case "view_planet":
		s.WriteString(m.renderPlanetFileDetailView() + "\n\n(esc to back)")
	case "delete_confirm":
		s.WriteString(m.renderDeleteConfirm() + "\n\n(esc to back)")
	}

	if m.errorMessage != "" {
		s.WriteString("\n" + filePickerErrorStyle.Render(m.errorMessage))
	}

	return s.String()
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
func (m AppViewModel) makePlanetFileItem(world *tools.World, fileName string) *configs.ZerotierPlanetFile {
	var root tools.Root
	var ep tools.InetAddress
	if len(world.Roots) > 0 {
		root = world.Roots[0]
		if len(root.StableEndpoints) > 0 {
			ep = root.StableEndpoints[0]
		}
	}
	if fileName == "" {
		fileName = ep.String()
	}
	return &configs.ZerotierPlanetFile{
		Hash:         hex.EncodeToString(world.Signature[:32]),
		Remark:       fileName,
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
func (m AppViewModel) removePlanet() error {
	var pList []configs.ZerotierPlanetFile
	for i := 0; i < len(m.config.Planets); i++ {
		item := m.config.Planets[i]
		if item.Hash != m.planetFile.Hash {
			pList = append(pList, item)
		}
	}
	m.config.Planets = pList
	return m.config.WriteAppConfig()
}
func (m AppViewModel) isPlanetFileExists(p *configs.ZerotierPlanetFile) bool {
	for i := 0; i < len(m.config.Planets); i++ {
		item := m.config.Planets[i]
		if item.Hash == p.Hash {
			return true
		}
	}
	return false
}
func (m AppViewModel) getActionPageTitle() string {
	rTitle := m.planetFile.Remark
	if len(rTitle) > 16 {
		rTitle = rTitle[:16] + "..."
	}
	return fmt.Sprintf("Planet file: %s (%v)", rTitle, m.planetFile.RootEndpoint)
}

func (m AppViewModel) renderPlanetFileDetailView() string {
	if m.planetFile == nil {
		return ""
	}
	world, err := tools.ParsePlanetBase64(m.planetFile.Data)
	if err != nil {
		return err.Error()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintln("ZeroTier Planet Information:"))
	sb.WriteString(fmt.Sprintf("  ID: %d\n", world.ID))
	sb.WriteString(fmt.Sprintf("  Type: %d (1=Planet, 127=Moon)\n", world.Type))
	sb.WriteString(fmt.Sprintf("  Timestamp: %d\n", world.Timestamp))
	sb.WriteString(fmt.Sprintf("  Update Signer Public Key: %s\n", hex.EncodeToString(world.UpdatesMustBeSignedBy[:])))
	sb.WriteString(fmt.Sprintf("  Signature: %s...\n", hex.EncodeToString(world.Signature[:16])))
	sb.WriteString(fmt.Sprintf("  Number of Roots: %d\n", len(world.Roots)))

	for i, root := range world.Roots {
		sb.WriteString(fmt.Sprintf("\nRoot Server %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Identity: %s\n", root.Identity.String()))
		for j, ep := range root.StableEndpoints {
			sb.WriteString(fmt.Sprintf("  Endpoint %d: %s\n", j+1, ep.String()))
		}
	}
	return sb.String()
}

func (m AppViewModel) renderDeleteConfirm() string {
	s := strings.Builder{}
	s.WriteString("Do you want to delete the planet file?\n\n")
	choices := []string{"Yes", "No"}
	for i := 0; i < len(choices); i++ {
		if m.confirmCursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	return s.String()
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
	m.filePickerView.CurrentDirectory, _ = os.UserHomeDir()
	return &m, nil
}
