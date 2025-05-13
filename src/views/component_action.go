package views

import (
	"github.com/charmbracelet/bubbles/textinput"
)

func CreateRemarkInput(placeholder string, textLimit int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = textLimit
	ti.Width = 32
	return ti
}
