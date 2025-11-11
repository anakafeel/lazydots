package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    DotfilesPath string `json:"dotfiles_path"`
}

func Path() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "lazydots", "config.json")
}

func Exists() bool {
    _, err := os.Stat(Path())
    return err == nil
}

func Load() (Config, error) {
    var cfg Config
    data, err := os.ReadFile(Path())
    if err != nil {
        return cfg, err
    }
    err = json.Unmarshal(data, &cfg)
    return cfg, err
}

func Save(cfg Config) error {
    cfgDir := filepath.Dir(Path())
    os.MkdirAll(cfgDir, 0755)
    data, _ := json.MarshalIndent(cfg, "", "  ")
    return os.WriteFile(Path(), data, 0644)
}
