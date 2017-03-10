package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type config struct {
	GtdFile   string `toml:"gtdfile"`
	MemoDir   string `toml:"memodir"`
	OutputDir string `toml:"outputdir"`
	FilterCmd string `toml:"filtercmd"`
	Editor    string `toml:"editor"`
}

func (cfg *config) load() error {
	var dir string
	dir = filepath.Join(os.Getenv("HOME"), ".config", "gtd")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create directory: %v", err)
	}

	configfile := filepath.Join(dir, "config.toml")
	_, err := os.Stat(configfile)
	// if file exists in ~/.config/gtd/config.toml, decode file and return nil
	if err == nil {
		_, err := toml.DecodeFile(configfile, cfg)
		if err != nil {
			return fmt.Errorf("cannot decode toml file: %v", err)
		}
		return nil
	}

	cfg.Editor = "vi"
	gtdfile := filepath.Join(os.Getenv("HOME"), "gtd.json")
	cfg.GtdFile = gtdfile
	f, err := os.Create(configfile)
	// if file not exists in ~/.config/gtd/config.toml, write f with cfg
	return toml.NewEncoder(f).Encode(cfg)
}
