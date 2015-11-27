package mezvaro

import "net/http"

// Handler defines interface for Mezvaro middlewares and handlers.
type Handler interface {
	// Handle does real job of processing request. It receives context
	// in which it is executing in.
	Handle(Context)
}

// HandlerFunc is function that implements Handler interface.
type HandlerFunc func(Context)

// Handle is implementation of Handler interface for HandlerFunc type.
func (hf HandlerFunc) Handle(c Context) {
	hf(c)
}

// Mezvaro is simply chain of handlers that will be executed in order they are added.
type Mezvaro struct {
	handlerChain []Handler
}

// New creates new instance of Mezvaro with provided handlers.
func New(handlers ...Handler) *Mezvaro {
	return &Mezvaro{
		handlerChain: handlers,
	}
}

// Use adds new handler to used instance of Mezvaro.
func (m *Mezvaro) Use(handler ...Handler) *Mezvaro {
	m.handlerChain = append(m.handlerChain, handler...)
	return m
}

// UseFunc adds function that matches signature of HandlerFunc to used instance
// of Mezvaro.
func (m *Mezvaro) UseFunc(handlerFuncs ...func(Context)) *Mezvaro {
	handlers := make([]Handler, 0, len(handlerFuncs))

	for _, h := range handlerFuncs {
		handlers = append(handlers, HandlerFunc(h))
	}
	m.Use(handlers...)
	return m
}

// UseHandlerMiddleware adds handler defined in format popular in Go community to used instance
// Mezvaro. This format (func (http.Handler) http.Handler) is popular in bunch of
// other frameworks, and a lot of useful middlewares exist out there.
func (m *Mezvaro) UseHandlerMiddleware(middleware func(http.Handler) http.Handler) *Mezvaro {
	m.Use(WrapHandlerMiddleware(middleware))
	return m
}

// UseHandler adds handler in standard library format to chain of handlers.
func (m *Mezvaro) UseHandler(handler http.Handler) *Mezvaro {
	m.Use(WrapHandler(handler))
	return m
}

// UseHandlerFunc adds handler function in standard library format to chain of handlers.
func (m *Mezvaro) UseHandlerFunc(handler func(http.ResponseWriter, *http.Request)) *Mezvaro {
	m.Use(WrapHandlerFunc(handler))
	return m
}

// Fork creates new instance of Mezvaro with copied handlers from current instance
// and added new provided handlers.
func (m *Mezvaro) Fork(handlers ...Handler) *Mezvaro {
	n := make([]Handler, 0, len(m.handlerChain)+len(handlers))
	n = append(n, m.handlerChain...)
	n = append(n, handlers...)
	return New(n...)
}

// ServeHTTP implements http.Handler interface.
func (m *Mezvaro) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r, m.handlerChain, nil)
	c.Next()
}

// WrapHandlerMiddleware wraps middleware defined in format popular in bunch
// of other Go frameworks to Handler compatible with Mezvaro.
func WrapHandlerMiddleware(middleware func(http.Handler) http.Handler) Handler {
	fn := func(c Context) {
		var calledNext bool
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: replace response and request in context, since from
			// now on we should use objects provided by middleware
			calledNext = true
			c.Next()
		}))
		handler.ServeHTTP(c.Response(), c.Request())
		if !calledNext {
			// standard way of aborting chain for this style of middleware is
			// not to call next handler, so if next handler was not called,
			// we abort our chain
			c.Abort()
		}
	}
	return HandlerFunc(fn)
}

// WrapHandler wraps standard library handler to Mezvaro handler. This handler
// can be used as middleware (next middleware is automatically called) or it
// can be used as final handler.
func WrapHandler(handler http.Handler) Handler {
	return HandlerFunc(func(c Context) {
		handler.ServeHTTP(c.Response(), c.Request())
		c.Next()
	})
}

// WrapHandlerFunc wraps standard library handler function to Mezvaro handler. This
// handler can be used as middleware (next middleware is automatically called) or
// it can be used as final handler.
func WrapHandlerFunc(handler func(http.ResponseWriter, *http.Request)) Handler {
	return WrapHandler(http.HandlerFunc(handler))
}
