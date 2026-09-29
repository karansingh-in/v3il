# v3il

**A local file and folder encryption tool, built in Go from cryptographic primitives — with a CLI and an interactive TUI.**

No cloud, no telemetry, no server. Everything is encrypted with a key derived from your own password, using Argon2id for key derivation and AES-256-GCM for authenticated encryption.

```
v3il encrypt taxes.pdf      →  taxes.pdf.v3il
v3il decrypt taxes.pdf.v3il  →  taxes.pdf
```

It also ships with a small encrypted journal, built on the same crypto core, for anyone who wants a private place to write that isn't a plaintext file on disk.

---

## Why

Most "quickly encrypt a file" workflows are either an afterthought bolted onto an archiver (7-Zip's password protection) or a tool with a genuinely rough CLI (GPG, OpenSSL's `enc`). This isn't trying to replace battle-tested, widely-audited tools like [`age`](https://github.com/FiloSottile/age) for people with real security needs — it was built to *understand* modern authenticated encryption by implementing it correctly from primitives, rather than trusting a library without knowing what it's doing underneath. It happens to also be genuinely usable day to day.

## Features

- **File & folder encryption** — `v3il encrypt <path>` / `v3il decrypt <path>.v3il`
- **Encrypted journal** — add, view, edit, and delete private notes, all encrypted at rest
- **Interactive TUI** — arrow-key navigation, add/edit/delete forms, and file encryption, all without memorizing flags (built with [Bubble Tea](https://github.com/charmbracelet/bubbletea))
- **Scriptable CLI** — every feature also works as a single one-shot command, for automation or muscle memory
- **Session auto-lock** — the TUI wipes the derived encryption key from memory after 60 seconds of inactivity
- **Real cryptography, explained, not hand-waved:**
  - **Argon2id** for password → key derivation — memory-hard (I used 64 MB), which specifically raises the cost of *parallel* brute-forcing on GPUs/ASICs (an attacker's core count stops helping once each guess requires significant memory)
  - **AES-256-GCM** for authenticated encryption — a wrong password or a tampered file fails loudly with an explicit error, never silently returns corrupted data
  - A fresh random salt per vault and a fresh random nonce on **every single** encryption operation — nonce reuse under the same key is a well-known, catastrophic GCM failure mode, and this project generates a new one every time, no exceptions

## Installation

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/karansingh-in/v3il.git
cd v3il
go build -o v3il
```
On Windows this produces `v3il.exe`.

**Cross-compiling:**
```bash
GOOS=windows GOARCH=amd64 go build -o v3il.exe   # Windows
GOOS=darwin  GOARCH=amd64 go build -o v3il         # macOS
GOOS=linux   GOARCH=amd64 go build -o v3il         # Linux
```

## Usage

### TUI (recommended)
```bash
v3il
```
Launches an interactive terminal interface — navigate your journal with arrow keys, add/edit/delete entries, encrypt or decrypt files, all in one running session behind a single password prompt.

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate entries |
| `enter` | Open selected entry |
| `a` | Add a new entry |
| `e` | Edit selected entry |
| `d` | Delete selected entry |
| `f` | Encrypt a file |
| `x` | Decrypt a file |
| `esc` | Back / cancel |
| `ctrl+c` | Quit |

### CLI (scripting / one-shot commands)
```bash
v3il encrypt <path>            # encrypt any file → <path>.v3il
v3il decrypt <path>.v3il        # decrypt it back

v3il add                        # add a journal entry
v3il get <name>                  # view an entry
v3il list                        # list all entries
v3il update <name>               # edit an entry
v3il delete <name>               # delete an entry
```
Every command prompts for your master password (hidden input, never echoed to the terminal). The journal vault lives at `~/.v3il/vault`.

## Architecture

```
v3il/
├── main.go          CLI entry point, command routing, password prompts
├── vault.go           Entry struct, in-memory CRUD logic
├── crypto.go           Argon2id key derivation, AES-GCM encrypt/decrypt
├── storage.go           Encrypted vault read/write, on-disk file format
├── file_crypto.go        Arbitrary file encryption/decryption (CLI path)
├── tui.go                TUI shell: screen routing, session state, auto-lock
├── tui_unlock.go           Password/unlock screen
├── tui_list.go              Journal entry list, navigation
├── tui_detail.go             Single entry view
├── tui_form.go                Add/edit form (multiline body input)
└── tui_fileop.go               In-TUI file encrypt/decrypt
```

**On-disk format** (vault file and encrypted files alike): `[salt][nonce][ciphertext + auth tag]` — fully self-contained, so the file can be renamed or moved without breaking decryption.

## Security model

**Protects against:**
- Someone obtaining your vault or an encrypted file (stolen device, leaked backup) — without the password, it's unrecoverable ciphertext
- Tampering — AES-GCM's authentication tag means a modified file fails to decrypt instead of silently returning corrupted data

**Does not protect against:**
- A keylogger or malware running on your machine while you type your password
- Anything reading live process memory while a session is unlocked
- A weak password — Argon2id slows down brute-forcing, it does not turn a weak password into a strong one

This is a personal project built to deeply understand the cryptography involved, not an independently audited security product. Read the code, and use it with that in mind.


## License
MIT