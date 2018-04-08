package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "meeseeks"

// ReceivedCommandsCount is the count of commands that have been received
var ReceivedCommandsCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "received_commands_count",
	Help:      "Commands that have been received and are known",
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

// FailedTasksCount is the count of tasks that have finished in failure
var FailedTasksCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "failed_tasks_count",
	Help:      "Tasks that have finished in failure",
}, []string{"command"})

// SuccessfulTasksCount is the count of tasks that have finished successfully
var SuccessfulTasksCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "successful_tasks_count",
	Help:      "Tasks that have finished successfully",
}, []string{"command"})

// TaskDurations provides buckets to observe task execution latencies
var TaskDurations = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: namespace,
	Name:      "tasks_durations_seconds",
	Buckets:   prometheus.ExponentialBuckets(0.00025, 2, 18), // exponential buckets, starting at 0.25ms up to over 1h,
	Help:      "Command execution time distributions in seconds.",
})

// LogLinesCount is the count of tasks that have been accepted
var LogLinesCount = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "log_lines_count",
	Help:      "Count of lines that have been written to the log",
})

func init() {
	prometheus.MustRegister(ReceivedCommandsCount)
	prometheus.MustRegister(UnknownCommandsCount)
	prometheus.MustRegister(RejectedCommandsCount)
	prometheus.MustRegister(AcceptedCommandsCount)
	prometheus.MustRegister(FailedTasksCount)
	prometheus.MustRegister(SuccessfulTasksCount)
	prometheus.MustRegister(TaskDurations)
	prometheus.MustRegister(LogLinesCount)
}
