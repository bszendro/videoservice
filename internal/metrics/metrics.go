package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	VidUploaded prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		VidUploaded: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "videoservice_example_ticks_total",
			Help: "The total number of uploaded videos",
		}),
	}
}

func NewMetricsServer(m *Metrics) *http.Server {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	reg.MustRegister(m.VidUploaded)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return &http.Server{Addr: ":7000", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
}
