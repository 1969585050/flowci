package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/flowci/flowci/internal/api"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				writeError(w, http.StatusInternalServerError, api.CodeInternalError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, code api.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.Error(code, message))
}
