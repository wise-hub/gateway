package routes

import (
	"encoding/json"
	"net/http"

	"fibank.bg/fis-gateway-ws/internal/configuration"
	"fibank.bg/fis-gateway-ws/internal/middleware_custom"
	"fibank.bg/fis-gateway-ws/internal/util"
	"github.com/go-chi/chi/v5"
)

const endpointsFile = "./allowed_endpoints.txt"

func SetupRoutes(r chi.Router, d *configuration.Dependencies) {
	err := configuration.LoadAllowedEndpoints(endpointsFile)
	if err != nil {
		panic(err)
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "resource not found", http.StatusNotFound)
	})

	r.Post("/admin/register-endpoints", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Pwd string `json:"pwd"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if body.Pwd != d.Cfg.LoadEndpointsPwd {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := configuration.LoadAllowedEndpoints(endpointsFile); err != nil {
			http.Error(w, "Failed to load endpoints", http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Allowed endpoints refreshed"}`))
		}
	})

	r.Get("/admin/test500", func(w http.ResponseWriter, r *http.Request) {
		panic("simulating a server error")
	})

	r.Get("/admin/test401", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

	// invalidate token for userId - when role changes or user is blocked in db - to call this method
	r.Post("/admin/invalidate-token", func(w http.ResponseWriter, r *http.Request) {
		// TODO: add something better
		if r.Header.Get("pwd") != "12345" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var body struct {
			UserID string `json:"user_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		util.UserCache.DeleteTokensByUserID(body.UserID)

		response := util.H{
			"message": "Token invalidated",
		}
		util.JSON(w, http.StatusOK, response)

	})

	// for dev purpose print the cache
	r.Get("admin/cache", func(w http.ResponseWriter, r *http.Request) {
		entries := util.UserCache.GetAllEntries()
		util.JSON(w, http.StatusOK, entries)
	})

	apiGroup := chi.NewRouter()

	apiGroup.Post("/public/login", http.HandlerFunc(middleware_custom.LoginAction))

	apiGroup.With(middleware_custom.AuthMiddleware).Group(func(api chi.Router) {
		api.Get("/public/accounts", func(w http.ResponseWriter, r *http.Request) {
			userData, ok := middleware_custom.GetUserDataFromContext(r)
			if !ok {
				util.ErrorJSON(w, http.StatusInternalServerError, "User data not found")
				return
			}

			response := util.H{
				"user":     userData.Username,
				"accounts": userData.Accounts,
			}

			util.JSON(w, http.StatusOK, response)
		})
	})

	// Protected routes
	protectedGroup := chi.NewRouter()
	protectedGroup.Use(middleware_custom.AuthMiddleware)
	setupProxyRoutes(protectedGroup, d, "protected")

	r.Mount("/api", apiGroup)
	r.Mount("/api/v1", protectedGroup)
}
