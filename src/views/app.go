package views

import (
	"encoding/hex"
	"fmt"
	"github.com/LanceLRQ/zerotier-switcher/src/configs"
	"github.com/LanceLRQ/zerotier-switcher/src/tools"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var planetListStyle = lipgloss.NewStyle().Margin(1, 2)
var filePickerStyle = lipgloss.NewStyle().Margin(4, 2)
var filePickerErrorStyle = lipgloss.NewStyle().Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15"))
var filePickerSuccessStyle = lipgloss.NewStyle().Background(lipgloss.Color("10")).Foreground(lipgloss.Color("15"))
var activateTitleStyle = lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15"))
var progressBarPadding = 2
var progressBarMaxWidth = 80
var maxActivateStep = float64(7)

type progressMsg struct {
	step  int
	desc  string
	error bool
}

const MaxRemarkLength = 64
const MaxAutoJoinNetworkLength = 64

type AppViewModel struct {
	IsRunAsRoot        bool
	Program            *tea.Program
	screen             string
	config             *configs.ZerotierSwitcherProfile
	currentPlanetItem  PlanetItem
	planetFile         *configs.ZerotierPlanetFile
	planetList         list.Model
	actionList         list.Model
	filePickerView     filepicker.Model
	errorMessage       string
	successMessage     string
	filePickerSelected string
	remarkInput        textinput.Model
	autoJoinInput      textinput.Model
	progressBar        progress.Model
	activateStep       int
	activateLock       bool
	activateStepDesc   string
	confirmCursor      int
	currentWindowSize  tea.WindowSizeMsg
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
		if m.successMessage != "" {
			m.successMessage = ""
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
			case "action", "file_picker", "import_tip":
				m.screen = "list"
			case "activate", "view_planet", "delete_confirm", "rename", "auto_join":
				m.screen = "action"
			case "activate_process":
				if !m.activateLock {
					m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
					m.screen = "list"
				}
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
					} else if p.Id == "backup" {
						currentDir, err := os.Getwd()
						if err != nil {
							m.errorMessage = err.Error()
						}
						bakName := path.Join(currentDir, fmt.Sprintf("zerotier-switcher.backup.%d.json", time.Now().Unix()))
						err = m.config.WriteAppConfigWithPath(bakName)
						if err != nil {
							m.errorMessage = err.Error()
						}
						m.successMessage = fmt.Sprintf("Saved to current folder")
					} else if p.Id == "import" {
						m.screen = "import_tip"
					} else {
						m.planetFile = p.Planet
						m.currentPlanetItem = p
						m.actionList.Title = m.getActionPageTitle()
						m.actionList.SetItems(RenderActionListItem(m.currentPlanetItem, len(m.config.Planets) > 1))
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
					case "activate":
						m.screen = "activate"
						return m, nil
					case "rename":
						m.screen = "rename"
						m.remarkInput.SetValue(m.planetFile.Remark)
						m.remarkInput.SetCursor(0)
						return m, textinput.Blink
					case "auto_join":
						m.screen = "auto_join"
						m.autoJoinInput.SetValue(m.planetFile.AutoJoinNetwork)
						m.autoJoinInput.SetCursor(0)
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
			case "rename", "auto_join":
				if m.screen == "rename" {
					newVal := m.remarkInput.Value()
					if newVal == "" {
						newVal = m.planetFile.RootEndpoint
					}
					m.planetFile.Remark = newVal
				} else if m.screen == "auto_join" {
					newVal := m.autoJoinInput.Value()
					m.planetFile.AutoJoinNetwork = newVal
				}
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
			case "activate":
				if !m.IsRunAsRoot {
					break
				}
				m.activateStep = 0
				cmd = m.progressBar.SetPercent(0)
				m.screen = "activate_process"
				m.activateLock = true
				go func() {
					currentStep := 0
					if err := tools.ReplacePlanetAndJoinNetwork(
						m.planetFile.Data,
						m.planetFile.AutoJoinNetwork,
						func(step int, desc string) {
							currentStep = step
							m.Program.Send(progressMsg{
								step:  step,
								desc:  desc,
								error: false,
							})
						},
					); err != nil {
						m.Program.Send(progressMsg{
							step:  currentStep,
							desc:  filePickerErrorStyle.Width(m.currentWindowSize.Width).Render(err.Error()),
							error: true,
						})
					}
				}()
			case "activate_process":
				if !m.activateLock {
					m.planetList.SetItems(RenderPlanetListItem(m.config.Planets))
					m.screen = "list"
				}
			}
		}
	case tea.WindowSizeMsg:
		m.currentWindowSize = msg
		h, v := planetListStyle.GetFrameSize()
		_, fv := filePickerStyle.GetFrameSize()
		m.planetList.SetSize(msg.Width-h, msg.Height-v)
		m.actionList.SetSize(msg.Width-h, msg.Height-v)
		m.filePickerView.SetHeight(msg.Height - fv)
		m.progressBar.Width = msg.Width - progressBarPadding*2 - 4
		if m.progressBar.Width > progressBarMaxWidth {
			m.progressBar.Width = progressBarMaxWidth
		}
	case progressMsg:
		m.activateStep = msg.step
		m.activateStepDesc = msg.desc
		if msg.error || msg.step >= int(maxActivateStep) {
			m.activateLock = false
		} else {
			m.activateLock = true
		}
		cmd = m.progressBar.SetPercent(float64(m.activateStep) / maxActivateStep)
		// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd
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
	case "auto_join":
		m.autoJoinInput, cmd = m.autoJoinInput.Update(msg)
	}

	return m, cmd
}

