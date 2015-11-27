package mezvaro

import (
	"net/http"
	"sync"

	"math"
	"time"

	"golang.org/x/net/context"
)

const MaxHandlers = int(math.MaxInt16)

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

func (c *Context) netContext() context.Context {
	return c.netCtx
}

func (c *Context) setNetContext(netContext context.Context) {
	c.netCtx = netContext
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlerChain)
	for ; c.index < s; c.index++ {
		c.handlerChain[c.index].Handle(c)
	}
}

func (c *Context) Abort() {
	c.index = MaxHandlers
}

func (c *Context) IsAborted() bool {
	return c.index >= MaxHandlers
}

func (c *Context) UrlParam(name string) string {
	// map is read only, so it should be safe for concurrent access
	return c.urlParams[name]
}

/////////////////////////////////////////////
// net/context implementation
/////////////////////////////////////////////
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.netCtx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.netCtx.Done()
}

func (c *Context) Err() error {
	return c.netCtx.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.netCtx.Value(key)
}

/////////////////////////////////////////////
// create new contexts functions
/////////////////////////////////////////////
func (c *Context) WithCancel() (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cancelContext, cancelFunc := context.WithCancel(c.netContext())
	c.setNetContext(cancelContext)
	return cancelFunc
}

func (c *Context) WithDeadline(deadline time.Time) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	deadlineContext, cancelFunc := context.WithDeadline(c.netContext(), deadline)
	c.setNetContext(deadlineContext)
	return cancelFunc
}

func (c *Context) WithTimeout(timeout time.Duration) (cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	timeoutContext, cancelFunc := context.WithTimeout(c.netContext(), timeout)
	c.setNetContext(timeoutContext)
	return cancelFunc
}

func (c *Context) WithValue(key interface{}, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	valueContext := context.WithValue(c.netContext(), key, val)
	c.setNetContext(valueContext)
}
