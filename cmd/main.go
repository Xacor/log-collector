package main

import (
	"log"

	"net/http"

	"github.com/Xacor/log-collector/internal/config"
	api "github.com/Xacor/log-collector/internal/http"
	"github.com/gorilla/mux"
)

func main() {
	err := config.Load("./config")
	if err != nil {
		log.Fatal(err)
	}

	api, err := api.New()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/log/", api.Add).Methods("POST")

	log.Println("start serving :5050")
	err = http.ListenAndServe(":5050", r) //nolint:gosec//no need of timout here
	if err != nil {
		log.Fatal(err)
	}
}