func (m AppViewModel) View() string {
	var s strings.Builder

	pad := strings.Repeat(" ", progressBarPadding)

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
			"Write a remark for the planet file:\n\n%s\n\n(%d/%d)\n\n%s\n\n",
			m.remarkInput.View(),
			len(m.remarkInput.Value()),
			MaxRemarkLength,
			"(ESC to back)",
		) + "\n")
	case "auto_join":
		s.WriteString(fmt.Sprintf(
			"Set the network id:\n\n%s\n\n(%d/%d)\n\n%s\n\n",
			m.autoJoinInput.View(),
			len(m.autoJoinInput.Value()),
			MaxAutoJoinNetworkLength,
			"(ESC to back)",
		) + "\n")
	case "view_planet":
		s.WriteString(m.renderPlanetFileDetailView() + "\n\n(ESC to back)")
	case "delete_confirm":
		s.WriteString(m.renderDeleteConfirm() + "\n\n(ESC to back)")
	case "activate":
		s.WriteString("\n" + pad)
		s.WriteString(m.renderActivateView() + "\n\n")
		if m.IsRunAsRoot {
			s.WriteString("ENTER to continue, ")
		}
		s.WriteString("ESC to back")
	case "activate_process":
		s.WriteString("\n" + pad + activateTitleStyle.Render("Processing"))
		s.WriteString("\n\n" + pad + m.progressBar.View() + "\n\n")
		s.WriteString(m.activateStepDesc)
		if !m.activateLock {
			s.WriteString("\n\n(ENTER to back)")
		}
		s.WriteString("\n\n")
	case "import_tip":
		s.WriteString(fmt.Sprintf(
			"%s\n\n%s\n\n;-)\n\n(ESC to back)",
			activateTitleStyle.Render("Please replace your config file to:"),
			lipgloss.NewStyle().Width(m.currentWindowSize.Width).Render(path.Join(configs.GetDefaultConfigPath(), "profile.json")),
		))
	}

	if m.errorMessage != "" {
		s.WriteString("\n" + filePickerErrorStyle.Render(m.errorMessage))
	}
	if m.successMessage != "" {
		s.WriteString("\n" + filePickerSuccessStyle.Render(m.successMessage))
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
func (m AppViewModel) renderActivateView() string {
	if m.planetFile == nil {
		return ""
	}
	world, err := tools.ParsePlanetBase64(m.planetFile.Data)
	if err != nil {
		return err.Error()
	}
	var sb strings.Builder
	sb.WriteString(activateTitleStyle.Render("Activate planet file") + "\n\n")

	sb.WriteString(fmt.Sprintln("ZeroTier Planet Information:"))
	sb.WriteString(fmt.Sprintf("  ID: %d\n", world.ID))
	sb.WriteString(fmt.Sprintf("  Type: %d (1=Planet, 127=Moon)\n", world.Type))

	for i, root := range world.Roots {
		sb.WriteString(fmt.Sprintf("\nRoot Server %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Identity: %s\n", root.Identity.String()))
		for j, ep := range root.StableEndpoints {
			sb.WriteString(fmt.Sprintf("  Endpoint %d: %s\n", j+1, ep.String()))
		}
	}

	sb.WriteString(fmt.Sprintf("\nJoin network: %s\n", m.planetFile.AutoJoinNetwork))

	if !m.IsRunAsRoot {
		sb.WriteString("\n" + filePickerErrorStyle.Render("You must run this program as root (administrator)"))
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
		IsRunAsRoot:    tools.IsRunAsRoot(),
		screen:         "list",
		config:         cfg,
		planetList:     CreatePlanetListView(cfg),
		actionList:     CreateActionListView(),
		filePickerView: filepicker.New(),
		remarkInput:    CreateRemarkInput("remark text", MaxRemarkLength),
		autoJoinInput:  CreateRemarkInput("network id", MaxAutoJoinNetworkLength),
		progressBar:    progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C")),
	}
	m.filePickerView.CurrentDirectory, _ = os.UserHomeDir()
	return &m, nil
}
