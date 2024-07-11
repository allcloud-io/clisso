package log

import (
	"io"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/go-homedir"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var Log = logrus.New()
var logger = logrus.New()
var Hook *test.Hook

func SetupLogger(logLevelFlag, logFilePath string, enableLogFile, testLogger bool) (*logrus.Logger, *test.Hook) {

	if logger != nil {
		logger.Tracef("Logger already initialized")
		return Log, Hook
	}
	if testLogger {
		logger, Hook = test.NewNullLogger()
		logger.ExitFunc = func(int) {}
		return logger, Hook
	}

	// parse log level flag and set log level
	logLevel, err := logrus.ParseLevel(logLevelFlag)
	if err != nil {
		logrus.Fatalf("Error parsing log level: %v", err)
	}
	logger.SetLevel(logLevel)

	if enableLogFile {
		logFile, err := homedir.Expand(logFilePath)
		if err != nil {
			logger.Fatalf("Error expanding homedir: %v", err)
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
		logger.Hooks.Add(lfshook.NewHook(
			pathMap,
			&logrus.JSONFormatter{},
		))
		logger.Out = io.Discard
	} else {
		if runtime.GOOS == "windows" {
			// Handle terminal colors on Windows machines.
			// TODO, check if still required with the switch to logrus
			logger.SetOutput(colorable.NewColorableStdout())
		}
		logger.SetFormatter(&logrus.TextFormatter{PadLevelText: true})
	}
	logger.Warnf("Log level set to %s", logLevelFlag)
	return logger, nil
}

// Fatal is a wrapper for Logrus Fatal
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf is a wrapper for Logrus Fatalf
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Fatalln is a wrapper for Logrus Fatalln
func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

// Panic is a wrapper for Logrus Panic
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf is a wrapper for Logrus Panicf
func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

// Panicln is a wrapper for Logrus Panicln
func Panicln(args ...interface{}) {
	logger.Panicln(args...)
}

// Print is a wrapper for Logrus Print
func Print(args ...interface{}) {
	logger.Print(args...)
}

// Printf is a wrapper for Logrus Printf
func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

// Println is a wrapper for Logrus Println
func Println(args ...interface{}) {
	logger.Println(args...)
}

// Error is a wrapper for Logrus Error
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf is a wrapper for Logrus Errorf
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Errorln is a wrapper for Logrus Errorln
func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

// Warn is a wrapper for Logrus Warn
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf is a wrapper for Logrus Warnf
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Warnln is a wrapper for Logrus Warnln
func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

// Info is a wrapper for Logrus Info
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof is a wrapper for Logrus Infof
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Infoln is a wrapper for Logrus Infoln
func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

// Debug is a wrapper for Logrus Debug
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf is a wrapper for Logrus Debugf
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Debugln is a wrapper for Logrus Debugln
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

// Trace is a wrapper for Logrus Trace
func Trace(args ...interface{}) {
	logger.Trace(args...)
}

// Tracef is a wrapper for Logrus Tracef
func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

// Traceln is a wrapper for Logrus Traceln
func Traceln(args ...interface{}) {
	logger.Traceln(args...)
}

// WithFields is a wrapper for Logrus WithFields
func WithFields(fields Fields) *logrus.Entry {
	return logger.WithFields(logrus.Fields(fields))
}

// WithField is a wrapper for Logrus WithField
func WithField(key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

// WithError is a wrapper for Logrus WithError
func WithError(err error) *logrus.Entry {
	return logger.WithError(err)
}

// wrap log.Fields with a type so we can use it in the WithFields method
type Fields logrus.Fields

// GetLevel returns the current log level
func GetLevel() Level {
	return Level(logger.GetLevel())
}

// SetLevel sets the log level
func SetLevel(level Level) {
	logger.SetLevel(logrus.Level(level))
}

// Level is a wrapper for Logrus Level
type Level logrus.Level

// TraceLevel is a wrapper for Logrus TraceLevel
const TraceLevel = Level(logrus.TraceLevel)

// DebugLevel is a wrapper for Logrus DebugLevel
const DebugLevel = Level(logrus.DebugLevel)
