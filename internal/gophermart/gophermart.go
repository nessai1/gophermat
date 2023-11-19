package gophermart

import (
	"fmt"
	"github.com/nessai1/gophermat/internal/config"
	"github.com/nessai1/gophermat/internal/database"
	"github.com/nessai1/gophermat/internal/handler"
	"github.com/nessai1/gophermat/internal/logger"
	"github.com/nessai1/gophermat/internal/user"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func Start() error {
	router := chi.NewRouter()
	cfg := config.GetConfig()

	log, err := logger.NewLogger(cfg.EnvType)
	if err != nil {
		return fmt.Errorf("cannot initialize logger on start service: %w", err)
	}

	db, err := database.InitSQLDriverByConnectionURI(cfg.DBConnectionStr)
	if err != nil {
		return fmt.Errorf("cannot initialize database on start service: %w", err)
	}

	authHandler := handler.AuthHandler{
		Logger:         log,
		UserController: user.NewController(user.CreatePGXRepository(db)),
	}

	authMux := chi.NewMux()
	authMux.HandleFunc("/api/user/register", authHandler.HandleRegisterUser)
	authMux.HandleFunc("/api/user/login", authHandler.HandleAuthUser)

	orderHandler := handler.OrderHandler{}

	orderMux := chi.NewMux()
	orderMux.Use(authHandler.MiddlewareAuthorizeRequest())
	orderMux.HandleFunc("/", orderHandler.HandleGetUserOrders)

	router.Mount("/", authMux)
	router.Mount("/api/user/orders", orderMux)

	log.Info("starting service", zap.String("service address", cfg.ServiceAddr))

	if err := http.ListenAndServe(cfg.ServiceAddr, router); err != nil {
		return err
	}

	return nil
}
