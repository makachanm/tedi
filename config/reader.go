package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const fileName = "config.json"

func candidatePaths() []string {
	var paths []string
	if env := os.Getenv("TEDI_CONFIG"); env != "" {
		paths = append(paths, env)
	}
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, fileName))
	}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), fileName))
	}
	if p := userConfigPath(); p != "" {
		paths = append(paths, p)
	}
	return paths
}

func userConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "tedi", fileName)
}

func Load(explicitPath string) (Configuration, string, error) {
	paths := candidatePaths()
	if explicitPath != "" {
		paths = []string{explicitPath}
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return Configuration{}, p, fmt.Errorf("read config %s: %w", p, err)
		}

		var cfg Configuration
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Configuration{}, p, fmt.Errorf("parse config %s: %w", p, err)
		}
		if cfg.Session.Type != "" && !cfg.Session.Type.Valid() {
			return Configuration{}, p, fmt.Errorf("config %s: unknown session type %q (expected %q or %q)", p, cfg.Session.Type, Mastodon, Misskey)
		}
		cfg.applyDefaults()

		if cfg.Session.InKeyring() {
			err := readSecret(&cfg.Session)
			switch {
			case errors.Is(err, keyring.ErrNotFound):
				cfg.Session.CredentialID = ""
				cfg.Session.ClearCredentials()
			case err != nil:
				return Configuration{}, p, fmt.Errorf("read access token from the system keyring: %w", err)
			}
		} else if cfg.Session.Token != "" {
			Save(p, &cfg)
		}
		return cfg, p, nil
	}

	savePath := explicitPath
	if savePath == "" {
		savePath = userConfigPath()
	}
	if savePath == "" {
		savePath = fileName
	}

	var cfg Configuration
	cfg.applyDefaults()
	return cfg, savePath, nil
}

func InsecurePermissions(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0o077 != 0
}

func Save(path string, cfg *Configuration) error {
	s := &cfg.Session
	switch {
	case s.Token != "" || s.ClientSecret != "":
		if err := writeSecret(s); err != nil && s.InKeyring() {
			return fmt.Errorf("store access token in the system keyring: %w", err)
		}
	case s.InKeyring():
		if err := deleteSecret(s.CredentialID); err != nil {
			return fmt.Errorf("remove access token from the system keyring: %w", err)
		}
		s.CredentialID = ""
	}

	disk := *cfg
	if disk.Session.InKeyring() {
		disk.Session.Token = ""
		disk.Session.ClientSecret = ""
	}

	data, err := json.MarshalIndent(disk, "", "    ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.json")
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("save config: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("save config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}
