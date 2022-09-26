package prometheus

import (
	"encoding/json"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/athenianco/cloud-common/report"
)

func init() {
	addr := os.Getenv("PUSHGATEWAY_ENDPOINT")
	job := os.Getenv("PROMETHEUS_JOB")
	if addr == "" || job == "" {
		return
	}
	var labels map[string]string
	if s := os.Getenv("PROMETHEUS_LABELS"); s != "" {
		if err := json.Unmarshal([]byte(s), &labels); err != nil {
			panic(err)
		}
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	const keyInstanceID = "instance"
	if _, ok := labels[keyInstanceID]; !ok {
		labels[keyInstanceID] = report.InstanceID()
	}
	p := push.New(addr, job).Gatherer(&staticLabels{
		g:      prometheus.DefaultGatherer,
		labels: labels,
	}).Format(expfmt.FmtText)
	report.RegisterFlusher(p.AddContext)
}

// staticLabels is a Gatherer that attaches a static set of labels to all metrics.
type staticLabels struct {
	g      prometheus.Gatherer
	labels map[string]string
}

// Gather implements prometheus.Gatherer.
func (s *staticLabels) Gather() ([]*io_prometheus_client.MetricFamily, error) {
	list, err := s.g.Gather()
	if len(s.labels) == 0 {
		return list, err
	}
	for _, f := range list {
		if f == nil {
			continue
		}
		for _, m := range f.Metric {
			if m == nil {
				continue
			}
			for k, v := range s.labels {
				k, v := k, v
				m.Label = append(m.Label, &io_prometheus_client.LabelPair{
					Name: &k, Value: &v,
				})
			}
		}
	}
	return list, err
}
