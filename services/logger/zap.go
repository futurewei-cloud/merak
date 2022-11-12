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
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*MerakLog)(nil)

type MerakLog struct {
	Zap   *zap.SugaredLogger
	Level zap.AtomicLevel
}

type Level int8

// The supported Log Levels
const (
	DEBUG Level = Level(zap.DebugLevel)
	INFO  Level = Level(zap.InfoLevel)
	WARN  Level = Level(zap.WarnLevel)
	ERROR Level = Level(zap.ErrorLevel)
	PANIC Level = Level(zap.PanicLevel)
	FATAL Level = Level(zap.FatalLevel)
)

// Creates a new logger with the given log level
func NewLogger(level Level) (*MerakLog, error) {
	atomicLevel := zap.NewAtomicLevel()
	config := zap.NewProductionEncoderConfig()

	atomicLevel.SetLevel(zapcore.Level(level))
	zap_logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	))
	sugar := zap_logger.Sugar()
	return &MerakLog{sugar, atomicLevel}, nil
}

// Changes the log level to the given level
func (log *MerakLog) SetLevel(level Level) {
	log.Level.SetLevel(zapcore.Level(level))
}

// Returns the current log level
func (log *MerakLog) GetLevel() Level {
	return Level(log.Level.Level())
}

// Writes an info log
func (log *MerakLog) Info(msg string, kv ...any) {
	log.Zap.Infow(msg, kv...)
}

// Writes an error log
func (log *MerakLog) Error(msg string, kv ...any) {
	log.Zap.Errorw(msg, kv...)
}

// Writes a warning log
func (log *MerakLog) Warn(msg string, kv ...any) {
	log.Zap.Warnw(msg, kv...)
}

// Writes a debug log
func (log *MerakLog) Debug(msg string, kv ...any) {
	log.Zap.Debugw(msg, kv...)
}

// Writes a fatal log
func (log *MerakLog) Fatal(msg string, kv ...any) {
	log.Zap.Fatalw(msg, kv...)
}

// Writes a panic log
func (log *MerakLog) Panic(msg string, kv ...any) {
	log.Zap.Panicw(msg, kv...)
}

// Flushes the logs
func (log *MerakLog) Flush() error {
	return log.Zap.Sync()
}
