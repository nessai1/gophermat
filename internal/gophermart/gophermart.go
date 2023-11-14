package gophermart

import (
	"fmt"
	"github.com/nessai1/gophermat/internal/config"
	"github.com/nessai1/gophermat/internal/logger"
	"go.uber.org/zap"
	"net/http"

	"github.com/go-chi/chi"
)

func Start() error {
	router := chi.NewRouter()
	cfg := config.GetConfig()

	log, err := logger.NewLogger(cfg.EnvType)
	if err != nil {
		return fmt.Errorf("cannot initialize logger: %w", err)
	}

	log.Info("starting service", zap.String("service address", cfg.ServiceAddr))

	if err := http.ListenAndServe(cfg.ServiceAddr, router); err != nil {
		return err
	}

	return nil
}
