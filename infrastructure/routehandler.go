package infrastructure

import (
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/olahol/melody.v1"

	"github.com/zimwip/hello/interfaces/rest"
)

var (
	appContext rest.AppContext
)

type Session struct {
	session *melody.Session
	melody  *melody.Melody
}

func (s *Session) BroadcastOthers(msg []byte) error {
	err := s.melody.BroadcastOthers(msg, s.session)
	return err
}

func (s *Session) Write(msg []byte) {
	s.Write(msg)
}

type Handler struct {
	*melody.Melody
}

func (h *Handler) RegisterHandler(handler *rest.WebsocketHandler) {
	h.HandleConnect(func(s *melody.Session) {
		handler.HandleConnect(&Session{session: s, melody: h.Melody})
	})
}

//NewRouter returns a new Gorrila Mux router
func NewRouter(c rest.AppContext) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	appContext = c

	// Now websocket test
	mrouter := &Handler{melody.New()}
	rest.NewGopher(&appContext, mrouter)

	for _, route := range rest.GetRoutes() {
		//Check all routes to make sure the users are properly authenticated
		router.
			Methods(route.Method...).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.ContextedHandler)
	}
	return router
}
