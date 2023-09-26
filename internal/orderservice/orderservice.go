package orderservice

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BelyaevEI/gophermart/internal/logger"
	"github.com/BelyaevEI/gophermart/internal/orderepository"
	"github.com/BelyaevEI/gophermart/internal/utils"
)

type OrderService struct {
	orderepository *orderepository.Order
	Log            *logger.Logger
}

func New(orderepository *orderepository.Order, log *logger.Logger) *OrderService {
	return &OrderService{
		Log:            log,
		orderepository: orderepository,
	}
}

func (orderservice *OrderService) UploadOrder(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	cookie, err := request.Cookie("Token")
	if err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get number of order
	numOrder, err := io.ReadAll(request.Body)
	if err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Empty number
	if len(numOrder) == 0 {
		orderservice.Log.Log.Error("Empty order")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check with Luhn alrgotim
	if ok := utils.CheckOrder(numOrder); !ok {
		orderservice.Log.Log.Error("Wrong number of order")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Convert slice of byte to int
	numOrderInt, err := utils.ConvertByte2Int(numOrder)
	if err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check order not uploaded
	status := orderservice.orderepository.CheckOrder(ctx, numOrderInt, cookie.Value)
	if status == http.StatusOK {
		writer.WriteHeader(http.StatusOK)
		return
	} else if status == http.StatusConflict {
		writer.WriteHeader(http.StatusConflict)
		return
	} else if status == 0 {
		orderservice.Log.Log.Errorln("Error database in check order")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Upload order in database
	if err := orderservice.orderepository.Upload2System(ctx, numOrderInt, cookie.Value); err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusAccepted)

}

func (orderservice *OrderService) OrderStatusChecker() {

	for {
		neworders, err := orderservice.orderepository.GetNewOrders2Upload()
		if err != nil || len(neworders) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, order := range neworders {
			numStr := strconv.Itoa(order.NumOrder)

			answerLoyalty, err := orderservice.orderepository.Upload2Loyalty(numStr)
			if len(answerLoyalty.Order) != 0 || err != nil {
				if err = orderservice.orderepository.UpdateOrder(answerLoyalty, order.Token); err == nil {
					if err = orderservice.orderepository.UpdateUserScores(answerLoyalty.Accrual, order.Token); err != nil {
						continue
					}
				}
			}
		}
	}
}

func (orderservice *OrderService) GetOrdersUser(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	cookie, err := request.Cookie("Token")
	if err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get orders of user
	listOrder, err := orderservice.orderepository.GetOrdersUser(ctx, cookie.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Lets form a response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	encode := json.NewEncoder(writer)

	if err := encode.Encode(listOrder); err != nil {
		orderservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
