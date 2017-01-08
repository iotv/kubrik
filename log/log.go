package log

import "github.com/Sirupsen/logrus"

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.Level = logrus.DebugLevel
}