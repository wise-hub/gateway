package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"fibank.bg/fis-gateway-ws/internal/repository"
	"fibank.bg/fis-gateway-ws/internal/util"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MinimalUserData struct {
	UserID        int      `json:"user_id"`
	SoftAuthToken string   `json:"soft_auth_token"`
	Username      string   `json:"username"`
	Roles         []string `json:"roles"`
}

func LoginHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			User string `json:"u"`
			Pass string `json:"p"`
			CSRF string `json:"csrf"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.ErrorJSON(w, http.StatusBadRequest, "Invalid request")
			return
		}
	
		if !util.ValidateSoftAuthToken(req.CSRF, r.UserAgent()) {
			util.ErrorJSON(w, http.StatusUnauthorized, "Invalid CSRF token")
			return
		}
	
		userData, err := repository.GetUserDataFromDB(db, req.User, req.Pass)
		if err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve user data")
			return
		}

		softAuthToken := util.GenerateSoftAuthToken(r.UserAgent())
		sessionToken, err := util.GenerateUniqueToken()
		if err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, "Failed to generate session token")
			return
		}

		userData.SoftAuthToken = softAuthToken
		userData.Token = sessionToken
		userData.ExpiresAt = time.Now().Add(1 * time.Hour)
		util.UserCache.Set(userData.Token, *userData)

		http.SetCookie(w, &http.Cookie{
			Name:     "AppSessionToken",
			Value:    userData.Token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		minimalUserData := MinimalUserData{
			UserID:        userData.UserID,
			SoftAuthToken: userData.SoftAuthToken,
			Username:      userData.Username,
			Roles:         userData.Roles,
		}

		util.JSON(w, http.StatusOK, util.H{
			"message":  "Login successful",
			"userData": minimalUserData,
		})
	}
}
