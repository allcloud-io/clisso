package log

import (
	"io"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/go-homedir"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func NewLogger(logLevelFlag, logFilePath string, enableLogFile bool) *logrus.Logger {
	if Log != nil {
		return Log
	}

	// parse log level flag and set log level
	logLevel, err := logrus.ParseLevel(logLevelFlag)
	if err != nil {
		Log.Fatalf("Error parsing log level: %v", err)
	}
	Log = logrus.New()
	Log.SetLevel(logLevel)

	if enableLogFile {
		logFile, err := homedir.Expand(logFilePath)
		if err != nil {
			Log.Fatalf("Error expanding homedir: %v", err)
		}

		pathMap := lfshook.PathMap{
			logLevel: logFile,
		}
		Log.Hooks.Add(lfshook.NewHook(
			pathMap,
			&logrus.JSONFormatter{},
		))
		Log.Out = io.Discard
	} else {
		if runtime.GOOS == "windows" {
			// Handle terminal colors on Windows machines.
			// TODO, check if still required with the switch to logrus
			Log.SetOutput(colorable.NewColorableStdout())
		}
		Log.SetFormatter(&logrus.TextFormatter{PadLevelText: true})
	}
	Log.Infof("Log level set to %s", logLevelFlag)
	return Log
}

