package mezvaro

import (
	"net/http"
	"sync"

	"math"
	"time"

	"golang.org/x/net/context"
)

// MaxHandlers number of handlers supported. Theoretically there is no hard limit
// in code, this value is only used as marker for aborting middleware chain. In the
// off chance that it is needed, this can be increased.
const MaxHandlers = int(math.MaxInt16)

// Context is main way of communication between handlers and with outside world.
// Context instance carries http.Request and http.ResponseWriter objects, implements
// x/net/context with all its features and provides some utility functions.
// Same context object is shared between all middlewares in chain.
type Context struct {
	context.Context
	Response     http.ResponseWriter
	Request      *http.Request
	handlerChain []Handler
	index        int
	urlParams    map[string]string
	netCtx       context.Context
	mu           sync.Mutex
}

func newContext(
	w http.ResponseWriter, r *http.Request,
	handlerChain []Handler, urlParams map[string]string) *Context {
	return &Context{
		Response:     w,
		Request:      r,
		index:        -1,
		handlerChain: handlerChain,
		urlParams:    urlParams,
		netCtx:       context.Background(),
	}
}

// Next invokes next handler in middleware chain. All middlewares should call
// this or Abort method at some point of execution and Next should be called only
// once. It is undefined what happens if Next is called more then once in same
// handler.
func (c *Context) Next() {
	c.index++
	s := len(c.handlerChain)
	for ; c.index < s; c.index++ {
		c.handlerChain[c.index].Handle(c)
	}
}

// Abort stops middleware chain from executing. After Abort has been called, no
// more middlewares will be called.
func (c *Context) Abort() {
	c.index = MaxHandlers
}

// IsAborted returns boolean that indicates if middleware chain has been aborted.
func (c *Context) IsAborted() bool {
	return c.index >= MaxHandlers
}

// UrlParam returns parameter from URL Path by name. If parameter with required
// name does not exist, empty string is returned.
func (c *Context) UrlParam(name string) string {
	// map is read only, so it should be safe for concurrent access
	return c.urlParams[name]
}

/////////////////////////////////////////////
// net/context implementation
/////////////////////////////////////////////

// Deadline implements net/context.Context.Deadline by delegating the call.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.netCtx.Deadline()
}

// Done implements net/context.Context.Deadline by delegating the call.
func (c *Context) Done() <-chan struct{} {
	return c.netCtx.Done()
}

// Err implements net/context.Context.Deadline by delegating the call.
func (c *Context) Err() error {
	return c.netCtx.Err()
}

// Value implements net/context.Context.Deadline by delegating the call.
func (c *Context) Value(key interface{}) interface{} {
	return c.netCtx.Value(key)
}

// WithCancel updates context's Done channel to be closed when returned cancel
// function is called or when parent context closes channel, whichever happens
// first.
//
// Canceling context releases resources associated with it, so code should call
// cancel as soon as operation running in this Context completes.
func (c *Context) WithCancel() (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cancelContext, cancelFunc := context.WithCancel(c.netCtx)
	c.netCtx = cancelContext
	return cancelFunc
}

// WithDeadline updates contest with deadline adjusted to be no later then d.
// If deadline is later then already set deadline, semantically nothing changes.
// Context Done channel is closed when deadline expires, when returned cancel
// function is returned or when parents Done channel is closed, whichever happens
// first.
//
// Canceling context releases resources associated with it, so code should call
// cancel as soon as operation running in this Context completes.
func (c *Context) WithDeadline(deadline time.Time) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	deadlineContext, cancelFunc := context.WithDeadline(c.netCtx, deadline)
	c.netCtx = deadlineContext
	return cancelFunc
}

// WithTimeout returns WithDeadline(time.Now().Add(timeout)).
//
// Canceling context releases resources associated with it, so code should call
// cancel as soon as operation running in this Context completes.
func (c *Context) WithTimeout(timeout time.Duration) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	timeoutContext, cancelFunc := context.WithTimeout(c.netCtx, timeout)
	c.netCtx = timeoutContext
	return cancelFunc
}

// WithValue sets value to context associated with provided key.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
func (c *Context) WithValue(key interface{}, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	valueContext := context.WithValue(c.netCtx, key, val)
	c.netCtx = valueContext
}
