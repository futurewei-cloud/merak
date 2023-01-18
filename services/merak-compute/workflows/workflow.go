package workflow

import (
	"github.com/futurewei-cloud/merak/services/common/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	MerakMetrics       metrics.Metrics
	PrometheusRegistry *prometheus.Registry
)
