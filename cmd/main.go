package main

import (
	"log"

	"net/http"

	"github.com/Xacor/log-collector/internal/config"
	api "github.com/Xacor/log-collector/internal/http"
	"github.com/Xacor/log-collector/pkg/yandex"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	err := config.Load("./config")
	if err != nil {
		log.Fatal(err)
	}

	api, err := api.New(&api.HandlerConfig{
		LogGroupID: viper.GetString("log_group_id"),
		IAMconf: yandex.Config{
			ServiceAccountID: viper.GetString("service_account_id"),
			KeyFile:          viper.GetString("key_file"),
			KeyID:            viper.GetString("key_id"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/log/", api.Add).Methods("POST")

	addr := viper.GetString("address")
	log.Println("start serving on ", addr)
	err = http.ListenAndServe(addr, r) //nolint:gosec//no need of timout here
	if err != nil {
		log.Fatal(err)
	}
}
