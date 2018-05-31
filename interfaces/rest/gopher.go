package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/zimwip/hello/domain"
)

type Session interface {
	Write(data []byte)
	BroadcastOthers(data []byte) error
}

type WebsocketHandler struct {
	lock       *sync.Mutex
	counter    int
	gophers    map[Session]*domain.GopherInfo
	interactor WebsocketInteractor
}

type WebsocketInteractor interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) error
	RegisterHandler(handler *WebsocketHandler)
}

func (handler *WebsocketHandler) HandleConnect(s Session) {
	handler.lock.Lock()
	for _, info := range handler.gophers {
		s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
	}
	handler.gophers[s] = &domain.GopherInfo{strconv.Itoa(handler.counter), "0", "0"}
	s.Write([]byte("iam " + handler.gophers[s].ID))
	handler.counter++
	handler.lock.Unlock()
}

func (handler *WebsocketHandler) HandleDisconnect(s Session) {
	handler.lock.Lock()
	s.BroadcastOthers([]byte("dis " + handler.gophers[s].ID))
	delete(handler.gophers, s)
	handler.lock.Unlock()
}
func (handler *WebsocketHandler) HandleMessage(s Session, msg []byte) {
	p := strings.Split(string(msg), " ")
	handler.lock.Lock()
	info := handler.gophers[s]
	if len(p) == 2 {
		info.X = p[0]
		info.Y = p[1]
		s.BroadcastOthers([]byte("set " + info.ID + " " + info.X + " " + info.Y))
	}
	handler.lock.Unlock()
}

func (handler *WebsocketHandler) HandleError(s Session, err error) {
	fmt.Printf("%s", err)
}

func NewGopher(context *AppContext, interactor WebsocketInteractor) {
	handler := &WebsocketHandler{
		gophers: make(map[Session]*domain.GopherInfo),
		lock:    new(sync.Mutex),
	}
	interactor.RegisterHandler(handler)

	contextedHandler := &ContextedHandler{
		AppContext: context,
		ContextedHandlerFunc: func(c *AppContext, w http.ResponseWriter, r *http.Request) {
			interactor.HandleRequest(w, r)
		},
	}

	route := Route{
		Name:             "Websocket",
		Method:           []string{}, //You can handle more than just GET requests here, but for this tutorial we'll just do GETs
		Pattern:          "/ws",
		ContextedHandler: contextedHandler, // We defined HelloWorldHandler in Part1
	}
	AddRoute(route)
}
