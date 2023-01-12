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

// Merak's wrapper for the zap logger

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"strconv"
	"strings"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/pkg/errors"
	"github.com/tchap/zapext/zapsyslog"
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

// Where to log
type Location int

const (
	Syslog Location = iota + 1
	File
	Stdout
	Default
)

type options struct {
	location    Location
	level       zap.AtomicLevel
	serviceName string
	fileName    string
}

type merakLogError struct {
	Err     error
	Message string
}

func (e merakLogError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func newCore(opts *options) (any, error) {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "time"
	switch loc := opts.location; loc {
	case Syslog:
		encoder := zapcore.NewJSONEncoder(config)
		flagTag := flag.String("app", opts.serviceName, "syslog tag")
		writer, err := syslog.New(syslog.LOG_ERR|syslog.LOG_LOCAL0, *flagTag)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set up syslog")
		}
		core := zapsyslog.NewCore(opts.level, encoder, writer)
		return core, nil
	case File:
		encoder := zapcore.NewJSONEncoder(config)
		f, err := os.Create(opts.fileName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to write to file %s", opts.fileName))
		}
		core := zapcore.NewCore(encoder, zapcore.AddSync(f), opts.level)
		return core, nil
	case Stdout:
		encoder := zapcore.NewConsoleEncoder(config)
		return zapcore.NewCore(encoder, os.Stdout, opts.level), nil
	case Default:
		f, err := os.Create(opts.fileName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to write to file %s", opts.fileName))
		}
		consoleEncoder := zapcore.NewConsoleEncoder(config)
		fileEncoder := zapcore.NewJSONEncoder(config)
		return zapcore.NewTee(
			zapcore.NewCore(fileEncoder, zapcore.AddSync(f), opts.level),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), opts.level),
		), nil
	}
	return nil, merakLogError{errors.New("invalid logger case"), strconv.Itoa(int(opts.location))}
}

// Creates a new logger that writes to stdout with the given log level
func NewConsoleLogger(level Level) (*MerakLog, error) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(level))
	c, err := newCore(&options{
		level:    atomicLevel,
		location: Stdout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create core for stdout")
	}
	core, ok := c.(zapcore.Core)
	if !ok {
		return nil, merakLogError{errors.New("invalid zap core"), ""}
	}
	zap_logger := zap.New(
		core,
		zap.Development(),
		zap.AddStacktrace(zapcore.ErrorLevel))
	sugar := zap_logger.Sugar()
	return &MerakLog{sugar, atomicLevel}, nil
}

// Creates a new logger that writes to stdout and /var/log/merak/serviceName
func NewLogger(level Level, serviceName string) (*MerakLog, error) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(level))
	err := os.MkdirAll(constants.LOG_LOCATION, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "create path "+constants.LOG_LOCATION)
	}
	c, err := newCore(&options{
		level:    atomicLevel,
		location: Default,
		fileName: constants.LOG_LOCATION + serviceName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create core for stdout")
	}
	core, ok := c.(zapcore.Core)
	if !ok {
		return nil, merakLogError{errors.New("invalid zap core"), ""}
	}
	zap_logger := zap.New(
		core,
		zap.Development(),
		zap.AddStacktrace(zapcore.ErrorLevel))
	sugar := zap_logger.Sugar()
	return &MerakLog{sugar, atomicLevel}, nil
}

// Creates a new logger that writes to syslog with the given log level
func NewSysLogger(level Level, serviceName string) (*MerakLog, error) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(level))
	c, err := newCore(&options{
		level:       atomicLevel,
		serviceName: serviceName,
		location:    Syslog,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create core for syslog")
	}
	core, ok := c.(zapcore.Core)
	if !ok {
		return nil, merakLogError{errors.New("invalid zap core"), ""}
	}
	zap_logger := zap.New(
		core,
		zap.Development(),
		zap.AddStacktrace(zapcore.ErrorLevel))
	sugar := zap_logger.Sugar()
	return &MerakLog{sugar, atomicLevel}, nil
}

// Creates a new logger that writes to the given filepath with the given log level
func NewFileLogger(level Level, filepath string) (*MerakLog, error) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(level))
	c, err := newCore(&options{
		level:    atomicLevel,
		location: File,
		fileName: filepath,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create core for syslog")
	}
	core, ok := c.(zapcore.Core)
	if !ok {
		return nil, merakLogError{errors.New("invalid zap core"), ""}
	}
	zap_logger := zap.New(
		core,
		zap.Development(),
		zap.AddStacktrace(zapcore.ErrorLevel))
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
	macError := errors.New("sync /dev/stdout: bad file descriptor")
	linuxError := errors.New("sync /dev/stdout: invalid argument")
	e := log.Zap.Sync()
	if e != nil {
		// Will get error when flushing stdout logs. Ok to ignore EINVAL https://github.com/uber-go/zap/issues/328
		if e.Error() == macError.Error() || e.Error() == linuxError.Error() {
			return nil
		}
	}
	return e
}

func LevelEnvParser(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	case "panic":
		return PANIC, nil
	default:
		return -1, errors.New("invalid log level")
	}
}
