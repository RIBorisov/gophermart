package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/RIBorisov/gophermart/internal/logger"
)

type Service struct {
	RunAddress            string        `env:"RUN_ADDRESS" envDefault:"localhost:8089"`
	AccrualSystemAddress  string        `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
	AccrualOrderInfoRoute string        `env:"ACCRUAL_ORDER_INFO_ROUTE" envDefault:"/api/orders/{orderID}"`
	DatabaseDSN           string        `env:"DATABASE_URI" envDefault:""`
	AccrualPollInterval   time.Duration `env:"ACCRUAL_POLL_INTERVAL" envDefault:"10s"`
}

type Secret struct {
	SecretKey string `env:"SECRET_KEY,unset" envDefault:"Qpm9^vmz13@ja"`
}

type Config struct {
	Secret  Secret
	Service Service
}

func LoadConfig(_ *logger.Log) (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed parse env: %w", err)
	}

	flags := parseFlags()
	if flags.RunAddress != "" {
		cfg.Service.RunAddress = flags.RunAddress
	}
	if flags.AccrualSystemAddress != "" {
		cfg.Service.AccrualSystemAddress = flags.AccrualSystemAddress
	}
	if flags.DatabaseDSN != "" {
		cfg.Service.DatabaseDSN = flags.DatabaseDSN
	} else if cfg.Service.DatabaseDSN == "" {
		return nil, errors.New("failed read DATABASE_URI value")
	}

	return cfg, nil
}
