package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Generates log of random level at every given interval
func DummyLogging(interval time.Duration) {
	t := time.NewTicker(interval)
	for range t.C {
		switch rand.Intn(3) {
		case 0:
			log.Info("An info log.")
		case 1:
			log.Warn("A warning log!")
		case 2:
			log.Error("An error log!")
		}
	}

}

func main() {

	DummyLogging(time.Second)
}

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
}
