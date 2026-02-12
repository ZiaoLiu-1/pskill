package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Version info set at build time via cli package.
var appVersion = "dev"

// SetVersion stores the app version.
func SetVersion(v string) { appVersion = v }

// GetVersion returns the app version.
func GetVersion() string { return appVersion }

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
	RegistryAPIKey     string   `mapstructure:"registryApiKey" yaml:"registryApiKey"`
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
	_ = os.MkdirAll(filepath.Dir(configPath()), 0o755)

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
	// Ensure config dirs exist (use loaded config, not defaults)
	_ = os.MkdirAll(cfg.StoreDir, 0o755)
	_ = os.MkdirAll(cfg.CacheDir, 0o755)
	_ = os.MkdirAll(cfg.IndexDir, 0o755)
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
	v.SetDefault("homeDir", cfg.HomeDir)
	v.SetDefault("storeDir", cfg.StoreDir)
	v.SetDefault("cacheDir", cfg.CacheDir)
	v.SetDefault("indexDir", cfg.IndexDir)
	v.SetDefault("statsDb", cfg.StatsDB)
	v.SetDefault("registryUrl", cfg.RegistryURL)
	v.SetDefault("registryApiKey", cfg.RegistryAPIKey)
	v.SetDefault("targetClis", cfg.TargetCLIs)
	v.SetDefault("defaultSkills", cfg.DefaultSkills)
	v.SetDefault("autoUpdateTrending", cfg.AutoUpdateTrending)
}
