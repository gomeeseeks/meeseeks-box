package metrics

import (
	"net/http"
	"time"

	"github.com/gomeeseeks/meeseeks-box/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var namespace = "meeseeks"

// ReceivedCommandsCount is the count of commands that have been received
var ReceivedCommandsCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "received_commands_count",
	Help:      "Commands that have been received and are known",
})

// AliasedCommandsCount is the count of aliased commands that have been received
var AliasedCommandsCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "aliased_commands_count",
	Help:      "Commands that have been received and are an alias for another command",
})

// UnknownCommandsCount is the count of commands that have been received but are unknown
var UnknownCommandsCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "unknown_commands_count",
	Help:      "Commands that have been received but are unknown",
})

// RejectedCommandsCount is the count of commands that have been rejected
var RejectedCommandsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "rejected_commands_count",
	Help:      "Commands that have been rejected due to an auth failure",
}, []string{"command"})

// AcceptedCommandsCount is the count of commands that have been accepted
var AcceptedCommandsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "accepted_commands_count",
	Help:      "Commands that have been accepted",
}, []string{"command"})

// TaskDurations provides buckets to observe task execution latencies
var TaskDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Name:      "tasks_durations_seconds",
	Buckets:   prometheus.ExponentialBuckets(0.00025, 2, 18), // exponential buckets, starting at 0.25ms up to over 1h,
	Help:      "Command execution time distributions in seconds.",
}, []string{"command", "status"})

// LogLinesCount is the count of tasks that have been accepted
var LogLinesCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "log_lines_count",
	Help:      "Count of lines that have been written to the log",
})

var bootTime = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "boot_time_seconds",
	Help:      "unix timestamp of when the meeseeks process was started",
})

var buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: namespace,
	Name:      "build_info",
	Help:      "Version of the meeseeks executable",
}, []string{"name", "version", "date", "revision"})

func init() {
	bootTime.Set(float64(time.Now().Unix()))
	buildInfo.WithLabelValues(version.Name, version.Version, version.Date, version.Commit).Set(1)

	prometheus.MustRegister(buildInfo)
	prometheus.MustRegister(bootTime)
}

// RegisterServerMetrics registers all the metrics thar belong to a meeseeks server
func RegisterServerMetrics() {
	prometheus.MustRegister(ReceivedCommandsCount)
	prometheus.MustRegister(AliasedCommandsCount)
	prometheus.MustRegister(UnknownCommandsCount)
	prometheus.MustRegister(RejectedCommandsCount)
	prometheus.MustRegister(AcceptedCommandsCount)
	prometheus.MustRegister(TaskDurations)
	prometheus.MustRegister(LogLinesCount)
}

// RegisterPath registers prometheus metrics path
func RegisterPath(metricsPath string) {
	http.Handle(metricsPath, promhttp.Handler())
}
