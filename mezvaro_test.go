package mezvaro

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEmpty(t *testing.T) {
	// just make sure there is no boom
	m := New()
	if len(m.handlerChain) > 0 {
		t.Fatal("Fresh instance is not empty")
	}
	m.ServeHTTP(httptest.NewRecorder(), nil)
}

func TestCreate(t *testing.T) {
	m := New(HandlerFunc(func(c *Context) {}))
	if len(m.handlerChain) != 1 {
		t.Fatal("Expected 1 handler, found: ", len(m.handlerChain))
	}
}

func TestUse(t *testing.T) {
	m := New()
	m.Use(
		HandlerFunc(func(c *Context) {}),
		HandlerFunc(func(c *Context) {}),
	)
	if len(m.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers, found: ", len(m.handlerChain))
	}
}

func TestUseFunc(t *testing.T) {
	m := New()
	m.UseFunc(
		func(c *Context) {},
		func(c *Context) {},
	)
	if len(m.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers, found: ", len(m.handlerChain))
	}
}

func TestUseHandler(t *testing.T) {
	m := New()
	m.UseHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	if len(m.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers, found: ", len(m.handlerChain))
	}
}

func TestUseHandlerFunc(t *testing.T) {
	m := New()
	m.UseHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
		func(w http.ResponseWriter, r *http.Request) {},
	)
	if len(m.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers, found: ", len(m.handlerChain))
	}
}

func TestUseHandlerMiddleware(t *testing.T) {
	m := New()
	m.UseHandlerMiddleware(
		func(h http.Handler) http.Handler { return nil },
		func(h http.Handler) http.Handler { return nil },
	)
	if len(m.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers, found: ", len(m.handlerChain))
	}
}

func TestForkHandlerCount(t *testing.T) {
	original := HandlerFunc(func(c *Context) {})
	forkHandler := HandlerFunc(func(c *Context) {})
	m := New(original)
	fork := m.Fork(forkHandler)
	if len(fork.handlerChain) != 2 {
		t.Fatal("Expected 2 handlers in fork mezvaro, found: ", len(fork.handlerChain))
	}
}

func TestServeHTTP(t *testing.T) {
	var called bool
	m := New(HandlerFunc(
		func(c *Context) {
			called = true
		}))
	response := httptest.NewRecorder()
	m.ServeHTTP(response, nil)
	if !called {
		t.Fatal("Handler not called.")
	}
}

func TestHandle(t *testing.T) {
	var called bool
	m := New(HandlerFunc(
		func(c *Context) {
			called = true
		}))
	response := httptest.NewRecorder()
	ctx := &Context{Response: response}
	m.Handle(ctx)
	if !called {
		t.Fatal("Handler not called.")
	}
}

func TestWrapHandlerMiddleware(t *testing.T) {
	var called bool
	middleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			h.ServeHTTP(w, r)
		})
	}

	handler := WrapHandlerMiddleware(middleware)
	ctx := &Context{Response: httptest.NewRecorder()}
	handler.Handle(ctx)
	if !called {
		t.Fatal("Wrapping middleware not called.")
	}
}

func TestWrapHandlerMiddlewareReplaceResponseRequest(t *testing.T) {
	originalResponse := httptest.NewRecorder()
	originalRequest, _ := http.NewRequest("GET", "", nil)
	replacementResponse := httptest.NewRecorder()
	replacementRequest, _ := http.NewRequest("POST", "", nil)
	middleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(replacementResponse, replacementRequest)
		})
	}
	handler := WrapHandlerMiddleware(middleware)
	ctx := &Context{Response: originalResponse, Request: originalRequest}
	handler.Handle(ctx)
	if ctx.Response != replacementResponse {
		t.Fatal("Response not replaced by handler middleware.")
	}
	if ctx.Request != replacementRequest {
		t.Fatal("Request not replaced by handler middleware.")
	}
}

func TestWrapHandlerMiddlewareAbort(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "", nil)
	middleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// do not call "h", next middleware in chain
		})
	}
	handler := WrapHandlerMiddleware(middleware)
	ctx := &Context{Response: response, Request: request}
	handler.Handle(ctx)
	if !ctx.IsAborted() {
		t.Fatal("Response should be aborted.")
	}
}

func TestWrapHandler(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "", nil)
	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	mezvaroHandler := WrapHandler(handler)
	ctx := &Context{Response: response, Request: request}
	mezvaroHandler.Handle(ctx)
	if !called {
		t.Fatal("Handler not called.")
	}
	if ctx.index == -1 {
		t.Fatal("Next handler not called.")
	}
}

func TestWrapHandlerFunc(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "", nil)
	var called bool
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}
	mezvaroHandler := WrapHandlerFunc(handlerFunc)
	ctx := &Context{Response: response, Request: request}
	mezvaroHandler.Handle(ctx)
	if !called {
		t.Fatal("Handler func not called.")
	}
	if ctx.index == -1 {
		t.Fatal("Next handler in line not called.")
	}
}

func TestDefaultParamsExtractor(t *testing.T) {
	var urlParams map[string]string
	m := New(HandlerFunc(func(c *Context) {
		urlParams = c.urlParams
	}))
	m.ServeHTTP(httptest.NewRecorder(), nil)
	if urlParams != nil {
		t.Fatal("Default URL parameters extractors did not return nil.")
	}
}

func TestCustomParamsExtractor(t *testing.T) {
	var urlParams map[string]string
	m := New(HandlerFunc(func(c *Context) {
		urlParams = c.urlParams
	}))
	extractor := func(r *http.Request) map[string]string {
		return map[string]string{
			"param": "value",
		}
	}
	SetUrlParamsExtractor(extractor)
	m.ServeHTTP(httptest.NewRecorder(), nil)
	if val, ok := urlParams["param"]; !ok {
		t.Fatal("URL parameters key do not match.")
	} else {
		if val != "value" {
			t.Fatal("URL parameter value does not match.")
		}
	}
}
