package middleware_custom

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"fibank.bg/fis-gateway-ws/internal/configuration"
)

var logMutex sync.Mutex

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LogMiddleware(deps *configuration.Dependencies) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path += "?" + r.URL.RawQuery
			}
			startTime := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			defer func() {
				if rval := recover(); rval != nil {
					logMutex.Lock()
					defer logMutex.Unlock()

					stack := string(debug.Stack())
					errorMsg := fmt.Sprintf("Panic recovered: %v\nStack Trace:\n%s", rval, stack)

					elapsed := time.Since(startTime)
					logEntry := formatLogEntry(elapsed, http.StatusInternalServerError, r.RemoteAddr, r.Method, path)

					// Log the error and stack trace to the error log
					deps.ErrorLogger.Println(logEntry)
					deps.ErrorLogger.Println(errorMsg)

					// Respond with a generic internal server error
					http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(rw, r)

			logMutex.Lock()
			defer logMutex.Unlock()

			elapsed := time.Since(startTime)
			status := rw.status

			logEntry := formatLogEntry(elapsed, status, r.RemoteAddr, r.Method, path)
			if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/admin") {
				deps.AccessLogger.Println(logEntry)
			}

			if status >= http.StatusInternalServerError {
				deps.ErrorLogger.Println(logEntry)
			}
		})
	}
}

func formatLogEntry(elapsed time.Duration, status int, clientIP, method, path string) string {
	return fmt.Sprintf("| %3d | %6.2fms | %15s | %s %s",
		status,
		float64(elapsed)/float64(time.Millisecond),
		clientIP,
		method,
		path)
}
