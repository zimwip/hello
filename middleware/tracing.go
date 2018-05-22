package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

// LoggerEntry is the structure
// passed to the template.
type LoggerEntry struct {
	StartTime string
	Status    int
	Duration  time.Duration
	Hostname  string
	Method    string
	Path      string
	Request   *http.Request
}

// LoggerDefaultDateFormat is the
// format used for date by the
// default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// ALogger implements just enough log.Logger interface to be compatible with other implementations
	*zap.Logger
	dateFormat string
}

// NewLogger returns a new Logger instance
func NewLogger(log *zap.Logger) *Logger {
	logger := &Logger{Logger: log, dateFormat: LoggerDefaultDateFormat}
	return logger
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	//Add data to context
	ctx := context.WithValue(r.Context(), "span", "toto")
	next(rw, r.WithContext(ctx))

	res := rw.(negroni.ResponseWriter)
	l.Info(r.URL.Path,
		zap.String("StartTime", start.Format(l.dateFormat)),
		zap.Int("Status", res.Status()),
		zap.Duration("Duration", time.Since(start)),
		zap.String("Hostname", r.Host),
		zap.String("Method", r.Method),
		zap.String("URL", r.URL.Path),
	)
}
