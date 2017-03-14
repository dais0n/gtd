package gtd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	return toml.NewEncoder(f).Encode(cfg)
}

func (cfg *config) runcmd(command string, files ...string) error {
	var args []string
	for _, file := range files {
		args = append(args, fmt.Sprintf("%q", file))
	}
	cmdargs := strings.Join(args, " ")
	command += " " + cmdargs

	var cmd *exec.Cmd
	cmd = exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
