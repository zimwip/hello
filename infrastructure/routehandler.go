package infrastructure

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"gopkg.in/olahol/melody.v1"

	"github.com/zimwip/hello/crosscutting"
	"github.com/zimwip/hello/domain"
	"github.com/zimwip/hello/interfaces/rest"
	"github.com/zimwip/hello/middleware"
)

// Session struct for melody session
type Session struct {
	session *melody.Session
	melody  *melody.Melody
}

func (s *Session) BroadcastOthers(msg []byte) error {
	err := s.melody.BroadcastOthers(msg, s.session)
	return err
}

func (s *Session) WriteMessage(msg []byte) {
	s.session.Write(msg)
}

type rootHandler struct {
	route      *mux.Router
	subRoutes  map[string]*mux.Router
	appContext *domain.AppContext
	websocket  rest.WebsocketInteractor
}

type websocketInteractor struct {
	*melody.Melody
	sessions map[*melody.Session]*Session
}

func (h *websocketInteractor) localSession(s *melody.Session) *Session {
	var localSession *Session
	if val, ok := h.sessions[s]; ok {
		localSession = val
	} else {
		localSession = &Session{session: s, melody: h.Melody}
		h.sessions[s] = localSession
	}
	return localSession
}

func (h *websocketInteractor) RegisterHandler(handler rest.WebsocketHandler) {
	h.HandleConnect(func(s *melody.Session) {
		localSession := h.localSession(s)
		handler.HandleConnect(localSession)
	})
	h.HandleDisconnect(func(s *melody.Session) {
		localSession := h.localSession(s)
		handler.HandleDisconnect(localSession)
	})
	h.HandleMessage(func(s *melody.Session, msg []byte) {
		localSession := h.localSession(s)
		handler.HandleMessage(localSession, msg)
	})
	h.HandleError(func(s *melody.Session, err error) {
		localSession := h.localSession(s)
		handler.HandleError(localSession, err)
	})

}

func (h rootHandler) AddRoute(name string, method []string, pattern string, parent string, handler rest.ContextedHandlerFunc) {
	contextedHandler := &rest.ContextedHandler{h.appContext, handler}

	cur := h.route
	if len(parent) > 0 {
		if val, present := h.subRoutes[parent]; present {
			cur = val
		} else {
			cur = h.route.PathPrefix(parent).Subrouter().StrictSlash(true)
			h.subRoutes[parent] = cur
		}
	}
	newRoute := cur.Path(pattern).
		Name(name).
		Handler(contextedHandler)
	if len(method) > 0 {
		newRoute = newRoute.Methods(method...)
	}
}

func (h rootHandler) AddWebsocketHandler(name string, pattern string, handler rest.WebsocketHandler) {
	h.websocket.RegisterHandler(handler)

	contextedHandler := &rest.ContextedHandler{
		AppContext: h.appContext,
		HandlerFunc: func(c *domain.AppContext, w http.ResponseWriter, r *http.Request) {
			h.websocket.HandleRequest(w, r)
		},
	}
	h.route.Path(pattern).
		Name(name).
		Handler(contextedHandler)
}

//NewRouter returns a new Gorrila Mux router
func NewServer(appContext *domain.AppContext, secured_port string, port string, staticDir string) (rest.RouteInteractor, *http.Server) {
	// Now websocket test
	mrouter := &websocketInteractor{melody.New(), make(map[*melody.Session]*Session)}
	route := rootHandler{mux.NewRouter(), make(map[string]*mux.Router), appContext, mrouter}

	// Set up classic Negroni Middleware
	recovery := middleware.NewRecovery()
	recovery.Formatter = &middleware.HTMLPanicFormatter{}
	recovery.PrintStack = true

	logger := crosscutting.Logger()

	route.route.Use(middleware.NewTracer().Middleware)
	route.route.Use(recovery.Middleware)

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
		IsDevelopment: crosscutting.Config().IsDev(),
	})

	route.route.Use(secureMiddleware.Handler)

	// static route
	static := route.route.PathPrefix("/").Subrouter().StrictSlash(true)
	static.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	static.Use(middleware.NewStatic(http.Dir(staticDir)).Middleware)

	// get TLSConfig
	tlsConfig, manager := GetTLSConfig()
	// create the server,
	srv := &http.Server{
		Addr:      secured_port,
		TLSConfig: tlsConfig,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      route.route, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			logger.Fatal("Server stop with error", logger.Error(err))
		}
	}()

	// allow ACME call to be performed
	go func() {
		if !crosscutting.Config().IsDev() && manager != nil {
			if err := http.ListenAndServe(port, manager.HTTPHandler(nil)); err != nil {
				logger.Fatal("Server stop with error", logger.Error(err))
			}
		}
	}()
	return route, srv
}
