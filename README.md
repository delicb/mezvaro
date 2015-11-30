# Mezvaro [![Build Status](https://travis-ci.org/delicb/mezvaro.svg?branch=master)](https://travis-ci.org/delicb/mezvaro)[![Coverage](http://gocover.io/_badge/github.com/delicb/mezvaro)](http://gocover.io/github.com/delicb/mezvaro)[![GoDoc](http://godoc.org/github.com/delicb/mezvaro?status.png)](http://godoc.org/github.com/delicb/mezvaro)
Middleware management library for Golang. 

## Why?
There are bunch of great libraries for Golang when it comes to middleware managemant. So, why create another one? Most of the libraries are constrained by respecting [http.Handler](https://godoc.org/net/http#Handler) interface. This is fine, but I wanted to have full power of [context](https://godoc.org/golang.org/x/net/context) and I do not mind having different signature for my handlers then one defined in standard library.

Having context stored in global map somewhere defeates the purpose of having context at all IMHO and storing it in [ResponseWriter](https://godoc.org/net/http#ResponseWriter) wrapper is just bad workaround since context logically does not belong in response object of any kind, so aproach taken by Mezvaro is to provide context object as parameter to every handler and middleware. Context holds [ResponseWriter](https://godoc.org/net/http#ResponseWriter) and [*Request](https://godoc.org/net/http#Request) in it, with couple of other things.

[Net Context](https://godoc.org/golang.org/x/net/context#Context) has much more use cases then just for web services and goal was to keep that power. However, for web service development, having context in which handler is executed can potentially make bunch of stuff easier, since context is shared between middlewares and final handler. For other use cases that NetContext can be used for, Mezvaro context fully implements [Net Context](https://godoc.org/golang.org/x/net/context#Context) interface.

## So, mezvaro is not compatible with standard library handlers?
**It is.** All handlers that respect [http.Handler](https://godoc.org/net/http#Handler) interface can be used. Also, all middlewares that rely on this interface (they have signature `func(next http.Handler) http.Handler`) like [gorilla handlers](https://github.com/gorilla/handlers) can be used.

However, these handlers will not be able to use features that context provides, since they are not aware its existance. In order to use features that context provides, new handlers and middlewares have to be written.

## Inspiration and credits
Much of inspiration for this library was taken from [Gin framework](github.com/gin-gonic/gin). I think that Gin is great. However, it is full blown framework, which is not intention of this library. Also, [Negroni](https://github.com/codegangsta/negroni) middleware management library had influence on designing Mezvaro. 

## Router
Mezvaro does not bind itself to any router. It has been designed like that from the start and it is hardly going to change. Core library uses only dependencies from standard library (with `net/context` as addition). However, it is possible to use mezvaro with any router that respects [http.Handler](https://godoc.org/net/http#Handler) interface. In following days/weeks I will publish spearate projects for couple of most popular router libraries that will provide tighter integration with Mezvaro. For now, I am working on support for [Gorilla Mux](https://github.com/gorilla/mux) and [HttpRouter](https://github.com/julienschmidt/httprouter) support. 

## Example
More documentation is under way. However, for first glimpse, here are few short examples.

#### Hello world
```go
package hello_world

import (
	"net/http"

	mv "github.com/delicb/mezvaro"
)

func main() {
	m := mv.New()
	m.UseFunc(func(c *mv.Context) {
		c.Response.Write([]byte("Hello world."))
	})
	http.Handle("/", m)
	http.ListenAndServe(":8000", nil)
}
```

#### Middleware
```go
package middleware

import (
	"net/http"

	mv "github.com/delicb/mezvaro"
)

func HeaderMiddleware(c *mv.Context) {
	c.Response.Header().Set("Server", "Golang-Mezvaro")
	c.Next()
}

func HelloWorldHandler(c *mv.Context) {
	c.Response.Write([]byte("Response"))
}

func main() {
	m := mv.New()
	m.UseFunc(HeaderMiddleware, HelloWorldHandler)
	http.Handle("/", m)
	http.ListenAndServe(":8000", nil)
}
```

#### http.Handler middleware and handler
This example is identical to previous one, but it uses [http.Handler](https://godoc.org/net/http#Handler) interface and middleware to do same thing.
```go

import (
	"net/http"

	mv "github.com/delicb/mezvaro"
)

func HeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Golang-Mezvaro")
		next.ServeHTTP(w, r)
	})
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Response"))
}

func main() {
	m := mv.New()
	m.UseHandlerMiddleware(HeaderMiddleware)
	m.UseHandlerFunc(HelloWorldHandler)
	http.Handle("/", m)
	http.ListenAndServe(":8000", nil)
}
```

#### Context usage
This is simple example of usage of context to pass information between middlewares and handler. As mentioned earlier, Context implements `net/context` and can be used for setting timeout, deadline, getting cancel function or passing values (as shown in this example).
```go
import (
	"fmt"
	"net/http"

	mv "github.com/delicb/mezvaro"
	"github.com/mssola/user_agent"
)

type BrowserInfo struct {
	Name    string
	Version string
	OS      string
}

type userInfo int

const browserInfoKey userInfo = 1

func UserAgentMiddleware(c *mv.Context) {
	ua := user_agent.New(c.Request.Header["User-Agent"][0])
	name, version := ua.Browser()
	browserInfo := BrowserInfo{
		Name:    name,
		Version: version,
		OS:      ua.OS(),
	}
	c.WithValue(browserInfoKey, browserInfo)
	c.Next()
}

func HelloWorldHandler(c *mv.Context) {
	browserInfo := c.Value(browserInfoKey).(BrowserInfo)
	msg := fmt.Sprintf(
		"You are accessing with browser: %s in version %s from OS: %s",
		browserInfo.Name, browserInfo.Version, browserInfo.OS,
	)
	c.Response.Write([]byte(msg))
}

func main() {
	m := mv.New()
	m.UseFunc(UserAgentMiddleware, HelloWorldHandler)
	http.Handle("/", m)
	http.ListenAndServe(":8000", nil)
}

```

## Plans
This is under heavy development, so breaking changes are possible. However, feel free to report tickets and send pull requests if you find this aproach interesting. Any feedback is welcome.

Next steps are to write more tests and improve quality of existing ones (100% coverage is goal in next couple of weeks). Also, in the same period, bindings for few routers will be released (and linked here).

## Installation
To install Mezvaro, run `go get github.com/delicb/mezvaro` from command line.
