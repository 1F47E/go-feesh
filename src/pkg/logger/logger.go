package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	Log.Out = os.Stdout

	Log.Level = logrus.DebugLevel

	// The TextFormatter is default, you don't actually have to do this.
	Log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
		// DisableColors: true,
		ForceQuote:    true,
		FullTimestamp: true,
		// DisableLevelTruncation: false,
	}
}
