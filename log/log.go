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
		Log.Tracef("Logger already initialized")
		return Log
	}

	// parse log level flag and set log level
	logLevel, err := logrus.ParseLevel(logLevelFlag)
	if err != nil {
		logrus.Fatalf("Error parsing log level: %v", err)
	}
	Log = logrus.New()
	Log.SetLevel(logLevel)

	if enableLogFile {
		logFile, err := homedir.Expand(logFilePath)
		if err != nil {
			Log.Fatalf("Error expanding homedir: %v", err)
		}

		// set all log levels to write to the log file
		pathMap := lfshook.PathMap{
			logrus.TraceLevel: logFile,
			logrus.DebugLevel: logFile,
			logrus.InfoLevel:  logFile,
			logrus.WarnLevel:  logFile,
			logrus.ErrorLevel: logFile,
			logrus.FatalLevel: logFile,
			logrus.PanicLevel: logFile,
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
	Log.Warnf("Log level set to %s", logLevelFlag)
	return Log
}
