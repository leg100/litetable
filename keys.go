package main

import (
	"github.com/charmbracelet/bubbles/key"
)

var keys = struct {
	Quit        key.Binding
	PreviousRow key.Binding
	NextRow     key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
}{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	PreviousRow: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "previous row"),
	),
	NextRow: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "next row"),
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
