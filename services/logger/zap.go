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

import "go.uber.org/zap"

var _ Logger = (*MerakLog)(nil)

type MerakLog struct {
	Zap *zap.SugaredLogger
}

func NewLogger() (*MerakLog, error) {
	zap_logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	sugar := zap_logger.Sugar()
	return &MerakLog{sugar}, nil
}

func (log *MerakLog) Infoln(msg string, args ...any) {
	log.Zap.Infoln(msg)
}

func (log *MerakLog) Errorln(msg string, args ...any) {
	log.Zap.Error(msg)
}

func (log *MerakLog) Warnln(msg string, args ...any) {
	log.Zap.Warn(msg)
}

func (log *MerakLog) Debugln(msg string, args ...any) {
	log.Zap.Debugln(msg)
}

func (log *MerakLog) Fatalln(msg string, args ...any) {
	log.Zap.Fatalln(msg, args)
}

func (log *MerakLog) Panicln(msg string, args ...any) {
	log.Zap.Panicln(msg, args)
}

func (log *MerakLog) Flush() error {
	return log.Zap.Sync()
}
