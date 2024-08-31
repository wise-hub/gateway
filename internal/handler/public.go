package handler

import (
	"net/http"

	"fibank.bg/fis-gateway-ws/internal/util"
)

func PublicHandler(w http.ResponseWriter, r *http.Request) {
	response := util.H{"message": "This is open to the public"}
	util.JSON(w, http.StatusOK, response)
}
