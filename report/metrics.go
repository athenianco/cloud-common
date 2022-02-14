package report

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelUserID   = "user_id"
	labelUserName = "user_name"
	labelSeverity = "severity"

	severityInfo    = "info"
	severityMessage = "message"
	severityError   = "error"
)

var (
	countReportLogs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "count_report_logs",
		Help: "The count of log entries",
	}, []string{labelUserID, labelUserName, labelSeverity})
)

func countReportLogsInc(ctx context.Context, severity string) {
	countReportLogs.WithLabelValues(GetUserID(ctx), GetUserName(ctx), severity)
}
