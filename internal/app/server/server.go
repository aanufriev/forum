package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/aanufriev/forum/configs"
	"github.com/aanufriev/forum/internal/pkg/middleware"
	userDelivery "github.com/aanufriev/forum/internal/pkg/user/delivery"
	userRepository "github.com/aanufriev/forum/internal/pkg/user/repository"
	userUsecase "github.com/aanufriev/forum/internal/pkg/user/usecase"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func StartApiServer() {
	db, err := sql.Open(configs.Postgres, configs.DataSourceNamePostgres)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return
	}

	userRepository := userRepository.New(db)
	userUsecase := userUsecase.New(userRepository)
	userDelivery := userDelivery.New(userUsecase)

	mux := mux.NewRouter().PathPrefix(configs.ApiUrl).Subrouter()
	mux.HandleFunc("/user/{nickname}/create", userDelivery.Create).Methods("POST")
	mux.HandleFunc("/user/{nickname}/profile", userDelivery.Get).Methods("GET")
	mux.HandleFunc("/user/{nickname}/profile", userDelivery.Update).Methods("POST")

	mixWithAccessLog := middleware.AccessLog(mux)
	muxWithCORS := middleware.CORS(mixWithAccessLog)

	log.Printf("server started at port %v", configs.ApiPort)
	log.Fatal(http.ListenAndServe(configs.ApiPort, muxWithCORS))
}
