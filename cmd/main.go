package main

import (
	"context"
	"log"
	"time"

	"net/http"

	"github.com/Xacor/log-collector/internal/config"
	api "github.com/Xacor/log-collector/internal/http"
	"github.com/Xacor/log-collector/internal/storage"
	"github.com/Xacor/log-collector/pkg/yandex"
	"github.com/gorilla/mux"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const (
	flushTimeout = 5
)

func main() {
	err := config.Load("./config")
	if err != nil {
		log.Fatal(err)
	}

	iam, err := yandex.NewIAM()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.NewIAMTokenCredentials(iam.Value()),
	})
	if err != nil {
		log.Fatal(err)
	}
	api := &api.LogHandler{
		Store:  storage.NewLogStore(),
		SDK:    sdk,
		Ticker: time.NewTicker(time.Second * flushTimeout),
	}

	go api.OnTimeout()

	r := mux.NewRouter()
	r.HandleFunc("/log/", api.Add).Methods("POST")
	log.Println("start serving :5050")
	err = http.ListenAndServe(":5050", r) //nolint:gosec//no need of timout here
	if err != nil {
		log.Fatal(err)
	}
}
