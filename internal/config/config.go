package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// IsFirstRun returns true if no config.yaml exists yet (never initialized).
func IsFirstRun() bool {
	_, err := os.Stat(configPath())
	return errors.Is(err, os.ErrNotExist)
}

type Config struct {
	HomeDir            string   `mapstructure:"homeDir" yaml:"homeDir"`
	StoreDir           string   `mapstructure:"storeDir" yaml:"storeDir"`
	CacheDir           string   `mapstructure:"cacheDir" yaml:"cacheDir"`
	IndexDir           string   `mapstructure:"indexDir" yaml:"indexDir"`
	StatsDB            string   `mapstructure:"statsDb" yaml:"statsDb"`
	RegistryURL        string   `mapstructure:"registryUrl" yaml:"registryUrl"`
	TargetCLIs         []string `mapstructure:"targetClis" yaml:"targetClis"`
	DefaultSkills      []string `mapstructure:"defaultSkills" yaml:"defaultSkills"`
	AutoUpdateTrending bool     `mapstructure:"autoUpdateTrending" yaml:"autoUpdateTrending"`
}

func defaultHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".pskill")
}

func defaults() Config {
	home := defaultHome()
	return Config{
		HomeDir:            home,
		StoreDir:           filepath.Join(home, "store"),
		CacheDir:           filepath.Join(home, "cache"),
		IndexDir:           filepath.Join(home, "index"),
		StatsDB:            filepath.Join(home, "stats.db"),
		RegistryURL:        "https://skillsmp.com",
		TargetCLIs:         []string{"cursor", "claude", "codex"},
		DefaultSkills:      []string{},
		AutoUpdateTrending: true,
	}
}

func configPath() string {
	return filepath.Join(defaultHome(), "config.yaml")
}

func LoadGlobal() (Config, error) {
	d := defaults()
	_ = os.MkdirAll(d.HomeDir, 0o755)
	_ = os.MkdirAll(d.StoreDir, 0o755)
	_ = os.MkdirAll(d.CacheDir, 0o755)
	_ = os.MkdirAll(d.IndexDir, 0o755)

	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetConfigType("yaml")
	setDefaults(v, d)

	if _, err := os.Stat(configPath()); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveGlobal(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(configPath()), 0o755); err != nil {
		return err
	}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), raw, 0o644)
}

func setDefaults(v *viper.Viper, cfg Config) {
	v.Set("homeDir", cfg.HomeDir)
	v.Set("storeDir", cfg.StoreDir)
	v.Set("cacheDir", cfg.CacheDir)
	v.Set("indexDir", cfg.IndexDir)
	v.Set("statsDb", cfg.StatsDB)
	v.Set("registryUrl", cfg.RegistryURL)
	v.Set("targetClis", cfg.TargetCLIs)
	v.Set("defaultSkills", cfg.DefaultSkills)
	v.Set("autoUpdateTrending", cfg.AutoUpdateTrending)
}
