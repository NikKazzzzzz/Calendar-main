package prometheus

import (
	"github.com/NikKazzzzzz/Calendar-main/monitoring"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		monitoring.RequestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()
		monitoring.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		if rw.Status() >= 400 {
			monitoring.RequestErrors.WithLabelValues(r.Method, r.URL.Path, http.StatusText(rw.Status())).Inc()
		}
	})
}
