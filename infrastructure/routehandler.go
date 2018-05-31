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

//NewRouter returns a new Gorrila Mux router
func NewRouter(c rest.AppContext) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	appContext = c

	// Now websocket test
	mrouter := melody.New()
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
