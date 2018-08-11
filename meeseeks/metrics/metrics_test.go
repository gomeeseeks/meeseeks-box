package metrics_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/mocks"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisteringServerMetrics(t *testing.T) {
	metrics.RegisterServerMetrics()

	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.AcceptedCommandsCount))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.AliasedCommandsCount))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.LogLinesCount))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.ReceivedCommandsCount))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.RejectedCommandsCount))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.TaskDurations))
	mocks.AssertEquals(t, true, prometheus.Unregister(metrics.UnknownCommandsCount))
}
