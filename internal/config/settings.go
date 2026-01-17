package config

import (
	"flag"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	RunAddress     string
	AccuralAddress string
	DatabaseInfo   string
	AuthSecret     []byte
	AuthTokenLife  time.Duration
)

type envConfig struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	AccuralAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseInfo   string `env:"DATABASE_URI"`
	AuthSecret     []byte `env:"AUTH_SECRET"`
}

func Load() {
	flag.StringVar(&RunAddress, "a", ":8080", "server address")
	flag.StringVar(&DatabaseInfo, "d", "", "database dsn")
	flag.StringVar(&AccuralAddress, "r", "", "accural system address")
	flag.StringVar(&AccuralAddress, "s", "SECRET", "auth secret")

	flag.Parse()

	var e envConfig
	if err := cleanenv.ReadEnv(&e); err != nil {
		log.Fatalf("config error: %v", err)
	}

	if e.RunAddress != "" {
		RunAddress = e.RunAddress
	}
	if e.DatabaseInfo != "" {
		DatabaseInfo = e.DatabaseInfo
	}
	if e.AccuralAddress != "" {
		AccuralAddress = e.AccuralAddress
	}
	if len(e.AuthSecret) > 0 {
		AuthSecret = e.AuthSecret
	}

	AuthTokenLife = 24 * time.Hour
}
