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
	"fmt"
	"os"
	"testing"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Used for validating log output
func createObservedLogger(level Level) (*MerakLog, *observer.ObservedLogs) {
	observedZapCore, observedLogs := observer.New(zapcore.Level(level))
	observedLogger := zap.New(observedZapCore, zap.OnFatal(zapcore.CheckWriteAction(zapcore.WriteThenPanic)))
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(level))
	logger := MerakLog{observedLogger.Sugar(), atomicLevel}
	return &logger, observedLogs
}

func TestNewConsoleLogger(t *testing.T) {
	logger, err := NewConsoleLogger(INFO)
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	assert.Nil(t, logger.Flush())
	logger.Info("hi", "1", "2")
}

func TestNewSysLogger(t *testing.T) {
	logger, err := NewSysLogger(INFO, "Wrapper Test")
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	logger.Info("hi", "1", "2")
	defer assert.Nil(t, logger.Flush())
}
func TestNewFileLogger(t *testing.T) {
	logger, err := NewFileLogger(INFO, "/tmp/zap_test")
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	logger.Info("hi", "1", "2")
	assert.Nil(t, logger.Flush())
	e := os.Remove("/tmp/zap_test")
	assert.Nil(t, e)
}

func TestFileAndConsoleLogger(t *testing.T) {
	logger, err := NewLogger(INFO, "test")
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	logger.Info("hi", "1", "2")
	assert.Nil(t, logger.Flush())
	e := os.Remove(constants.LOG_LOCATION + "test")
	assert.Nil(t, e)
}

func loggerLevelsTest(t *testing.T, level Level) {
	tests := []struct {
		message    string
		expMessage string
		fields     []string
		expFields  []string
	}{
		{
			message:    "Hello!",
			expMessage: "Hello!",
			fields:     []string{"field1", "1", "2", "val2"},
			expFields:  []string{"field1", "1", "2", "val2"},
		},
		{
			message:    "Goodbye!",
			expMessage: "Goodbye!",
			fields:     []string{"123", "1.0", "-1", "101112"},
			expFields:  []string{"123", "1.0", "-1", "101112"},
		},
		{
			message:    "Hello!",
			expMessage: "Hello!",
			fields:     []string{"field1", "val1", "field2", "val2"},
			expFields:  []string{"field1", "val1", "field2", "val2"},
		},
	}
	logger, observedLogs := createObservedLogger(level)
	for i, test := range tests {
		t.Run(fmt.Sprintf("Test Log Level %d", level), func(t *testing.T) {
			switch l := level; l {
			case INFO:
				logger.Info(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			case DEBUG:
				logger.Debug(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			case ERROR:
				logger.Error(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			case WARN:
				logger.Warn(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			case PANIC:
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Logger did not trigger panic")
					}
				}()
				logger.Panic(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			case FATAL:
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Logger did not trigger fatal")
					}
				}()
				logger.Fatal(test.message, test.fields[0], test.fields[1], test.fields[2], test.fields[3])
			}
			logs := observedLogs.All()[i]
			assert.Equal(t, i+1, observedLogs.Len())
			assert.Equal(t, logs.Level, zapcore.Level(level))
			assert.Equal(t, test.expMessage, logs.Message)
			assert.ElementsMatch(t, []zap.Field{
				zap.String(test.expFields[0], test.expFields[1]),
				zap.String(test.expFields[2], test.expFields[3]),
			}, logs.Context)
		})
	}
	defer assert.Nil(t, logger.Flush())
}

func TestLoggerDebug(t *testing.T) {
	loggerLevelsTest(t, DEBUG)
}

func TestLoggerInfo(t *testing.T) {
	loggerLevelsTest(t, INFO)
}

func TestLoggerWarn(t *testing.T) {
	loggerLevelsTest(t, WARN)
}

func TestLoggerError(t *testing.T) {
	loggerLevelsTest(t, ERROR)
}

func TestLoggerPanic(t *testing.T) {
	loggerLevelsTest(t, PANIC)
}
func TestLoggerFatal(t *testing.T) {
	loggerLevelsTest(t, FATAL)
}

func TestLoggerFlush(t *testing.T) {
	logger, _ := createObservedLogger(INFO)
	logger.Info("hi", "1", 2)
	defer assert.Nil(t, logger.Flush())
}
func TestLoggerSetLevel(t *testing.T) {
	logger, err := NewConsoleLogger(INFO)
	assert.Nil(t, err)
	logger.SetLevel(DEBUG)
	assert.Equal(t, DEBUG, logger.GetLevel())
}
func TestLoggerGetLevel(t *testing.T) {
	logger, err := NewConsoleLogger(INFO)
	assert.Nil(t, err)
	assert.Equal(t, INFO, logger.GetLevel())
}
