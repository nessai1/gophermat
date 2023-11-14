package logger

import (
	"fmt"
	"github.com/nessai1/gophermat/internal/config"
	"go.uber.org/zap"
	"os"

	"go.uber.org/zap/zapcore"
)

func createZapLogger(envType config.EnvType) (*zap.Logger, error) {
	atom := zap.NewAtomicLevel()

	logLevel, err := getZapLogLevelByEnvLevel(envType)
	if err != nil {
		return nil, err
	}

	atom.SetLevel(logLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger, nil
}

func getZapLogLevelByEnvLevel(envType config.EnvType) (zapcore.Level, error) {
	if envType == config.EnvTypeProduction {
		return zapcore.ErrorLevel, nil
	} else if envType == config.EnvTypeStage {
		return zapcore.InfoLevel, nil
	} else if envType == config.EnvTypeDevelopment {
		return zapcore.DebugLevel, nil
	}

	return 0, fmt.Errorf("unexpected EnvType got (%d)", envType)
}
