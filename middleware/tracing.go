package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	oplog "github.com/opentracing/opentracing-go/log"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Create our own logResponseWriter to wrap a standard http.ResponseWriter
// so we can store the status code.
type logResponseWriter struct {
	status int
	http.ResponseWriter
}

func NewlogResponseWriter(res http.ResponseWriter) *logResponseWriter {
	// Default the status code to 200
	return &logResponseWriter{200, res}
}

// Give a way to get the status
func (w logResponseWriter) Status() int {
	return w.status
}

// Satisfy the http.ResponseWriter interface
func (w logResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w logResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (rw *logResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (w logResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.status = statusCode

	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}

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

func opName(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if tpl, err := route.GetPathTemplate(); err == nil {
			return r.Proto + " " + r.Method + " " + tpl
		}
	}
	return r.Proto + " " + r.Method + " " + r.URL.Path
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		opName := opNameFunc(r)
		//Add data to context
		sp := opentracing.StartSpan(opName) // Start a new root span.
		defer sp.Finish()
		ctx := opentracing.ContextWithSpan(r.Context(), sp)
		res := NewlogResponseWriter(rw)
		next.ServeHTTP(res, r.WithContext(ctx))
		duration := time.Since(start)
		sp.LogFields(
			oplog.String("StartTime", start.Format(l.dateFormat)),
			oplog.Int("Status", res.Status()),
			oplog.Int64("Duration", int64(duration/time.Microsecond)),
			oplog.String("Hostname", r.Host),
			oplog.String("Method", r.Method),
			oplog.String("URL", r.URL.Path))

		l.Info(r.URL.Path,
			zap.String("StartTime", start.Format(l.dateFormat)),
			zap.Int("Status", res.Status()),
			zap.Duration("Duration", duration),
			zap.String("Hostname", r.Host),
			zap.String("Method", r.Method),
			zap.String("URL", r.URL.Path),
		)
	})
}

func opNameFunc(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if tpl, err := route.GetPathTemplate(); err == nil {
			return r.Proto + " " + r.Method + " " + tpl
		}
	}
	return r.Proto + " " + r.Method
}
