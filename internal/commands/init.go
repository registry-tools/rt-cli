package commands

import (
	"errors"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

var (
	ErrLoginRequired = errors.New("no credentials found. Set the credentials in the environment or use `rt login`")
)

func init() {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "ERROR"
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "rt",
		Level: hclog.LevelFromString(level),
	})

	log.SetOutput(appLogger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}
