package server

import (
	"log"
	"net/http"

	"github.com/aanufriev/forum/configs"
	"github.com/gorilla/mux"
)

func StartApiServer() {
	mux := mux.NewRouter()
	log.Fatal(http.ListenAndServe(configs.ApiPort, mux))
}
