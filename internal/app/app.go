package app

import (
	"net/http"

	"github.com/BelyaevEI/gophermart/internal/config"
	"github.com/BelyaevEI/gophermart/internal/database"
	"github.com/BelyaevEI/gophermart/internal/logger"
	"github.com/BelyaevEI/gophermart/internal/middlewares"
	"github.com/BelyaevEI/gophermart/internal/orderepository"
	"github.com/BelyaevEI/gophermart/internal/orderservice"
	"github.com/BelyaevEI/gophermart/internal/route"
	"github.com/BelyaevEI/gophermart/internal/userepository"
	"github.com/BelyaevEI/gophermart/internal/userservice"
	"github.com/go-chi/chi"
)

type App struct {
	flagRunAddr string
	route       *chi.Mux
}

func NewApp() (*App, error) {

	// Init logger
	log, err := logger.New()
	if err != nil {
		return nil, err
	}

	// Parse variable environment
	cfg := config.ParseFlags()
	log.Log.Info("Flag ^" + cfg.DBpath)
	// Init DB
	db, err := database.NewConnect(cfg.DBpath)
	if err != nil {
		return nil, err
	}

	// Entity for work with users
	userRepository := userepository.New(db)

	// Entity for work with orders
	orderRepository := orderepository.New(cfg.AccrualPath, db)

	// Init user service
	userService := userservice.New(userRepository, orderRepository, log)

	// Init order service
	orderService := orderservice.New(orderRepository, log)

	// Create middleware
	middleware := middlewares.New(userService)

	// Create route
	route := route.New(userService, orderService, middleware)

	// Infinity check status and upload order in loyalty system
	go orderService.OrderStatusChecker()

	return &App{
		flagRunAddr: cfg.FlagRunAddr,
		route:       route,
	}, err

}

func RunServer() error {

	// Init app
	app, err := NewApp()
	if err != nil {
		return err
	}

	return http.ListenAndServe("0.0.0.0:8080", app.route)
}
