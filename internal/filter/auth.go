package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"fibank.bg/fis-gateway-ws/internal/model"
	"fibank.bg/fis-gateway-ws/internal/util"
)

type contextKey string

const userContextKey contextKey = "userData"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		appSoftAuthToken := r.Header.Get("X-Soft-Auth-Token")
		if appSoftAuthToken == "" {
			util.ErrorJSON(w, http.StatusUnauthorized, "Missing soft auth token")
			return
		}

		appSessionToken, err := r.Cookie("AppSessionToken")
		if err != nil {
			util.ErrorJSON(w, http.StatusUnauthorized, "Missing session token")
			return
		}

		if !util.ValidateSoftAuthToken(appSoftAuthToken, r.UserAgent()) {
			util.ErrorJSON(w, http.StatusUnauthorized, "Invalid soft auth token")
			return
		}
		fmt.Println(appSessionToken)
		userData, exists := util.UserCache.Get(appSessionToken.Value)
		if !exists {
			util.ErrorJSON(w, http.StatusUnauthorized, "Invalid session token")
			return
		}

		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, "Failed to process user data")
			return
		}

		r.Header.Set("USER-METADATA-HEADER", string(userDataJSON))
		ctx := context.WithValue(r.Context(), userContextKey, userData)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserDataFromContext(r *http.Request) (model.UserData, bool) {
	userData, ok := r.Context().Value(userContextKey).(model.UserData)
	return userData, ok
}
