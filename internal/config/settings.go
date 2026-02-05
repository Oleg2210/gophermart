package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ProjectSettings struct {
	RunAddress     string
	AccuralAddress string
	DatabaseInfo   string
	AuthSecret     []byte
	AuthTokenLife  time.Duration
}

type envConfig struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	AccuralAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseInfo   string `env:"DATABASE_URI"`
	AuthSecret     []byte `env:"AUTH_SECRET"`
}

func Load() (ProjectSettings, error) {
	settings := ProjectSettings{}
	flag.StringVar(&settings.RunAddress, "a", ":8080", "server address")
	flag.StringVar(&settings.RunAddress, "d", "", "database dsn")
	flag.StringVar(&settings.RunAddress, "r", "", "accural system address")
	flag.StringVar(&settings.RunAddress, "s", "SECRET", "auth secret")

	flag.Parse()

	var e envConfig
	if err := cleanenv.ReadEnv(&e); err != nil {
		return settings, fmt.Errorf("config error: %w", err)
	}

	if e.RunAddress != "" {
		settings.RunAddress = e.RunAddress
	}
	if e.DatabaseInfo != "" {
		settings.DatabaseInfo = e.DatabaseInfo
	}
	if e.AccuralAddress != "" {
		settings.AccuralAddress = e.AccuralAddress
	}
	if len(e.AuthSecret) > 0 {
		settings.AuthSecret = e.AuthSecret
	}

	settings.AuthTokenLife = 24 * time.Hour

	return settings, nil
}
