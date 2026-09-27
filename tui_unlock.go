package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type UnlockModel struct {
	input  textinput.Model
	errMsg string
}

type unlockResult struct {
	success bool
	key     []byte
	salt    []byte
	entries []Entry
}

func newUnlockModel() UnlockModel {
	ti := textinput.New()
	ti.Placeholder = "Master password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	return UnlockModel{input: ti}
}

func (m UnlockModel) Init() tea.Cmd {
	return textinput.Blink
}

type unlockAttemptMsg struct {
	key     []byte
	salt    []byte
	entries []Entry
	err     error
}

func tryUnlock(password string, path string) tea.Cmd {
	return func() tea.Msg {
		pwBytes := []byte(password)
		v, err := LoadVault(pwBytes, path)
		if err != nil {
			return unlockAttemptMsg{err: err}
		}
		key := DeriveKey(pwBytes, v.salt)
		return unlockAttemptMsg{key: key, salt: v.salt, entries: v.List()}
	}
}

func updateUnlock(m UnlockModel, msg tea.Msg) (UnlockModel, tea.Cmd, unlockResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			password := m.input.Value()
			m.input.SetValue("")
			path, _ := VaultPath()
			return m, tryUnlock(password, path), unlockResult{}
		}
	case unlockAttemptMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil, unlockResult{}
		}
		return m, nil, unlockResult{success: true, key: msg.key, salt: msg.salt, entries: msg.entries}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd, unlockResult{}
}

func viewUnlock(m UnlockModel) string {
	s := "v3il - karansingh-in\n\n" + m.input.View() + "\n\n(ctrl+c to quit)\n"
	if m.errMsg != "" {
		s += "\nerror: " + m.errMsg + "\n"
	}
	return s
}
