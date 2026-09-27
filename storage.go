package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const saltSize = 32
const nonceSize = 12

func SaveVault(v *Vault, password []byte, path string) error {
	// converting go structure to json file
	data, err := json.Marshal(v.entries)
	if err != nil {
		return err
	}

	// deriving key and encryption the json data using master password and stored salt
	key := DeriveKey(password, v.salt)

	ciphertext, nonce, err := Encrypt(key, data)
	if err != nil {
		return err
	}

	var metadata []byte
	metadata = append(metadata, v.salt...)
	metadata = append(metadata, nonce...)
	metadata = append(metadata, ciphertext...)

	return os.WriteFile(path, metadata, 0600)

}

func LoadVault(password []byte, path string) (*Vault, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			newSalt, err := GenerateSalt()
			if err != nil {
				return nil, err
			}
			return &Vault{entries: make(map[string]Entry), salt: newSalt}, nil
		}
		return nil, err
	}

	if len(data) < saltSize+nonceSize {
		return nil, fmt.Errorf("file is too small to be a valid .v3il encrypted file")
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key := DeriveKey(password, salt)

	plaintext, err := Decrypt(key, nonce, ciphertext)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]Entry)
	if err := json.Unmarshal(plaintext, &entries); err != nil {
		return nil, err
	}

	return &Vault{entries: entries, salt: salt}, nil
}

func saveVaultWithKey(v *Vault, key []byte, path string) error {
	data, err := json.Marshal(v.entries)
	if err != nil {
		return err
	}

	ciphertext, nonce, err := Encrypt(key, data)
	if err != nil {
		return err
	}

	var out []byte
	out = append(out, v.salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return os.WriteFile(path, out, 0600)
}
