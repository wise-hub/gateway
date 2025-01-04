package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/wise-hub/gateway/internal/repository"
	"github.com/wise-hub/gateway/internal/util"
)

type MinimalUserData struct {
	UserID        int      `json:"user_id"`
	SoftAuthToken string   `json:"soft_auth_token"`
	Username      string   `json:"username"`
	Roles         []string `json:"roles"`
}

func LoginHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := r.Header.Get("X-Csrf-Token")
		if !util.ValidateSoftAuthToken(csrfToken, r.UserAgent()) {
			util.ErrorJSON(w, http.StatusUnauthorized, "Invalid CSRF token")
			return
		}

		var req struct {
			User string `json:"u"`
			Pass string `json:"p"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.ErrorJSON(w, http.StatusBadRequest, "Invalid request")
			return
		}

		usernameRegex := regexp.MustCompile(`^[a-z0-9]{4,10}$`)
		if !usernameRegex.MatchString(req.User) {
			util.ErrorJSON(w, http.StatusBadRequest, "Invalid username")
			return
		}

		passwordRegex := regexp.MustCompile(`^.{6,20}$`)
		if !passwordRegex.MatchString(req.Pass) {
			util.ErrorJSON(w, http.StatusBadRequest, "Invalid password")
			return
		}

		if !util.AllowRequest(req.User) {
			util.ErrorJSON(w, http.StatusBadRequest, "Rate limit exceeded")
			return
		}

		userData, err := repository.GetUserDataFromDB(db, req.User, req.Pass)
		if err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		softAuthToken := util.GenerateSoftAuthToken(r.UserAgent())
		sessionToken := util.GenerateUniqueToken()

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
