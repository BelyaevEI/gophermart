package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	ResponseData struct {
		Status int
		Size   int
	}

	// добавляем реализацию http.ResponseWriter
	LoggResponse struct {
		Writer   http.ResponseWriter
		RespData *ResponseData
	}

	Logger struct {
		Log zap.SugaredLogger
	}
)

func (r LoggResponse) Header() http.Header {
	return r.Writer.Header()
}

func (r LoggResponse) Write(b []byte) (int, error) {

	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.Writer.Write(b)
	r.RespData.Size += size
	return size, err
}

func (r LoggResponse) WriteHeader(statusCode int) {

	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.Writer.WriteHeader(statusCode)
	r.RespData.Status = statusCode // захватываем код статуса
}

func New() (*Logger, error) {
	// создаём предустановленный регистратор zap
	logg, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	defer logg.Sync()

	// делаем регистратор SugaredLogger
	sugar := *logg.Sugar()

	return &Logger{Log: sugar}, nil
}

func (l *Logger) Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseData := &ResponseData{
			Status: 0,
			Size:   0,
		}

		lw := LoggResponse{
			Writer:   w,
			RespData: responseData,
		}

		//Время запуска
		start := time.Now()

		// эндпоинт
		uri := r.RequestURI

		// метод запроса
		method := r.Method

		// обслуживание оригинального запроса
		// внедряем реализацию http.ResponseWriter
		h.ServeHTTP(&lw, r)

		//время выполнения
		duration := time.Since(start)

		// отправляем сведения о запросе в zap
		l.Log.Infoln(
			"uri", uri,
			"method", method,
			"status", responseData.Status,
			"duration", duration,
			"size", responseData.Size,
		)

	})
}
