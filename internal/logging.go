package internal

import (
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()
var log = Logger

func init() {
	log.SetLevel(logrus.DebugLevel)
}
