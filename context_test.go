package mezvaro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "", nil)

	c := newContext(response, request, nil, nil)
	if c.Response != response {
		t.Fatal("Response not stored in context.")
	}
	if c.Request != request {
		t.Fatal("Request not stored in context.")
	}
}

func TestNext(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GEt", "", nil)
	firstCount := 0
	secondCount := 0
	handlerChain := []Handler{
		HandlerFunc(func(c *Context) {
			firstCount++
			c.Next()
		}),
		HandlerFunc(func(c *Context) {
			secondCount++
			c.Next()
		}),
	}
	c := newContext(response, request, handlerChain, nil)
	c.Next()
	if firstCount != 1 {
		t.Fatal("First handler not called or called more then once.")
	}
	if secondCount != 1 {
		t.Fatal("Second handler not called or called more then once.")
	}
}

func TestNextWithoutExplicitCall(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GEt", "", nil)
	firstCount := 0
	secondCount := 0
	handlerChain := []Handler{
		HandlerFunc(func(c *Context) {
			firstCount++
		}),
		HandlerFunc(func(c *Context) {
			secondCount++
		}),
	}
	c := newContext(response, request, handlerChain, nil)
	c.Next()
	if firstCount != 1 {
		t.Fatal("First handler not called or called more then once.")
	}
	if secondCount != 1 {
		t.Fatal("Second handler not called or called more then once.")
	}
}

func TestAbort(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GEt", "", nil)
	firstCount := 0
	secondCount := 0
	handlerChain := []Handler{
		HandlerFunc(func(c *Context) {
			firstCount++
			c.Abort()
		}),
		HandlerFunc(func(c *Context) {
			secondCount++
			c.Next()
		}),
	}
	c := newContext(response, request, handlerChain, nil)
	c.Next()
	if firstCount != 1 {
		t.Fatal("First handler not called or called more then once.")
	}
	if secondCount != 0 {
		t.Fatal("Second handler called and it should not be.")
	}
	if !c.IsAborted() {
		t.Fatal("Context not aborted and it should be.")
	}
}

type netContext struct {
	deadlineCalled bool
	doneCalled     bool
	errCalled      bool
	valueCalled    bool
	deadline       time.Time
	ok             bool
	done           chan struct{}
	err            error
	data           map[interface{}]interface{}
}

func (nc *netContext) Deadline() (deadline time.Time, ok bool) {
	nc.deadlineCalled = true
	return nc.deadline, nc.ok
}

func (nc *netContext) Done() <-chan struct{} {
	nc.doneCalled = true
	return nc.done
}

func (nc *netContext) Err() error {
	nc.errCalled = true
	return nc.err
}

func (nc *netContext) Value(key interface{}) interface{} {
	nc.valueCalled = true
	return nc.data[key]
}

func TestDeadline(t *testing.T) {
	deadline := time.Now()
	netCtx := &netContext{deadline: deadline, ok: true}
	c := &Context{netCtx: netCtx}
	ctxDeadline, ok := c.Deadline()
	if !netCtx.deadlineCalled {
		t.Fatal("Net context deadline not called")
	}
	if deadline != ctxDeadline {
		t.Fatal("Got wrong deadline value.")
	}
	if !ok {
		t.Fatal("Got wrong ok value.")
	}
}

func TestDone(t *testing.T) {
	doneCh := make(chan struct{})
	netCtx := &netContext{done: doneCh}
	c := &Context{netCtx: netCtx}
	ctxDone := c.Done()
	if !netCtx.doneCalled {
		t.Fatal("Net context done not called.")
	}
	if ctxDone != doneCh {
		t.Fatal("Got wrong done channel.")
	}
}

func TestErr(t *testing.T) {
	err := fmt.Errorf("ErrorMsg")
	netCtx := &netContext{err: err}
	c := &Context{netCtx: netCtx}
	ctxErr := c.Err()
	if !netCtx.errCalled {
		t.Fatal("Net context error not called.")
	}
	if ctxErr != err {
		t.Fatal("Got wrong error value.")
	}
}

func TestValue(t *testing.T) {
	vals := map[interface{}]interface{}{
		"key": "value",
	}
	netCtx := &netContext{data: vals}
	c := &Context{netCtx: netCtx}
	ctxValue := c.Value("key")
	if !netCtx.valueCalled {
		t.Fatal("Net context value not called.")
	}
	if v, ok := ctxValue.(string); ok {
		if v != "value" {
			t.Fatal("Got wrong value from context.")
		}
	} else {
		t.Fatal("Value extracted from context has wrong type.")
	}
}
