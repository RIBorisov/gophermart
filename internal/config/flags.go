package config

import (
	"flag"
)

var f Service

func parseFlags() *Service {
	if !flag.Parsed() {
		flag.StringVar(&f.RunAddress, "a", "localhost:8089", "Host:port where server running")
		flag.StringVar(&f.DatabaseDSN, "d", "", "Database DSN")
		flag.StringVar(&f.AccrualSystemAddress, "r", ":8080", "Accrual system address")
		flag.Parse()
	}
	return &f
}
