/*
Basic route management tools.
*/
package rest

import (
	"net/http"

	"github.com/zimwip/hello/domain"
)

//ContextedHandler is a wrapper to provide AppContext to our Handlers
type ContextedHandler struct {
	*domain.AppContext
	//ContextedHandlerFunc is the interface which our Handlers will implement
	ContextedHandlerFunc func(*domain.AppContext, http.ResponseWriter, *http.Request)
}

func (handler ContextedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.ContextedHandlerFunc(handler.AppContext, w, r)
}

//Route this struct is used for declaring a route
type Route struct {
	Name             string
	Method           []string
	Pattern          string
	ParentRoute      string
	ContextedHandler *ContextedHandler
}

//Routes just stores our Route declarations
type Routes []Route

var (
	routes Routes = make([]Route, 0)
)

func AddRoute(route Route) {
	routes = append(routes, route)
}

func GetRoutes() Routes {
	return routes
}

func DeclareNewRoute(context *domain.AppContext, name string, method []string, pattern string, parent string, handler func(c *domain.AppContext, w http.ResponseWriter, r *http.Request)) {

	contextedHandler := &ContextedHandler{
		AppContext:           context,
		ContextedHandlerFunc: handler,
	}

	route := Route{
		Name:             name,
		Method:           method,
		Pattern:          pattern,
		ParentRoute:      parent,
		ContextedHandler: contextedHandler, // We defined HelloWorldHandler in Part1
	}

	AddRoute(route)
}
