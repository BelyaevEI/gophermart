package route

import (
	"github.com/BelyaevEI/gophermart/internal/middlewares"
	"github.com/BelyaevEI/gophermart/internal/orderservice"
	"github.com/BelyaevEI/gophermart/internal/userservice"
	"github.com/go-chi/chi"
)

func New(userservice *userservice.UserService, orderservice *orderservice.OrderService, middleware *middlewares.Middlewares) *chi.Mux {

	route := chi.NewRouter()

	// Handlers
	route.Post("/api/user/register", userservice.Registration)                                // Registration user
	route.Post("/api/user/login", userservice.Auth)                                           // Authorization user
	route.Post("/api/user/orders", middleware.Authorization(orderservice.UploadOrder))        // Upload number of order
	route.Get("/api/user/orders", middleware.Authorization(orderservice.GetOrdersUser))       // Get upload list orders
	route.Get("/api/user/orders", middleware.Authorization(userservice.GetBalance))           // Get balance user
	route.Post("/api/user/balance/withdraw", middleware.Authorization(userservice.Withdrawn)) // Withdrawn
	route.Get("/api/user/withdrawals", middleware.Authorization(userservice.GetAllWithdrawn)) // Get all withdrawn
	return route
}
