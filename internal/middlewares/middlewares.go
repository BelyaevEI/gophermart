package middlewares

import (
	"net/http"

	"github.com/BelyaevEI/gophermart/internal/cookies"
	"github.com/BelyaevEI/gophermart/internal/userservice"
)

type Middlewares struct {
	UserService *userservice.UserService
}

func New(userService *userservice.UserService) *Middlewares {
	return &Middlewares{UserService: userService}
}

func (m *Middlewares) Authorization(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		ctx := request.Context()

		cookie, err := request.Cookie("Token")
		if err != nil {
			m.UserService.Log.Log.Error(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		key, err := m.UserService.GetKey(ctx, cookie.Value)
		if err != nil || len(key) == 0 {
			m.UserService.Log.Log.Error(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if ok := cookies.Validation(cookie.Value, key); !ok {
			m.UserService.Log.Log.Info("unauthorized user")
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(writer, request.WithContext(ctx))
	})

}
