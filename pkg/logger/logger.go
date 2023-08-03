package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

type LoggerEntry struct {
	logrus.Entry
}

func init() {
	Log.Out = os.Stdout

	Log.Level = logrus.DebugLevel

	// for production
	// log.SetFormatter(&log.JSONFormatter{})

	// The TextFormatter is default, you don't actually have to do this.
	Log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
		// DisableColors: true,
		ForceQuote:      true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		// DisableLevelTruncation: false,
	}
}
