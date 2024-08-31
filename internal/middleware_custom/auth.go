package middleware_custom

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"fibank.bg/fis-gateway-ws/internal/util"
)

type contextKey string

const userContextKey contextKey = "userData"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		userData, exists := util.UserCache.Get(token)
		if !exists {
			util.ErrorJSON(w, http.StatusUnauthorized, "Not Authorized")
			return
		}

		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		r.Header.Set("USER-METADATA-HEADER", string(userDataJSON))

		ctx := context.WithValue(r.Context(), userContextKey, userData)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserDataFromContext(r *http.Request) (util.UserData, bool) {
	userData, ok := r.Context().Value(userContextKey).(util.UserData)
	return userData, ok
}

// move to login handler
type LoginRequest struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func LoginAction(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := loginReq.User
	password := loginReq.Pass

	if userID == "" || password == "" {
		http.Error(w, "User and password are required", http.StatusBadRequest)
		return
	}

	userData, err := getUserDataFromDB(userID)
	if err != nil {
		http.Error(w, "Could not fetch user data", http.StatusInternalServerError)
		return
	}

	util.UserCache.Set(userData.Token, *userData)

	response := util.H{
		"message": "Login successful",
		"user":    userData,
	}
	util.JSON(w, http.StatusOK, response)
}

func getUserDataFromDB(userID string) (*util.UserData, error) {
	return &util.UserData{
		UserID:    userID,
		Token:     "generated_token_from_db",
		Username:  "user123",
		Roles:     []string{"role1", "role2"},
		Accounts:  []string{"acc1", "acc2"},
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}, nil
}
