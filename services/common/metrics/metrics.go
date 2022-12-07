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
package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MerakMetrics struct {
	ServiceName     string
	OpsTotalLatency *prometheus.HistogramVec
	OpsSuccess      *prometheus.CounterVec
	OpsFail         *prometheus.CounterVec
}

// Creates new metrics struct
func NewMetrics(reg *prometheus.Registry, serviceName string) *MerakMetrics {
	opsTotalLatency := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: serviceName + "_Operations_Latency",
		Help: "latency_total",
	}, []string{"operation"})
	opsSuccess := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "_Operations_Success",
		Help: "success",
	}, []string{"operation"})
	opsFail := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "_Operations_Failed",
		Help: "fail",
	}, []string{"operation"})

	m := MerakMetrics{
		ServiceName:     serviceName,
		OpsTotalLatency: opsTotalLatency,
		OpsSuccess:      opsSuccess,
		OpsFail:         opsFail,
	}
	reg.MustRegister(m.OpsTotalLatency)
	reg.MustRegister(m.OpsSuccess)
	reg.MustRegister(m.OpsFail)
	return &m
}

// For use with defer statement to get metrics
func GetMetrics(merakMetrics *MerakMetrics, err *error) func() {
	name := getCallerName(1) // Returns callers name
	start := time.Now()
	return func() {
		t := time.Since(start)
		if *err != nil {
			merakMetrics.OpsFail.With(prometheus.Labels{"operation": name}).Inc()
		} else {
			merakMetrics.OpsSuccess.With(prometheus.Labels{"operation": name}).Inc()
		}
		merakMetrics.OpsTotalLatency.With(prometheus.Labels{"operation": name}).Observe(float64(t.Milliseconds()))
	}
}

func getCallerName(skipFrames int) string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skipFrames+2, pc)
	if n < 1 {
		return "unknown_caller"
	}
	frame, _ := runtime.CallersFrames(pc).Next()
	return frame.Function
}
