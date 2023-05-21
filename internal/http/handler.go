package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Xacor/log-collector/internal/storage"
	"github.com/Xacor/log-collector/pkg/yandex"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const (
	storeTreshold = 50
	ctxTimeout    = 15
	flushTimeout  = 5
)

type LogHandler struct {
	Store  *storage.LogStore
	SDK    *ycsdk.SDK
	Ticker *time.Ticker
}

func New() (*LogHandler, error) {
	iam, err := yandex.NewIAM()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
	defer cancel()

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.NewIAMTokenCredentials(iam.Value()),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sdk instance")
	}

	handler := &LogHandler{
		Store:  storage.NewLogStore(),
		SDK:    sdk,
		Ticker: time.NewTicker(flushTimeout * time.Second),
	}

	go handler.FlushOnTimeout()
	return handler, nil
}

func (api *LogHandler) Add(w http.ResponseWriter, r *http.Request) {
	var in map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(in)

	if api.Store.Length() >= storeTreshold {
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
		defer cancel()

		_, err = api.FlushLogs(ctx)
		if err != nil {
			log.Println(err)
		}
	}

	_, err = api.Store.AddLog(in)
	if err != nil {
		log.Printf("unable to add log entry: %v", err)
	}
	w.WriteHeader(http.StatusOK)
}

func (api *LogHandler) FlushLogs(ctx context.Context) (*logging.WriteResponse, error) {
	request := &logging.WriteRequest{
		Destination: &logging.Destination{
			Destination: &logging.Destination_LogGroupId{
				LogGroupId: viper.GetString("log_group_id"),
			},
		},
		Entries: api.Store.GetLogs(),
	}
	p, err := api.SDK.LogIngestion().LogIngestion().Write(ctx, request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return p, nil
}

func (api *LogHandler) FlushOnTimeout() {
	for range api.Ticker.C {
		if api.Store.Length() == 0 {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout*time.Second)
		_, err := api.FlushLogs(ctx)
		if err != nil {
			log.Println(err)
		}
		cancel()
	}
}
