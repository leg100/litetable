package main

import (
	"github.com/charmbracelet/bubbles/key"
)

var keys = struct {
	Quit     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	),
}
