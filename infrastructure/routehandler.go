package infrastructure

import (
	"github.com/gorilla/mux"
	"gopkg.in/olahol/melody.v1"

	"github.com/zimwip/hello/interfaces/rest"
)

var (
	appContext rest.AppContext
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

type Handler struct {
	*melody.Melody
	sessions map[*melody.Session]*Session
}

func (h *Handler) localSession(s *melody.Session) *Session {
	var localSession *Session
	if val, ok := h.sessions[s]; ok {
		localSession = val
	} else {
		localSession = &Session{session: s, melody: h.Melody}
		h.sessions[s] = localSession
	}
	return localSession
}

func (h *Handler) RegisterHandler(handler *rest.WebsocketHandler) {
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

//NewRouter returns a new Gorrila Mux router
func NewRouter(appContext *rest.AppContext) *mux.Router {
	subRoutes := make(map[string]*mux.Router)
	router := mux.NewRouter()
	cur := router

	// Now websocket test
	mrouter := &Handler{melody.New(), make(map[*melody.Session]*Session)}
	rest.NewGopher(appContext, mrouter)
	rest.NewAPI(appContext)

	for _, route := range rest.GetRoutes() {
		//Check all routes to make sure the users are properly authenticated
		cur = router
		if len(route.ParentRoute) > 0 {
			if val, present := subRoutes[route.ParentRoute]; present {
				cur = val
			} else {
				cur = router.PathPrefix(route.ParentRoute).Subrouter().StrictSlash(true)
				subRoutes[route.ParentRoute] = cur
			}
		}
		newRoute := cur.Path(route.Pattern).
			Name(route.Name).
			Handler(route.ContextedHandler)
		if len(route.Method) > 0 {
			newRoute = newRoute.Methods(route.Method...)
		}

	}
	return router
}
