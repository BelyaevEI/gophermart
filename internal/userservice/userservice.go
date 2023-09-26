package userservice

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/BelyaevEI/gophermart/internal/cookies"
	"github.com/BelyaevEI/gophermart/internal/logger"
	"github.com/BelyaevEI/gophermart/internal/models"
	"github.com/BelyaevEI/gophermart/internal/orderepository"
	"github.com/BelyaevEI/gophermart/internal/userepository"
	"github.com/BelyaevEI/gophermart/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	user  *userepository.User
	order *orderepository.Order
	Log   *logger.Logger
}

func New(user *userepository.User, order *orderepository.Order, log *logger.Logger) *UserService {
	return &UserService{
		user:  user,
		order: order,
		Log:   log}
}

func (userservice *UserService) Registration(writer http.ResponseWriter, r *http.Request) {

	var (
		buf     bytes.Buffer
		regInfo models.Registration
	)

	ctx := r.Context()

	//Read body request
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Deserializing JSON
	if err = json.Unmarshal(buf.Bytes(), &regInfo); err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//Check unique login
	if unique := userservice.user.CheckUserLogin(ctx, regInfo.Login); !unique {
		writer.WriteHeader(http.StatusConflict)
		return
	}

	//Create unique ID
	userID := utils.GenerateUniqueID()

	//Create secret key
	key := utils.GenerateRandomString(7)

	token, err := cookies.NewCookie(writer, userID, key)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Save user in system
	err = userservice.user.SaveUser(ctx, regInfo.Login, regInfo.Password, key, token)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Authorization user
	err = userservice.user.Authorization(ctx, regInfo.Login, regInfo.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			userservice.Log.Log.Error(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// User registered and authorized
	writer.WriteHeader(http.StatusOK)
}

func (userservice *UserService) Auth(writer http.ResponseWriter, request *http.Request) {

	var (
		buf     bytes.Buffer
		regInfo models.Registration
	)

	ctx := request.Context()

	// Read body request
	_, err := buf.ReadFrom(request.Body)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Deserializing JSON
	if err = json.Unmarshal(buf.Bytes(), &regInfo); err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = userservice.user.Authorization(ctx, regInfo.Login, regInfo.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			userservice.Log.Log.Error(err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (userservice *UserService) GetKey(ctx context.Context, token string) (string, error) {
	key, err := userservice.user.GetKey(ctx, token)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (userservice *UserService) GetBalance(writer http.ResponseWriter, request *http.Request) {

	var balance models.Balance

	ctx := request.Context()

	cookie, err := request.Cookie("Token")
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user accrual
	accrual, err := userservice.user.GetAccrualUser(ctx, cookie.Value)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Get user withdrawn
	withdrawn, err := userservice.user.GetWithdrawnUser(ctx, cookie.Value)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Struct of response
	balance = models.Balance{
		Accrual:   accrual,
		Withdrawn: withdrawn,
	}

	// Lets form a response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	encode := json.NewEncoder(writer)

	if err := encode.Encode(balance); err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (userservice *UserService) Withdrawn(writer http.ResponseWriter, request *http.Request) {

	var (
		buf       bytes.Buffer
		withdrawn models.Withdrawn
	)

	ctx := request.Context()

	cookie, err := request.Cookie("Token")
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Read body request
	_, err = buf.ReadFrom(request.Body)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Deserializing JSON
	if err = json.Unmarshal(buf.Bytes(), &withdrawn); err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	order := strconv.Itoa(withdrawn.Order)

	// Check with Luhn alrgotim
	if ok := utils.CheckOrder([]byte(order)); !ok {
		userservice.Log.Log.Error("Wrong number of order")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Get loyalty score user
	scores, err := userservice.user.GetUSerScore(ctx, cookie.Value)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check correct sum
	if scores < withdrawn.Sum {
		writer.WriteHeader(http.StatusPaymentRequired)
		return
	}

	// Upload new order to system
	err = userservice.order.Upload2System(ctx, withdrawn.Order, cookie.Value)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update withdrawn in system
	err = userservice.user.UpdateWithdrawn(ctx, cookie.Value, withdrawn.Sum, withdrawn.Order)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	updateScore := scores - withdrawn.Sum

	// Update user score
	err = userservice.user.UpdateScores(ctx, cookie.Value, updateScore)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (userservice *UserService) GetAllWithdrawn(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	cookie, err := request.Cookie("Token")
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get all withdrawn
	allWithdrawn, err := userservice.user.GetAllWithdrawn(ctx, cookie.Value)
	if err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Not records withdrawn
	if allWithdrawn == nil {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Lets form a response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	encode := json.NewEncoder(writer)

	if err := encode.Encode(allWithdrawn); err != nil {
		userservice.Log.Log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
