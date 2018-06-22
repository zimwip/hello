package domain

//AppContext provides the app context to handlers.  This *cannot* contain request-specific keys like
//sessionId or similar.  It is shared across requests.
type AppContext struct {
}
