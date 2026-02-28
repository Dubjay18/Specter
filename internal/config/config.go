package config

import (
	"github.com/spf13/viper"
)

// Config is the top-level config structure that mirrors specter.yaml.
type Config struct {
	Specter  SpectreConfig  `mapstructure:"specter"`
	Store    StoreConfig    `mapstructure:"store"`
	Sampling SamplingConfig `mapstructure:"sampling"`
}

type SpectreConfig struct {
	Listen       string `mapstructure:"listen"`
	LiveTarget   string `mapstructure:"live_target"`
	ShadowTarget string `mapstructure:"shadow_target"`
	RoutingKey   string `mapstructure:"routing_key"`
}

type StoreConfig struct {
	Backend    string `mapstructure:"backend"`
	BadgerPath string `mapstructure:"badger_path"`
	PostgresDSN string `mapstructure:"postgres_dsn"`
}

type SamplingConfig struct {
	Rate           float64 `mapstructure:"rate"`
	DivergenceOnly bool    `mapstructure:"divergence_only"`
}

// Load reads the config file at the given path and returns a populated Config.
// It also applies sensible defaults for optional fields.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Defaults — applied when a field is missing from the config file.
	v.SetDefault("specter.listen", ":8080")
	v.SetDefault("specter.routing_key", "X-User-ID")
	v.SetDefault("store.backend", "badger")
	v.SetDefault("store.badger_path", "./data/specter")
	v.SetDefault("sampling.rate", 1.0)
	v.SetDefault("sampling.divergence_only", false)

	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}