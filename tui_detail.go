package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type DetailModel struct {
	entry Entry
}

const (
	detailActionNone = iota
	detailActionBack
	detailActionEdit
)

type detailAction struct {
	kind  int
	entry Entry
}

func newDetailModel(e Entry) DetailModel {
	return DetailModel{entry: e}
}

func updateDetail(m DetailModel, msg tea.Msg) (DetailModel, tea.Cmd, detailAction) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, detailAction{}
	}

	switch km.String() {
	case "esc", "b":
		return m, nil, detailAction{kind: detailActionBack}
	case "e":
		return m, nil, detailAction{kind: detailActionEdit, entry: m.entry}
	}

	return m, nil, detailAction{}
}

func viewDetail(m DetailModel) string {
	s := fmt.Sprintf("Name: %s\n\n", m.entry.Name)
	s += fmt.Sprintf("Body:\n%s\n\n", m.entry.Body)
	s += fmt.Sprintf("Created: %s\n\n", m.entry.Created.Format("2006-01-02 15:04:05"))
	s += "(e) edit  (esc/b) back  (ctrl+c) quit\n"
	return s
}
