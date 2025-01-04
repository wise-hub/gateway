package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wise-hub/gateway/internal/configuration"
	"github.com/wise-hub/gateway/internal/filter"
	"github.com/wise-hub/gateway/internal/handler"
	"github.com/wise-hub/gateway/internal/util"
)

const endpointsFile = "./allowed_endpoints.txt"

func SetupRoutes(r chi.Router, d *configuration.Dependencies) {
	initializeEndpoints(r)
	registerAdminRoutes(r, d)
	registerPublicRoutes(r, d)
	registerProtectedRoutes(r, d)
	serveStaticFileRoutes(r)
}

func serveStaticFileRoutes(r chi.Router) {
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {

		http.SetCookie(w, &http.Cookie{
			Name:     "AppSessionToken",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false, 
			MaxAge:   -1, 
		})

		htmlContent, err := os.ReadFile("./internal/static/login.html")
		if err != nil {
			http.Error(w, "Unable to load login page", http.StatusInternalServerError)
			return
		}

		pageContent := strings.ReplaceAll(string(htmlContent), "{{CSRF_TOKEN}}", util.GenerateSoftAuthToken(r.UserAgent()))

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(pageContent))
	})

	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/static"))).ServeHTTP)
}


func initializeEndpoints(r chi.Router) {
	if err := configuration.LoadAllowedEndpoints(endpointsFile); err != nil {
		panic(err)
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		util.ErrorJSON(w, http.StatusNotFound, "resource not found")
	})
}

func registerAdminRoutes(r chi.Router, d *configuration.Dependencies) {
	r.Route("/admin", func(r chi.Router) {
		r.Post("/register-endpoints", handleRegisterEndpoints(d))
		r.Get("/test500", func(w http.ResponseWriter, r *http.Request) {
			panic("simulating a server error")
		})
		r.Get("/test401", func(w http.ResponseWriter, r *http.Request) {
			util.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		})
		r.Post("/invalidate-token", handleInvalidateToken())
		r.Get("/cache", func(w http.ResponseWriter, r *http.Request) {
			util.JSON(w, http.StatusOK, util.UserCache.GetAllEntries())
		})
	})
}

func registerPublicRoutes(r chi.Router, d *configuration.Dependencies) {
	r.Post("/api/v1/login", handler.LoginHandler(d.Db))
}

func registerProtectedRoutes(r chi.Router, d *configuration.Dependencies) {
	apiGroup := chi.NewRouter()
	apiGroup.Use(filter.AuthMiddleware)

	apiGroup.Post("/logout", http.HandlerFunc(handler.LogoutHandler))

	apiGroup.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		userData, ok := filter.GetUserDataFromContext(r)
		if !ok {
			util.ErrorJSON(w, http.StatusInternalServerError, "User data not found")
			return
		}
		util.JSON(w, http.StatusOK, util.H{
			"userData": struct {
				SoftAuthToken string   `json:"soft_auth_token"`
				Username      string   `json:"username"`
				Roles         []string `json:"roles"`
			}{
				SoftAuthToken: userData.SoftAuthToken,
				Username:      userData.Username,
				Roles:         userData.Roles,
			},
		})
	})

	setupProxyRoutes(apiGroup, d, "protected")

	r.Mount("/api/v1", apiGroup)
}

func handleRegisterEndpoints(d *configuration.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Pwd string `json:"pwd"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Pwd != d.Cfg.LoadEndpointsPwd {
			util.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized or Invalid Request")
			return
		}
		if err := configuration.LoadAllowedEndpoints(endpointsFile); err != nil {
			util.ErrorJSON(w, http.StatusInternalServerError, "Failed to load endpoints")
			return
		}
		util.JSON(w, http.StatusOK, util.H{"message": "Allowed endpoints refreshed"})
	}
}

func handleInvalidateToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("pwd") != "12345" {
			util.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		var body struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			util.ErrorJSON(w, http.StatusBadRequest, "Invalid request")
			return
		}
		util.UserCache.DeleteTokensByUserID(body.UserID)
		util.JSON(w, http.StatusOK, util.H{"message": "Token invalidated"})
	}
}
