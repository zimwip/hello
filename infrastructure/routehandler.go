package infrastructure

import (
	"log"

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

func (s *Session) WriteMessage(msg []byte) {
	s.session.Write(msg)
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
func NewRouter(appContext *rest.AppContext) *mux.Router {
	subRoutes := make(map[string]*mux.Router)
	router := mux.NewRouter().StrictSlash(true)
	cur := router

	// Now websocket test
	mrouter := &Handler{melody.New()}
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
		log.Printf("Adding Route %s at %s%s, %p", route.Name, route.ParentRoute, route.Pattern, cur)
		cur.
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.ContextedHandler)
		if len(route.Method) > 0 {
			cur.Methods(route.Method...)
		}
	}
	return router
}
