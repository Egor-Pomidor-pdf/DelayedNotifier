package handler

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wb-go/wbf/ginext"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}


func MetricsMiddleware(c *ginext.Context) {
	start := time.Now()
	c.Next() // выполняем хендлер

	status := c.Writer.Status()
	path := c.FullPath() // или c.Request.URL.Path
	method := c.Request.Method
	labels := prometheus.Labels{
		"method": method,
		"path":   path,
		"status": strconv.Itoa(status),
	}

	httpRequestsTotal.With(labels).Inc()
	httpRequestDuration.With(labels).Observe(time.Since(start).Seconds())
}
