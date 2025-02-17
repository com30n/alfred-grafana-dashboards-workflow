package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

func InitLogger() *logrus.Logger {
	log := logrus.New()
	logFileName := os.Getenv("LOG_FILE")
	if logFileName == "" {
		logFileName = "dashboards.log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Out = logFile
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		log.SetLevel(logrus.DebugLevel)
	case "INFO":
		log.SetLevel(logrus.InfoLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	return log
}
