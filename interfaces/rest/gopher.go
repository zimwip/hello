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
}

type WebsocketHandler struct {
	lock       *sync.Mutex
	counter    int
	gophers    map[Session]*domain.GopherInfo
	interactor WebsocketInteractor
}

type WebsocketInteractor interface {
	HandleRequest(w http.ResponseWriter, r *http.Request) error
	HandleConnect(func(s Session))
	HandleDisconnect(func(s Session))
	HandleMessage(func(s Session, msg []byte))
	HandleError(func(s Session, err error))
	BroadcastOthers(data []byte, session Session)
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
	handler.interactor.BroadcastOthers([]byte("dis "+handler.gophers[s].ID), s)
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
		handler.interactor.BroadcastOthers([]byte("set "+info.ID+" "+info.X+" "+info.Y), s)
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
	interactor.HandleConnect(handler.HandleConnect)
	interactor.HandleDisconnect(handler.HandleDisconnect)
	interactor.HandleMessage(handler.HandleMessage)
	interactor.HandleError(handler.HandleError)

	webHandler := func(c *AppContext, w http.ResponseWriter, r *http.Request) {
		interactor.HandleRequest(w, r)
	}

	contextedHandler := &ContextedHandler{
		AppContext:           context,
		ContextedHandlerFunc: webHandler,
	}

	route := Route{
		"Websocket",
		//You can handle more than just GET requests here, but for this tutorial we'll just do GETs
		[]string{},
		"/ws",
		// We defined HelloWorldHandler in Part1
		contextedHandler,
	}
	AddRoute(route)
}
