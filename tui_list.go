package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	entries []Entry
	cursor  int
}

const (
	listActionNone = iota
	listActionOpen
	listActionAdd
	listActionEdit
	listActionDelete
	listActionEncryptFile
	listActionDecryptFile
)

type listAction struct {
	kind  int
	entry Entry
}

func newListModel(entries []Entry) ListModel {
	return ListModel{entries: entries}
}

func updateList(m ListModel, msg tea.Msg) (ListModel, tea.Cmd, listAction) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, listAction{}
	}

	switch km.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.entries) > 0 {
			return m, nil, listAction{kind: listActionOpen, entry: m.entries[m.cursor]}
		}
	case "a":
		return m, nil, listAction{kind: listActionAdd}
	case "e":
		if len(m.entries) > 0 {
			return m, nil, listAction{kind: listActionEdit, entry: m.entries[m.cursor]}
		}
	case "d":
		if len(m.entries) > 0 {
			return m, nil, listAction{kind: listActionDelete, entry: m.entries[m.cursor]}
		}
	case "f":
		return m, nil, listAction{kind: listActionEncryptFile}
	case "x":
		return m, nil, listAction{kind: listActionDecryptFile}
	}

	return m, nil, listAction{}
}

func viewList(m ListModel) string {
	if len(m.entries) == 0 {
		return "no entries yet.\n\n(a) add   (esc/ctrl+c) quit\n"
	}

	s := "your entries:\n\n"
	for i, e := range m.entries {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		s += fmt.Sprintf("%s%s\n", cursor, e.Name)
	}
	s += "\n(enter) open  (a) add  (e) edit  (d) delete  (f) encrypt file  (x) decrypt file  (ctrl+c) quit\n"
	return s
}
