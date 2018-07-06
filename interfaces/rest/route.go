package rest

import (
	"net/http"

	"github.com/zimwip/hello/domain"
)

type ContextedHandlerFunc func(c *domain.AppContext, w http.ResponseWriter, r *http.Request)

//ContextedHandler is a wrapper to provide AppContext to our Handlers
type ContextedHandler struct {
	*domain.AppContext
	//ContextedHandlerFunc is the interface which our Handlers will implement
	HandlerFunc ContextedHandlerFunc
}

// ServeHTTP Wrapper
func (handler ContextedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.HandlerFunc(handler.AppContext, w, r)
}

type Session interface {
	WriteMessage(data []byte)
	BroadcastOthers(data []byte) error
}

type WebsocketHandler interface {
	HandleConnect(s Session)
	HandleDisconnect(s Session)
	HandleMessage(s Session, msg []byte)
	HandleError(s Session, err error)
}

type WebsocketInteractor interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) error
	RegisterHandler(handler WebsocketHandler)
}

type RouteInteractor interface {
	AddRoute(name string, method []string, pattern string, parent string, handler ContextedHandlerFunc)
	AddWebsocketHandler(name string, pattern string, handler WebsocketHandler)
}
