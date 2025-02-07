package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
)

func main() {
	data, err := newData()
	if err != nil {
		panic(err.Error())
	}
	model := Model{
		data: data,
		lgt: table.New().
			Headers("cursor", "author", "title").
			Data(data),
	}
	model.Height(7)

	for _, book := range books {
		err := model.Append([]string{book[0], book[1]})
		if err != nil {
			panic(err.Error())
		}
	}

	prog := tea.NewProgram(model)
	_, err = prog.Run()
	if err != nil {
		panic(err.Error())
	}
}
