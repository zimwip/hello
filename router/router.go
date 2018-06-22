package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zimwip/hello/config"
	"github.com/zimwip/hello/domain"
	"github.com/zimwip/hello/infrastructure"
	"github.com/zimwip/hello/middleware"

	"github.com/gorilla/mux"
	"github.com/unrolled/secure"

	"go.uber.org/zap"
)

type APIRouter struct {
	http.Server
	router *mux.Router
	logger *zap.Logger
}

func PrintUsage(r *APIRouter) {
	err := r.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Printf("ROUTE: %s, %p\n", pathTemplate, route)
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

/*
Create a new server listen to API call and static file
return an APIRouter
*/
func NewServer(secured_port string, port string, staticDir string, logger *zap.Logger) *APIRouter {

	// master router
	appContext := domain.AppContext{}
	r := infrastructure.NewRouter(&appContext)
	// Set up classic Negroni Middleware
	recovery := middleware.NewRecovery()
	recovery.Formatter = &middleware.HTMLPanicFormatter{}
	recovery.PrintStack = true

	r.Use(middleware.NewLogger(logger).Middleware)
	r.Use(recovery.Middleware)

	secureMiddleware := secure.New(secure.Options{
		HostsProxyHeaders:    []string{"X-Forwarded-Host"},
		SSLRedirect:          true,
		SSLHost:              secured_port,
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		//		ContentSecurityPolicy: "script-src $NONCE",
		PublicKey:     `pin-sha256="base64+primary=="; pin-sha256="base64+backup=="; max-age=5184000; includeSubdomains; report-uri="https://www.example.com/hpkp-report"`,
		IsDevelopment: config.Config().IsDev(),
	})

	r.Use(secureMiddleware.Handler)

	// static route
	static := r.PathPrefix("/").Subrouter().StrictSlash(true)
	static.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	static.Use(middleware.NewStatic(http.Dir(staticDir)).Middleware)

	// get TLSConfig
	tlsConfig, manager := GetTLSConfig()
	// create the server,
	srv := &APIRouter{http.Server{
		Addr:      secured_port,
		TLSConfig: tlsConfig,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}, r, logger}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			logger.Fatal("Server stop with error", zap.Error(err))
		}
	}()

	// allow ACME call to be performed
	go func() {
		if !config.Config().IsDev() && manager != nil {
			if err := http.ListenAndServe(port, manager.HTTPHandler(nil)); err != nil {
				logger.Fatal("Server stop with error", zap.Error(err))
			}
		}
	}()

	return srv
}
