package main

import (
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

func init() {
	logLevel, isSet := os.LookupEnv("LOG_LEVEL")
	if !isSet {
		logLevel = "WARN"
	}

	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "publish",
		Level: hclog.LevelFromString(logLevel),
	})

	// Set up standard logging
	log.SetOutput(appLogger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}
