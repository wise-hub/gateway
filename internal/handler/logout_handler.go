package handler

import (
	"net/http"

	"fibank.bg/fis-gateway-ws/internal/util"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	appSessionCookie, err := r.Cookie("AppSessionToken")
	if err != nil || appSessionCookie.Value == "" {
		util.ErrorJSON(w, http.StatusUnauthorized, "No valid session token found")
		return
	}

	sessionToken := appSessionCookie.Value

	util.UserCache.Delete(sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:     "AppSessionToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, 
		MaxAge:   -1,    
	})

	util.JSON(w, http.StatusOK, util.H{"message": "Successfully logged out"})
}
