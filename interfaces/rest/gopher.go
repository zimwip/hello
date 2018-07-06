package rest

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/zimwip/hello/domain"
)

type GopherHandler struct {
	lock    *sync.Mutex
	counter int
	gophers map[Session]*domain.GopherInfo
}

func (handler *GopherHandler) HandleConnect(s Session) {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	for _, info := range handler.gophers {
		s.WriteMessage([]byte("set " + info.ID + " " + info.X + " " + info.Y))
	}
	handler.gophers[s] = &domain.GopherInfo{strconv.Itoa(handler.counter), "0", "0"}
	s.WriteMessage([]byte("iam " + handler.gophers[s].ID))
	handler.counter++
}

func (handler *GopherHandler) HandleDisconnect(s Session) {
	handler.lock.Lock()
	defer handler.lock.Unlock()
	s.BroadcastOthers([]byte("dis " + handler.gophers[s].ID))
	delete(handler.gophers, s)
}

func (handler *GopherHandler) HandleMessage(s Session, msg []byte) {
	p := strings.Split(string(msg), " ")
	handler.lock.Lock()
	defer handler.lock.Unlock()
	info := handler.gophers[s]
	if len(p) == 2 {
		info.X = p[0]
		info.Y = p[1]
		s.BroadcastOthers([]byte("set " + info.ID + " " + info.X + " " + info.Y))
	}
}

func (handler *GopherHandler) HandleError(s Session, err error) {
	fmt.Printf("%s", err)
}

func NewGopher(routeInteractor RouteInteractor) {
	handler := &GopherHandler{
		gophers: make(map[Session]*domain.GopherInfo),
		lock:    new(sync.Mutex),
	}
	routeInteractor.AddWebsocketHandler("Websocket", "/ws", handler)
}
