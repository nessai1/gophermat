package config

import (
	"flag"
	"github.com/joho/godotenv"
	"os"
)

type EnvType string

const (
	EnvTypeDevelopment EnvType = "development"
	EnvTypeStage       EnvType = "stage"
	EnvTypeProduction  EnvType = "production"
)

type Config struct {
	ServiceAddr        string
	AccrualServiceAddr string
	DBConnectionStr    string
	SecretKey          string
	EnvType            EnvType
}

func GetConfig() *Config {
	return fetchConfig()
}

func fetchConfig() *Config {

	serviceAddr := flag.String("a", "", "Address of service")
	databaseConnection := flag.String("d", "", "Database connection uri")
	accrualServiceAddr := flag.String("r", "", "Accrual service url")

	flag.Parse()
	godotenv.Load() // May not have .env

	if serviceAddrEnv := os.Getenv("RUN_ADDRESS"); serviceAddrEnv != "" {
		*serviceAddr = serviceAddrEnv
	}

	if databaseConnectionEnv := os.Getenv("DATABASE_URI"); databaseConnectionEnv != "" {
		*databaseConnection = databaseConnectionEnv
	}

	if accrualServiceAddrEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualServiceAddrEnv != "" {
		*accrualServiceAddr = accrualServiceAddrEnv
	}

	var envType EnvType
	envTypeStr := os.Getenv("ENV_TYPE")
	if envTypeStr == "" {
		envType = EnvTypeProduction
	} else {
		envType = EnvType(envTypeStr)
	}

	return &Config{
		ServiceAddr:        *serviceAddr,
		AccrualServiceAddr: *accrualServiceAddr,
		DBConnectionStr:    *databaseConnection,
		SecretKey:          os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
		EnvType:            envType,
	}
}
