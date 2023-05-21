package storage

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxStore = 100
)

type LogJSON map[string]interface{}

type LogStore struct {
	logs []*logging.IncomingLogEntry
	mu   sync.RWMutex
}

func NewLogStore() *LogStore {
	return &LogStore{
		mu:   sync.RWMutex{},
		logs: make([]*logging.IncomingLogEntry, 0, maxStore),
	}
}
func (ls *LogStore) Length() int {
	return len(ls.logs)
}

func (ls *LogStore) AddLog(in LogJSON) (LogJSON, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	jsonPayload, err := structpb.NewStruct(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct json payload")
	}
	level := logging.LogLevel_Level(logging.LogLevel_Level_value[strings.ToUpper(in["level"].(string))])

	timeStr, ok := in["time"].(string)
	if !ok {
		return nil, errors.New("unable to assert timestamp")
	}
	timestamp, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, errors.Wrap(err, "unamble to parse timestamp")
	}
	message, ok := in["msg"].(string)
	if !ok {
		message = ""
	}

	streamName, ok := in["stream_name"].(string)
	if !ok {
		streamName = ""
	}

	// construct proto here
	newLogEntry := &logging.IncomingLogEntry{
		Timestamp:   timestamppb.New(timestamp),
		Message:     message,
		Level:       level,
		JsonPayload: jsonPayload,
		StreamName:  streamName,
	}
	ls.logs = append(ls.logs, newLogEntry)

	log.Println(ls.logs)
	return in, nil
}

func (ls *LogStore) GetLogs() []*logging.IncomingLogEntry {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	result := make([]*logging.IncomingLogEntry, len(ls.logs))
	for i := 0; i < len(ls.logs); i++ {
		result = append(result, ls.logs[i])      //nolint:makezero//result memory preallocated to avoid perfomanse loses
		ls.logs[i] = &logging.IncomingLogEntry{} // erase values from store
	}
	ls.logs = ls.logs[:0] // set len to zero

	return result
}
