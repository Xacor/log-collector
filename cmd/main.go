package main

import (
	"context"
	"log"

	"github.com/Xacor/log-collector/internal/config"
	"github.com/Xacor/log-collector/pkg/yandex"
	"github.com/spf13/viper"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	request := &logging.WriteRequest{
		Destination: &logging.Destination{
			Destination: &logging.Destination_LogGroupId{
				LogGroupId: viper.GetString("log_group_id"),
			},
		},
		Entries: []*logging.IncomingLogEntry{
			{
				Timestamp: timestamppb.Now(),
				Level:     logging.LogLevel_DEBUG,
				Message:   "La lal lala",
			},
		},
	}

	p, err := sdk.LogIngestion().LogIngestion().Write(ctx, request)

	if err != nil {
		log.Fatal(err)
	}
	log.Println(p)
}
