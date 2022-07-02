package logger

import (
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

var Log *Logger

type Logger struct {
	*logrus.Logger
}

type DefaultFieldHook struct {
	serviceName string
}

func (h *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *DefaultFieldHook) Fire(e *logrus.Entry) error {
	e.Data["Service"] = h.serviceName
	return nil
}

func NewLogger() *Logger {
	var baseLogger = logrus.New()
	var logger = &Logger{baseLogger}
	logger.Formatter = &logrus.JSONFormatter{}
	logger.SetOutput(os.Stdout)
	return logger
}

func StartLogger(service_name string, useSyslog bool, logLevel string) error {
	Log = NewLogger()
	Log.AddHook(&DefaultFieldHook{
		serviceName: service_name,
	})

	if useSyslog {
		Log.Out = ioutil.Discard
		sysLevel := syslogLevel(logLevel)
		hook, err := logrus_syslog.NewSyslogHook("", "", sysLevel, "")
		if err != nil {
			return err
		}
		Log.AddHook(hook)
	} else {
		Log.Out = os.Stdout
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			lvl = logrus.InfoLevel
		}
		Log.Level = logrus.Level(lvl)
	}
	return nil
}

func syslogLevel(level string) syslog.Priority {
	switch level {
	case "debug":
		return syslog.LOG_DEBUG
	case "warning", "warn":
		return syslog.LOG_WARNING
	case "error":
		return syslog.LOG_ERR
	case "fatal":
		return syslog.LOG_CRIT
	default:
		return syslog.LOG_INFO
	}
}
