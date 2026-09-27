package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FileOpMode int

const (
	FileOpEncrypt FileOpMode = iota
	FileOpDecrypt
)

type FileOpModel struct {
	mode      FileOpMode
	pathInput textinput.Model
	status    string
	err       string
	done      bool
}

const (
	fileOpActionNone = iota
	fileOpActionBack
)

type fileOpAction struct {
	kind int
}

func newFileOpModel(mode FileOpMode) FileOpModel {
	ti := textinput.New()
	if mode == FileOpEncrypt {
		ti.Placeholder = "path to file to encrypt"
	} else {
		ti.Placeholder = "path to .v3il file to decrypt"
	}
	ti.Focus()
	return FileOpModel{mode: mode, pathInput: ti}
}

func (m FileOpModel) Init() tea.Cmd {
	return textinput.Blink
}

// fileOpResultMsg carries the outcome of an async encrypt/decrypt attempt back into Update.
type fileOpResultMsg struct {
	outPath string
	err     error
}

// runFileOp does the actual work in the background (this is real file I/O + crypto,
// same reasoning as tryUnlock — never block Update directly).
func runFileOp(mode FileOpMode, path string, key []byte) tea.Cmd {
	return func() tea.Msg {
		if mode == FileOpEncrypt {
			data, err := os.ReadFile(path)
			if err != nil {
				return fileOpResultMsg{err: err}
			}
			ciphertext, nonce, err := Encrypt(key, data)
			if err != nil {
				return fileOpResultMsg{err: err}
			}
			var out []byte
			out = append(out, nonce...)
			out = append(out, ciphertext...)

			outPath := path + ".v3il"
			if err := os.WriteFile(outPath, out, 0600); err != nil {
				return fileOpResultMsg{err: err}
			}
			return fileOpResultMsg{outPath: outPath}
		}

		// decrypt
		data, err := os.ReadFile(path)
		if err != nil {
			return fileOpResultMsg{err: err}
		}
		const nonceSize = 12 // AES-GCM nonce size in bytes.
		if len(data) < nonceSize {
			return fileOpResultMsg{err: fmt.Errorf("file is too small to be a valid encrypted file")}
		}
		nonce := data[:nonceSize]
		ciphertext := data[nonceSize:]

		plaintext, err := Decrypt(key, nonce, ciphertext)
		if err != nil {
			return fileOpResultMsg{err: err}
		}

		outPath := strings.TrimSuffix(path, ".v3il")
		if outPath == path {
			outPath = path + ".decrypted"
		}
		if err := os.WriteFile(outPath, plaintext, 0600); err != nil {
			return fileOpResultMsg{err: err}
		}
		return fileOpResultMsg{outPath: outPath}
	}
}

func updateFileOp(m FileOpModel, msg tea.Msg, key []byte) (FileOpModel, tea.Cmd, fileOpAction) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil, fileOpAction{kind: fileOpActionBack}
		case "enter":
			if m.done {
				return m, nil, fileOpAction{kind: fileOpActionBack}
			}
			path := m.pathInput.Value()
			m.status = "working..."
			return m, runFileOp(m.mode, path, key), fileOpAction{}
		}
	case fileOpResultMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = ""
		} else {
			m.status = "done: " + msg.outPath
			m.err = ""
			m.done = true
		}
		return m, nil, fileOpAction{}
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd, fileOpAction{}
}

func viewFileOp(m FileOpModel) string {
	title := "encrypt a file"
	if m.mode == FileOpDecrypt {
		title = "decrypt a file"
	}

	s := title + "\n\n" + m.pathInput.View() + "\n\n"
	if m.status != "" {
		s += m.status + "\n\n"
	}
	if m.err != "" {
		s += "error: " + m.err + "\n\n"
	}
	if m.done {
		s += "(enter/esc) back to list\n"
	} else {
		s += "(enter) run  (esc) cancel\n"
	}
	return s
}
