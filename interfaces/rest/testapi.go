package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	oplog "github.com/opentracing/opentracing-go/log"
)

func handler(c *AppContext, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var parentCtx opentracing.SpanContext
	parentSpan := opentracing.SpanFromContext(r.Context())
	if parentSpan != nil {
		parentCtx = parentSpan.Context()
	}

	sp := opentracing.StartSpan("handler", opentracing.ChildOf(parentCtx)) // Start a new root span.
	defer sp.Finish()
	ctx = opentracing.ContextWithSpan(ctx, sp)

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	sp.LogFields(
		oplog.String("event", "soft error"),
		oplog.String("type", "cache timeout"),
		oplog.Int("waited.millis", 1500))
	csp := opentracing.StartSpan("Event 1", opentracing.ChildOf(sp.Context()))
	defer csp.Finish()
	csp.LogFields(oplog.String("test", "test"))
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
	w.Write([]byte("Gorilla!\n"))
}

func panicHandler(c *AppContext, w http.ResponseWriter, r *http.Request) {
	sp := opentracing.StartSpan("GET /panic") // Start a new root span.
	defer sp.Finish()
	panic("Oh no !")
}

func NewAPI(context *AppContext) {

	declareNewRoute(context, "Article", []string{"GET"}, "/articles/{category}", "/api", handler)
	declareNewRoute(context, "Panic", []string{}, "/panic", "/api", panicHandler)
	declareNewRoute(context, "Standard", []string{}, "/", "/api", handler)

}

func declareNewRoute(context *AppContext, name string, method []string, pattern string, parent string, handler func(c *AppContext, w http.ResponseWriter, r *http.Request)) {

	contextedHandler := &ContextedHandler{
		AppContext:           context,
		ContextedHandlerFunc: handler,
	}

	route := Route{
		Name:             name,
		Method:           method,
		Pattern:          pattern,
		ParentRoute:      parent,
		ContextedHandler: contextedHandler, // We defined HelloWorldHandler in Part1
	}

	AddRoute(route)
}
