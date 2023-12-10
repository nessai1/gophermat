package logger

import (
	"github.com/nessai1/gophermat/internal/config"

	"go.uber.org/zap"
)

func NewLogger(envType config.EnvType) (*zap.Logger, error) {
	logger, err := createZapLogger(envType)

	return logger, err
}
