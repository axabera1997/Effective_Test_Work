package middlas

import (
	"emobile/internal/models"
	"net/http"
	"time"
)

// responseData структура для logger.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter структура для logger
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

// захватываем размер
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// захватываем statusCode в заголовке
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithHTTPLogging(next http.Handler) http.Handler {
	loggedFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		duration := time.Since(start)

		models.Logger.Debug("Дёрнута ручка",
			"uri", r.URL.Path, // какой именно эндпоинт был дернут
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
			"UserAgent", r.UserAgent(),
		)

		next.ServeHTTP(&lw, r)

	}

	return http.HandlerFunc(loggedFunc)
}
