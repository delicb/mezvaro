package mezvaro

import (
	"net/http"
	"sync"
)

// Handler defines interface for Mezvaro middlewares and handlers.
type Handler interface {
	// Handle does real job of processing request. It receives context
	// in which it is executing in.
	Handle(*Context)
}

// HandlerFunc is function that implements Handler interface.
type HandlerFunc func(*Context)

// Handle is implementation of Handler interface for HandlerFunc type.
func (hf HandlerFunc) Handle(c *Context) {
	hf(c)
}

// URLParamsExtractor is function that extracts mutable parts or URL.
// Intended use of this is to allow creation of adapters for various routers.
type URLParamsExtractor func(*http.Request) map[string]string

// defaultURLParamsExtractor works with standard library multiplexer that does not
// support URL parameters, so it only returns nil.
func defaultURLParamsExtractor(r *http.Request) map[string]string {
	return nil
}

var (
	paramsExtractorLock sync.Mutex
	urlParamsExtractor  = defaultURLParamsExtractor
)

// SetURLParamsExtractor sets function that returns map of mutable parts of URL.
func SetURLParamsExtractor(extractor URLParamsExtractor) {
	paramsExtractorLock.Lock()
	defer paramsExtractorLock.Unlock()
	urlParamsExtractor = extractor
}

// Mezvaro is simply chain of handlers that will be executed in order they are added.
type Mezvaro struct {
	parent       *Mezvaro
	handlerChain []Handler
}

// New creates new instance of Mezvaro with provided handlers.
func New(handlers ...Handler) *Mezvaro {
	return &Mezvaro{
		parent:       nil,
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
func (m *Mezvaro) UseFunc(handlerFuncs ...func(*Context)) *Mezvaro {
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
func (m *Mezvaro) UseHandlerMiddleware(middleware ...func(http.Handler) http.Handler) *Mezvaro {
	mezvaroMiddlewares := make([]Handler, 0, len(middleware))
	for _, h := range middleware {
		mezvaroMiddlewares = append(mezvaroMiddlewares, WrapHandlerMiddleware(h))
	}
	m.Use(mezvaroMiddlewares...)
	return m
}

// UseHandler adds handler in standard library format to chain of handlers.
func (m *Mezvaro) UseHandler(handlers ...http.Handler) *Mezvaro {
	mezvaroHandlers := make([]Handler, 0, len(handlers))
	for _, h := range handlers {
		mezvaroHandlers = append(mezvaroHandlers, WrapHandler(h))
	}
	m.Use(mezvaroHandlers...)
	return m
}

// UseHandlerFunc adds handler function in standard library format to chain of handlers.
func (m *Mezvaro) UseHandlerFunc(handlers ...func(http.ResponseWriter, *http.Request)) *Mezvaro {
	mezvaroHandlers := make([]Handler, 0, len(handlers))
	for _, h := range handlers {
		mezvaroHandlers = append(mezvaroHandlers, WrapHandlerFunc(h))
	}
	m.Use(mezvaroHandlers...)
	return m
}

// Fork creates new instance of Mezvaro with copied handlers from current instance
// and added new provided handlers.
func (m *Mezvaro) Fork(handlers ...Handler) *Mezvaro {
	return &Mezvaro{
		parent:       m,
		handlerChain: handlers,
	}
	//	n := make([]Handler, 0, len(m.handlerChain)+len(handlers))
	//	n = append(n, m.handlerChain...)
	//	n = append(n, handlers...)
	//	return New(n...)
}

// wholeChain returns whole chain of handlers including this Mezvaro instance
// and all its parents.
func (m *Mezvaro) wholeChain() []Handler {
	// count number of handler in entire chain first, to allocate slice
	// of right size right away.
	var handlerNo int
	current := m
	// capacity of 5 is just a guess, most of the time no more the 5
	// instance will be in chain, so this should be enough to avoid
	// new allocations during append.
	parents := make([]*Mezvaro, 0, 5)
	for current != nil {
		parents = append(parents, current)
		handlerNo += len(current.handlerChain)
		current = current.parent
	}
	// allocate slice with array of appropriate size
	// this prevents dynamic expansion of array during appending
	handlers := make([]Handler, 0, handlerNo)
	// traverse parents in revers order, since that order of middlewares
	// is expected
	for i := len(parents) - 1; i >= 0; i-- {
		handlers = append(handlers, parents[i].handlerChain...)
	}
	return handlers
}

// H builds entire chain of middlewares and adds provided handler at the end.
// This function exists for optimisation, to avoid building middleware
// chain in runtime, so we are building it at boot up time.
func (m *Mezvaro) H(h Handler) http.Handler {
	wholeChain := append(m.wholeChain(), h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := newContext(w, r, wholeChain, urlParamsExtractor(r))
		c.Next()
	})
}

// HF builds entire chain of middlewares and adds provided handler func at the end.
// this function exists for optimization, to avoid building middleware
// chain in runtime, so we are building it at boot time.
func (m *Mezvaro) HF(h func(*Context)) http.Handler {
	return m.H(HandlerFunc(h))
}

// ServeHTTP implements http.Handler interface.
func (m *Mezvaro) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r, m.wholeChain(), urlParamsExtractor(r))
	c.Next()
}

// Handle implements Handler interface.
func (m *Mezvaro) Handle(c *Context) {
	// Reuse provided context, since request and response has to be the same
	// and stuff like timeout and deadline has to be preserved.
	c.handlerChain = m.wholeChain()
	c.index = -1
	c.Next()
}

// WrapHandlerMiddleware wraps middleware defined in format popular in bunch
// of other Go frameworks to Handler compatible with Mezvaro.
func WrapHandlerMiddleware(middleware func(http.Handler) http.Handler) Handler {
	fn := func(c *Context) {
		var calledNext bool
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calledNext = true
			// replace response and request objects with one provided from middleware,
			// since middleware might want to replace them with something similar
			c.Response = w
			c.Request = r
			c.Next()
		}))
		handler.ServeHTTP(c.Response, c.Request)
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
	return HandlerFunc(func(c *Context) {
		handler.ServeHTTP(c.Response, c.Request)
		c.Next()
	})
}

// WrapHandlerFunc wraps standard library handler function to Mezvaro handler. This
// handler can be used as middleware (next middleware is automatically called) or
// it can be used as final handler.
func WrapHandlerFunc(handler func(http.ResponseWriter, *http.Request)) Handler {
	return WrapHandler(http.HandlerFunc(handler))
}
