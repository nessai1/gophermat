package gophermart

import (
	"github.com/nessai1/gophermat/internal/config"
	"net/http"

	"github.com/go-chi/chi"
)

func Start() error {
	router := chi.NewRouter()
	cfg := config.GetConfig()

	if err := http.ListenAndServe(cfg.ServiceAddr, router); err != nil {
		return err
	}

	return nil
}
