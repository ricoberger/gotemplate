package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reqMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gotemplate",
		Name:      "chi_requests_total",
		Help:      "Number of HTTP requests processed, partitioned by status code, method and path.",
	}, []string{"response_code", "request_method", "request_path"})

	sumMetric = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "gotemplate",
		Name:       "chi_request_duration_milliseconds",
		Help:       "Latency of HTTP requests processed, partitioned by status code, method and path.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
	}, []string{"response_code", "request_method", "request_path"})
)

// Metrics is a middleware that handles the Prometheus metrics for our API and chi.
func Metrics(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrw, r)

		// In go-chi/chi, full route pattern could only be extracted once the request is executed
		// See: https://github.com/go-chi/chi/issues/150#issuecomment-278850733
		routeStr := strings.Join(chi.RouteContext(r.Context()).RoutePatterns, "")

		reqMetric.WithLabelValues(strconv.Itoa(wrw.Status()), r.Method, routeStr).Inc()
		sumMetric.WithLabelValues(strconv.Itoa(wrw.Status()), r.Method, routeStr).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
	}

	return http.HandlerFunc(fn)
}
