package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormMode int

const (
	FormModeAdd FormMode = iota
	FormModeEdit
)

const (
	focusName = iota
	focusBody
)

type FormModel struct {
	mode      FormMode
	original  Entry // for edit mode, the entry being replaced
	nameInput textinput.Model
	bodyArea  textarea.Model
	focus     int
}

const (
	formActionNone = iota
	formActionSave
	formActionCancel
)

type formAction struct {
	kind  int
	entry Entry
}

func newFormModel(mode FormMode, existing Entry) FormModel {
	ni := textinput.New()
	ni.Placeholder = "entry name"
	ta := textarea.New()
	ta.Placeholder = "write your entry..."

	if mode == FormModeEdit {
		ni.SetValue(existing.Name)
		ta.SetValue(existing.Body)
	}

	ni.Focus()

	return FormModel{
		mode:      mode,
		original:  existing,
		nameInput: ni,
		bodyArea:  ta,
		focus:     focusName,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func updateForm(m FormModel, msg tea.Msg) (FormModel, tea.Cmd, formAction) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			return m, nil, formAction{kind: formActionCancel}

		case "tab":
			if m.focus == focusName {
				m.focus = focusBody
				m.nameInput.Blur()
				m.bodyArea.Focus()
			} else {
				m.focus = focusName
				m.bodyArea.Blur()
				m.nameInput.Focus()
			}
			return m, nil, formAction{}

		case "ctrl+s":
			name := m.nameInput.Value()
			body := m.bodyArea.Value()
			created := time.Now()
			if m.mode == FormModeEdit {
				created = m.original.Created
			}
			entry := Entry{Name: name, Body: body, Created: created}
			return m, nil, formAction{kind: formActionSave, entry: entry}
		}
	}

	var cmd tea.Cmd
	if m.focus == focusName {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.bodyArea, cmd = m.bodyArea.Update(msg)
	}
	return m, cmd, formAction{}
}

func viewForm(m FormModel) string {
	title := "add entry"
	if m.mode == FormModeEdit {
		title = "edit entry"
	}

	s := title + "\n\n"
	s += "Name: " + m.nameInput.View() + "\n\n"
	s += "Body:\n" + m.bodyArea.View() + "\n\n"
	s += "(tab) switch field  (ctrl+s) save  (esc) cancel\n"
	return s
}
