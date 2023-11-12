package config

import (
	"flag"
	"os"
	"sync"
)

type Config struct {
	ServiceAddr        string
	AccrualServiceAddr string
	DBConnectionStr    string
	SecretKey          string
}

var config *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		config = fetchConfig()
	})

	return config
}

func fetchConfig() *Config {
	serviceAddr := flag.String("a", "", "Address of service")
	databaseConnection := flag.String("d", "", "Database connection uri")
	accrualServiceAddr := flag.String("r", "", "Accrual service url")

	if serviceAddrEnv := os.Getenv("RUN_ADDRESS"); serviceAddrEnv != "" {
		*serviceAddr = serviceAddrEnv
	}

	if databaseConnectionEnv := os.Getenv("DATABASE_URI"); databaseConnectionEnv != "" {
		*databaseConnection = databaseConnectionEnv
	}

	if accrualServiceAddrEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualServiceAddrEnv != "" {
		*accrualServiceAddr = accrualServiceAddrEnv
	}

	return &Config{
		ServiceAddr:        *serviceAddr,
		AccrualServiceAddr: *accrualServiceAddr,
		DBConnectionStr:    *databaseConnection,
		SecretKey:          os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
	}
}
