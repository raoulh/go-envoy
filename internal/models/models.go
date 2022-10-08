package models

import (
	"sync"

	logger "github.com/raoulh/go-envoy/internal/log"

	"github.com/sirupsen/logrus"
)

var (
	logging *logrus.Entry

	quitRefresh  chan interface{}
	wgDone       sync.WaitGroup
	runningTasks sync.Map
)

func init() {
	logging = logger.NewLogger("models")
}

// Init models
func Init() (err error) {
	return
}

// Shutdown models
func Shutdown() {
}
