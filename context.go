package mezvaro

import (
	"net/http"

	"math"
	"time"

	"golang.org/x/net/context"
)

const MaxHandlers = int(math.MaxInt16)

type Context interface {
	context.Context
	Response() http.ResponseWriter
	Request() *http.Request
	netContext() context.Context
	setNetContext(context.Context)
	Next()
	Abort()
	IsAborted() bool
	UrlParam(string) string
	WithValue(key interface{}, val interface{})
}

type ctx struct {
	netCtx       context.Context
	response     http.ResponseWriter
	request      *http.Request
	handlerChain []Handler
	index        int
	urlParams    map[string]string
}

func newContext(
	w http.ResponseWriter, r *http.Request,
	handlerChain []Handler, urlParams map[string]string) Context {
	return &ctx{
		netCtx:       context.Background(),
		response:     w,
		request:      r,
		index:        -1,
		handlerChain: handlerChain,
		urlParams:    urlParams,
	}
}

func (c *ctx) Response() http.ResponseWriter {
	return c.response
}

func (c *ctx) Request() *http.Request {
	return c.request
}

func (c *ctx) netContext() context.Context {
	return c.netCtx
}

func (c *ctx) setNetContext(netContext context.Context) {
	c.netCtx = netContext
}

func (c *ctx) Next() {
	c.index++
	s := len(c.handlerChain)
	for ; c.index < s; c.index++ {
		c.handlerChain[c.index].Handle(c)
	}
}

func (c *ctx) Abort() {
	c.index = MaxHandlers
}

func (c *ctx) IsAborted() bool {
	return c.index >= MaxHandlers
}

func (c *ctx) UrlParam(name string) string {
	// map is read only, so it should be safe for concurent access
	return c.urlParams[name]
}

/////////////////////////////////////////////
// net/context implementation
/////////////////////////////////////////////
func (c *ctx) Deadline() (deadline time.Time, ok bool) {
	return c.netCtx.Deadline()
}

func (c *ctx) Done() <-chan struct{} {
	return c.netCtx.Done()
}

func (c *ctx) Err() error {
	return c.netCtx.Err()
}

func (c *ctx) Value(key interface{}) interface{} {
	return c.netCtx.Value(key)
}

/////////////////////////////////////////////
// create new contexts functions
/////////////////////////////////////////////
//func (c *ctx) WithCancel() (cancel context.CancelFunc) {
//
//}

func (c *ctx) WithCancel() (cancel context.CancelFunc) {
	cancelContext, cancelFunc := context.WithCancel(c.netContext())
	c.setNetContext(cancelContext)
	return cancelFunc
}

func WithCancel(parent Context) (ctx Context, cancel context.CancelFunc) {
	cancelContext, cancelFunc := context.WithCancel(parent.netContext())
	parent.setNetContext(cancelContext)
	return parent, cancelFunc
}

func (c *ctx) WithDeadline(deadline time.Time) (cancel context.CancelFunc) {
	deadlineContext, cancelFunc := context.WithDeadline(c.netContext(), deadline)
	c.setNetContext(deadlineContext)
	return cancelFunc
}

func WithDeadline(parent Context, deadline time.Time) (Context, context.CancelFunc) {
	deadlineContext, cancelFunc := context.WithDeadline(parent.netContext(), deadline)
	parent.setNetContext(deadlineContext)
	return parent, cancelFunc
}

func (c *ctx) WithTimeout(timeout time.Duration) (cancel context.CancelFunc) {
	timeoutContext, cancelFunc := context.WithTimeout(c.netContext(), timeout)
	c.setNetContext(timeoutContext)
	return cancelFunc
}

func WithTimeout(parent Context, timeout time.Duration) (Context, context.CancelFunc) {
	timeoutContext, cancelFunc := context.WithTimeout(parent.netContext(), timeout)
	parent.setNetContext(timeoutContext)
	return parent, cancelFunc
}

func (c *ctx) WithValue(key interface{}, val interface{}) {
	valueContext := context.WithValue(c.netContext(), key, val)
	c.setNetContext(valueContext)
}

func WithValue(parent Context, key interface{}, val interface{}) Context {
	valueContext := context.WithValue(parent.netContext(), key, val)
	parent.setNetContext(valueContext)
	return parent
}
