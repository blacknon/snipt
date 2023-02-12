// Copyright (c) 2023 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Conf is global config variable
var Conf Config

// Config is a struct of config
type Config struct {
	General GeneralConfig  `toml:"General"`
	Gist    []GistConfig   `toml:"Gist"`
	GitLab  []GitLabConfig `toml:"GitLab"`
}

// GeneralConfig is a struct of general config
type GeneralConfig struct {
	Editor    string `toml:"editor"`
	SelectCmd string `toml:"selectcmd"`
}

func (generalCfg *GeneralConfig) SetDefault() {
	// Editor
	if generalCfg.Editor == "" {
		generalCfg.Editor = os.Getenv("EDITOR")
		if generalCfg.Editor == "" && runtime.GOOS != "windows" {
			if isCommandAvailable("sensible-editor") {
				generalCfg.Editor = "sensible-editor"
			} else {
				generalCfg.Editor = "vim"
			}
		}
	}

	// SelectCmd
	if generalCfg.SelectCmd == "" {
		generalCfg.SelectCmd = "fzf"
	}
}

// GistConfig is a struct of config for Gist
type GistConfig struct {
	AccessToken string `toml:"access_token"`
}

func (gistCfg *GistConfig) SetDefault() {

}

func (gistCfg *GistConfig) Check() (err error) {
	// Check Empty
	if gistCfg.AccessToken == "" {
		err = fmt.Errorf("")
		return err
	}

	// Check ENV
	return
}

// GitLabConfig is a struct of config for GitLab Snippet
type GitLabConfig struct {
	Url         string `toml:"url"`
	Insecure    bool   `toml:"skip_ssl"`
	AccessToken string `toml:"access_token"`

	// proxy
	Proxy     string `toml:"proxy"`
	ProxyUser string `toml:"proxy_user"`
	ProxyPass string `toml:"proxy_pass"`
}

func (gitlabCfg *GitLabConfig) SetDefault() {
	// Url
	if gitlabCfg.Url == "" {
		gitlabCfg.Url = "https://gitlab.com/api/v4"
	}
}

func (gitlabCfg *GitLabConfig) Check() (err error) {
	// Check Empty
	if gitlabCfg.AccessToken == "" {
		err = fmt.Errorf("")
		return err
	}

	// Check ENV

	return
}

// Load loads a config toml
func (cfg *Config) Load(file string) error {
	// Open file
	_, err := os.Stat(file)

	// Get config data
	if err == nil {
		_, err := toml.DecodeFile(file, cfg)
		if err != nil {
			return err
		}
		return nil
	}

	//
	if !os.IsNotExist(err) {
		return err
	}

	//
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	//
	cfg.General.SetDefault()

	//
	return toml.NewEncoder(f).Encode(cfg)
}

// GetDefaultConfigDir returns the default config directory
func GetDefaultConfigDir() (dir string, err error) {
	// Set dir
	if runtime.GOOS == "windows" {
		val, ok := os.LookupEnv("APPDATA")
		if ok {
			dir = val
		} else {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", "snipt")
		}
		dir = filepath.Join(dir, "snipt")
	} else {
		val, ok := os.LookupEnv("XDG_CONFIG_HOME")
		if ok {
			dir = filepath.Join(val, "snipt")
		} else {
			dir = filepath.Join(os.Getenv("HOME"), ".snipt")
		}
	}

	// Create dir
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create directory: %v", err)
	}

	return dir, nil
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
