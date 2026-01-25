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
    home, err := os.UserHomeDir()
    if err != nil {
        // Fallback to current directory if home cannot be determined
        return filepath.Join(".", ".config", "lazydots", "config.json")
    }
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
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        return err
    }
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(Path(), data, 0644)
}
