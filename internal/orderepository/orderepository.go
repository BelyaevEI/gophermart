package orderepository

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/BelyaevEI/gophermart/internal/database"
	"github.com/BelyaevEI/gophermart/internal/models"
)

type Order struct {
	host     string
	endpoint string
	db       *database.Database
}

func New(host string, db *database.Database) *Order {
	return &Order{
		host:     host,
		endpoint: "/api/orders/",
		db:       db,
	}
}

func (orderepository *Order) Upload2Loyalty(numOrder string) (models.Order, error) {
	var (
		order    models.Order
		response *http.Response
	)

	for {
		response, err := http.Get(orderepository.host + orderepository.host + numOrder)
		if err != nil || response.StatusCode == http.StatusTooManyRequests ||
			response.StatusCode == http.StatusNoContent {

			if response != nil && response.Body != nil {
				response.Body.Close()
			}

			time.Sleep(5 * time.Second)
			continue
		}

		break
	}

	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	if err := json.NewDecoder(response.Body).Decode(&order); err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (orderepository *Order) CheckOrder(ctx context.Context, numOrder int, token string) int {
	response := orderepository.db.CheckOrder(ctx, numOrder, token)
	if response == 1 {
		return http.StatusAccepted
	} else if response == 0 {
		return 0
	}
	return response
}

func (orderepository *Order) Upload2System(ctx context.Context, numOrder int, token string) error {
	return orderepository.db.Upload2System(ctx, numOrder, token)
}

func (orderepository *Order) GetNewOrders2Upload() ([]models.Orders, error) {
	orders, err := orderepository.db.GetNewOrders()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (orderepository *Order) UpdateOrder(order models.Order, token string) error {
	return orderepository.db.UpdateOrder(order, token)
}

func (orderepository *Order) UpdateUserScores(point int, token string) error {
	return orderepository.db.UpdateUserScores(point, token)
}

func (orderepository *Order) GetOrdersUser(ctx context.Context, token string) ([]models.UserOrders, error) {
	listOrders, err := orderepository.db.GetOrdersUser(ctx, token)
	if err != nil {
		return nil, err
	}

	// Sort list descending
	sort.SliceStable(listOrders, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02T15:04:05Z07:00", listOrders[i].Upload)
		timeJ, _ := time.Parse("2006-01-02T15:04:05Z07:00", listOrders[j].Upload)
		return timeI.Before(timeJ)
	})

	return listOrders, nil
}
