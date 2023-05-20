package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Xacor/log-collector/internal/storage"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const (
	storeTreshold = 50
)

type LogHandler struct {
	Store  *storage.LogStore
	SDK    *ycsdk.SDK
	Ticker *time.Ticker
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
		ctx := context.TODO()
		defer ctx.Done()

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

func (api *LogHandler) OnTimeout() {
	for range api.Ticker.C {
		if api.Store.Length() == 0 {
			continue
		}
		ctx := context.TODO()
		_, err := api.FlushLogs(ctx)
		if err != nil {
			log.Println(err)
		}
		ctx.Done()
	}
}
