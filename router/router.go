package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"strings"
	"time"
)

type APIRouter struct {
	http.Server
	router *mux.Router
}

func handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
	w.Write([]byte("Gorilla!\n"))
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("Oh no !")
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	w.Write([]byte(`{"alive": true}`))
}

func PrintUsage(r *APIRouter) {
	err := r.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
}

func NewHttpAPI(listen string) *APIRouter {
	statMiddleware := stats.New()
	r := mux.NewRouter()
	api := mux.NewRouter()
	api.HandleFunc("/", handler)
	api.HandleFunc("/articles/{category}", handler).Methods("GET")
	api.HandleFunc("/panic", panicHandler)
	api.HandleFunc("/health", HealthCheckHandler)
	api.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stats := statMiddleware.Data()

		b, _ := json.Marshal(stats)

		w.Write(b)
	})
	// Set up classic Negroni Middleware
	recovery := negroni.NewRecovery()
	recovery.Formatter = &negroni.HTMLPanicFormatter{}
	recovery.PrintStack = true
	logger := negroni.NewLogger()
	r.PathPrefix("/").Handler(negroni.New(
		recovery,
		logger,
		statMiddleware,
		negroni.Wrap(api),
	))

	srv := &APIRouter{http.Server{
		Addr: listen,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}, r}
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	return srv
}
