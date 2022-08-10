/*
MIT License
Copyright(c) 2022 Futurewei Cloud
    Permission is hereby granted,
    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
    to whom the Software is furnished to do so, subject to the following conditions:
    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

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
