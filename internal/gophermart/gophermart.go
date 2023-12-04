package gophermart

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/nessai1/gophermat/internal/config"
	"github.com/nessai1/gophermat/internal/database"
	"github.com/nessai1/gophermat/internal/handler"
	"github.com/nessai1/gophermat/internal/intransaction"
	"github.com/nessai1/gophermat/internal/logger"
	"github.com/nessai1/gophermat/internal/order"
	"github.com/nessai1/gophermat/internal/user"
	"github.com/nessai1/gophermat/internal/zip"
	"go.uber.org/zap"
	"net/http"
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

	router.Use(zip.GetZipMiddleware(log))

	authMux := chi.NewMux()
	authMux.Post("/api/user/register", authHandler.HandleRegisterUser)
	authMux.Post("/api/user/login", authHandler.HandleAuthUser)

	transaction := intransaction.NewPGXTransaction(db)

	enrollmentController := order.NewEnrollmentController(cfg.AccrualServiceAddr, order.CreatePGXEnrollmentRepository(db), userController)
	ch, err := order.StartEnrollmentWorker(userController, enrollmentController, log, cfg.AccrualServiceAddr, transaction)
	if err != nil {
		return fmt.Errorf("error while starting enrollment worker: %w", err)
	}

	enrollmentController.EnrollmentCh = ch

	enrollmentHandler := handler.EnrollmentOrderHandler{
		Logger:               log,
		EnrollmentController: enrollmentController,
	}
	enrollmentMux := chi.NewMux()
	enrollmentMux.Use(authHandler.MiddlewareAuthorizeRequest())
	enrollmentMux.Post("/", enrollmentHandler.HandleLoadOrders)
	enrollmentMux.Get("/", enrollmentHandler.HandleGetOrders)

	balanceHandler := handler.BalanceHandler{
		Logger:             log,
		WithdrawController: order.NewWithdrawController(order.NewPGXWithdrawRepository(db), userController, transaction),
	}

	balanceMux := chi.NewMux()
	balanceMux.Use(authHandler.MiddlewareAuthorizeRequest())
	balanceMux.Get("/", balanceHandler.HandleGetBalance)
	balanceMux.Post("/withdraw", balanceHandler.HandleAddWithdraw)

	withdrawInfoMux := chi.NewMux()
	withdrawInfoMux.Use(authHandler.MiddlewareAuthorizeRequest())
	withdrawInfoMux.Get("/", balanceHandler.HandleGetListWithdraw)

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
