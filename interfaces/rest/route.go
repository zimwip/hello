/*
Basic route management tools.
*/
package rest

import (
	"net/http"
)

//ContextedHandler is a wrapper to provide AppContext to our Handlers
type ContextedHandler struct {
	*AppContext
	//ContextedHandlerFunc is the interface which our Handlers will implement
	ContextedHandlerFunc func(*AppContext, http.ResponseWriter, *http.Request)
}

//AppContext provides the app context to handlers.  This *cannot* contain request-specific keys like
//sessionId or similar.  It is shared across requests.
type AppContext struct {
}

func (handler ContextedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.ContextedHandlerFunc(handler.AppContext, w, r)
}

//Route this struct is used for declaring a route
type Route struct {
	Name             string
	Method           []string
	Pattern          string
	ContextedHandler *ContextedHandler
}

//Routes just stores our Route declarations
type Routes []Route

var (
	routes     Routes = make([]Route, 1)
	appContext AppContext
)

func AddRoute(route Route) {
	routes = append(routes, route)
}

func GetRoutes() Routes {
	return routes
}
