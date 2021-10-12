package internal

import (
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()
var log = Logger

func init() {
	log.SetLevel(logrus.DebugLevel)

	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}

	log.SetFormatter(formatter)
}
