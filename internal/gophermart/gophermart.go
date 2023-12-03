package gophermart

import (
	"fmt"
	"github.com/nessai1/gophermat/internal/config"
	"github.com/nessai1/gophermat/internal/database"
	"github.com/nessai1/gophermat/internal/handler"
	"github.com/nessai1/gophermat/internal/logger"
	"github.com/nessai1/gophermat/internal/order"
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

	userController := user.NewController(user.CreatePGXRepository(db))
	authHandler := handler.AuthHandler{
		Logger:         log,
		UserController: userController,
	}

	authMux := chi.NewMux()
	authMux.Post("/api/user/register", authHandler.HandleRegisterUser)
	authMux.Post("/api/user/login", authHandler.HandleAuthUser)

	enrollmentController := handler.EnrollmentOrderHandler{
		Logger:               log,
		EnrollmentController: order.NewEnrollmentController(cfg.AccrualServiceAddr, order.CreatePGXEnrollmentRepository(db), userController),
	}
	enrollmentMux := chi.NewMux()
	enrollmentMux.Use(authHandler.MiddlewareAuthorizeRequest())
	enrollmentMux.Post("/", enrollmentController.HandleLoadOrders)
	enrollmentMux.Get("/", enrollmentController.HandleGetOrders)

	balanceController := handler.BalanceHandler{
		Logger:             log,
		WithdrawController: order.NewWithdrawController(order.NewPGXWithdrawRepository(db), userController),
	}

	balanceMux := chi.NewMux()
	balanceMux.Use(authHandler.MiddlewareAuthorizeRequest())
	balanceMux.Get("/", balanceController.HandleGetBalance)
	balanceMux.Post("/withdraw", balanceController.HandleAddWithdraw)

	withdrawInfoMux := chi.NewMux()
	withdrawInfoMux.Use(authHandler.MiddlewareAuthorizeRequest())
	withdrawInfoMux.Get("/", balanceController.HandleGetListWithdraw)

	router.Mount("/", authMux)
	router.Mount("/api/user/orders", enrollmentMux)
	router.Mount("/api/user/balance", balanceMux)
	router.Mount("/api/user/withdrawals", withdrawInfoMux)

	log.Info("starting service", zap.String("service address", cfg.ServiceAddr))

	if err := http.ListenAndServe(cfg.ServiceAddr, router); err != nil {
		return err
	}

	return nil
}
