package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/zalando/go-keyring"
)

const keyringService = "tedi"

type storedSecret struct {
	Token        string `json:"access_token"`
	ClientSecret string `json:"client_secret,omitempty"`
}

func readSecret(s *SessionConfiguration) error {
	raw, err := keyring.Get(keyringService, s.CredentialID)
	if err != nil {
		return err
	}
	var secret storedSecret
	if err := json.Unmarshal([]byte(raw), &secret); err != nil {
		return err
	}
	s.Token = secret.Token
	s.ClientSecret = secret.ClientSecret
	return nil
}

func writeSecret(s *SessionConfiguration) error {
	id := s.CredentialID
	if id == "" {
		b := make([]byte, 16)
		rand.Read(b)
		id = hex.EncodeToString(b)
	}

	payload, err := json.Marshal(storedSecret{Token: s.Token, ClientSecret: s.ClientSecret})
	if err != nil {
		return err
	}
	if err := keyring.Set(keyringService, id, string(payload)); err != nil {
		return err
	}
	s.CredentialID = id
	return nil
}

func deleteSecret(id string) error {
	err := keyring.Delete(keyringService, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
