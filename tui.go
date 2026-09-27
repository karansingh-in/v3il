package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	screenUnlock = "unlock"
	screenList   = "list"
	screenDetail = "detail"
	screenForm   = "form"
	screenFileOp = "fileop"
)

const idleTimeout = 60 * time.Second

type Model struct {
	screen     string
	key        []byte
	vaultSalt  []byte
	entries    []Entry
	cursor     int
	lastActive time.Time
	path       string

	unlock UnlockModel
	list   ListModel
	detail DetailModel
	form   FormModel
	fileop FileOpModel
}

type autolockTickMsg struct{}

func initialModel(path string) Model {
	return Model{
		screen:     screenUnlock,
		path:       path,
		lastActive: time.Now(),
		unlock:     newUnlockModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.unlock.Init(),
		tickAutolock(),
	)
}

func tickAutolock() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return autolockTickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// global quit
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// global: any keypress resets the idle timer
	if _, ok := msg.(tea.KeyMsg); ok {
		m.lastActive = time.Now()
	}

	// global: autolock check
	if _, ok := msg.(autolockTickMsg); ok {
		if m.screen != screenUnlock && time.Since(m.lastActive) > idleTimeout {
			for i := range m.key {
				m.key[i] = 0
			}
			m.key = nil
			m.entries = nil
			m.screen = screenUnlock
			m.unlock = newUnlockModel()
			return m, tea.Batch(m.unlock.Init(), tickAutolock())
		}
		return m, tickAutolock()
	}

	switch m.screen {
	case screenUnlock:
		newUnlock, cmd, result := updateUnlock(m.unlock, msg)
		m.unlock = newUnlock
		if result.success {
			m.key = result.key
			m.vaultSalt = result.salt
			m.entries = result.entries
			m.screen = screenList
			m.list = newListModel(m.entries)
		}
		return m, cmd

	case screenList:
		newList, cmd, action := updateList(m.list, msg)
		m.list = newList
		switch action.kind {
		case listActionOpen:
			m.detail = newDetailModel(action.entry)
			m.screen = screenDetail
		case listActionAdd:
			m.form = newFormModel(FormModeAdd, Entry{})
			m.screen = screenForm
		case listActionEdit:
			m.form = newFormModel(FormModeEdit, action.entry)
			m.screen = screenForm
		case listActionDelete:
			// delete immediately, save-on-mutation policy
			deleteEntry(&m, action.entry.Name)
			m.list = newListModel(m.entries)
		case listActionEncryptFile:
			m.fileop = newFileOpModel(FileOpEncrypt)
			m.screen = screenFileOp
		case listActionDecryptFile:
			m.fileop = newFileOpModel(FileOpDecrypt)
			m.screen = screenFileOp
		}
		return m, cmd

	case screenDetail:
		newDetail, cmd, action := updateDetail(m.detail, msg)
		m.detail = newDetail
		switch action.kind {
		case detailActionBack:
			m.screen = screenList
		case detailActionEdit:
			m.form = newFormModel(FormModeEdit, action.entry)
			m.screen = screenForm
		}
		return m, cmd

	case screenForm:
		newForm, cmd, action := updateForm(m.form, msg)
		m.form = newForm
		switch action.kind {
		case formActionCancel:
			m.screen = screenList
		case formActionSave:
			saveEntry(&m, action.entry, m.form.mode)
			m.list = newListModel(m.entries)
			m.screen = screenList
		}
		return m, cmd

	case screenFileOp:
		newFileOp, cmd, action := updateFileOp(m.fileop, msg, m.key)
		m.fileop = newFileOp
		if action.kind == fileOpActionBack {
			m.screen = screenList
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenUnlock:
		return viewUnlock(m.unlock)
	case screenList:
		return viewList(m.list)
	case screenDetail:
		return viewDetail(m.detail)
	case screenForm:
		return viewForm(m.form)
	case screenFileOp:
		return viewFileOp(m.fileop)
	}
	return "unknown screen\n"
}

// --- mutation helpers: these wrap your existing Vault methods + immediate save ---

func saveEntry(m *Model, e Entry, mode FormMode) {
	v := &Vault{entries: entriesToMap(m.entries), salt: currentSalt(m)}
	if mode == FormModeAdd {
		v.Add(e) // duplicate-name error intentionally ignored here for v1; surface later if needed
	} else {
		v.Update(e.Name, e)
	}
	m.entries = v.List()
	persistVault(m, v)
}

func deleteEntry(m *Model, name string) {
	v := &Vault{entries: entriesToMap(m.entries), salt: currentSalt(m)}
	v.Delete(name)
	m.entries = v.List()
	persistVault(m, v)
}

func entriesToMap(entries []Entry) map[string]Entry {
	result := make(map[string]Entry)
	for _, e := range entries {
		result[e.Name] = e
	}
	return result
}

// currentSalt and persistVault are split out because Model doesn't keep a
// live *Vault around (only entries + key) — see note below the file.
func currentSalt(m *Model) []byte {
	return m.vaultSalt
}

func persistVault(m *Model, v *Vault) {
	saveVaultWithKey(v, m.key, m.path)
}

func runTUI() error {
	path, err := VaultPath()
	if err != nil {
		return err
	}
	p := tea.NewProgram(initialModel(path))
	_, err = p.Run()
	return err
}
