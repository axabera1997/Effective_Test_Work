package middlas

import (
	"emobile/internal/models"
	"net/http"
	"time"
)

// ErrorLoggerMiddleware - middleware для логирования ошибок HTTP
func ErrorLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем обертку для ResponseWriter, чтобы перехватывать ошибки
		lrw := &logErrResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(lrw, r)
	})
}

type logErrResponseWriter struct {
	http.ResponseWriter
	request      *http.Request
	errorMessage string
	statusCode   int
	wroteHeader  bool // Флаг, чтобы отслеживать, был ли вызван WriteHeader
}

// WriteHeader перехватывает статус код ответа
func (lrw *logErrResponseWriter) WriteHeader(code int) {
	if !lrw.wroteHeader {
		lrw.statusCode = code
		lrw.wroteHeader = true
		lrw.ResponseWriter.WriteHeader(code)
	}
}

// Write перехватывает запись тела ответа
func (lrw *logErrResponseWriter) Write(b []byte) (int, error) {
	if !lrw.wroteHeader {
		lrw.WriteHeader(http.StatusOK) // Стандартный статус, если WriteHeader не вызван
	}

	// Если статус указывает на ошибку, сохраняем сообщение
	if lrw.statusCode >= 400 {
		lrw.errorMessage = string(b)
	}

	n, err := lrw.ResponseWriter.Write(b)

	// Логируем ошибку после записи (если она есть)
	if lrw.statusCode >= 400 && lrw.errorMessage != "" {
		lrw.logError()
	}

	return n, err
}

// logError логирует ошибку
func (lrw *logErrResponseWriter) logError() {
	if lrw.statusCode >= 400 && lrw.errorMessage != "" {

		models.Logger.Error("Ошибка вышла",
			"Time", time.Now().Format(time.RFC3339),
			"Method", lrw.request.Method,
			"uri", lrw.request.URL.Path,
			"code", lrw.statusCode,
			"message", lrw.errorMessage,
			"RemoteAddr", lrw.request.RemoteAddr,
			"UserAgent", lrw.request.UserAgent(),
		)

	}
}
