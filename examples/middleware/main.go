package main
import (
	"net/http"
	mv "github.com/delicb/mezvaro"
)

func Middleware1(c *mv.Context) {
	c.Response.Write([]byte("Added in middleware 1 before calling Next\n"))
	c.Next()
	c.Response.Write([]byte("Added in middleware 1 after calling Next\n"))
}

func Middleware2(c *mv.Context) {
	c.Response.Write([]byte("Added in middleware 2 before calling Next\n"))
	c.Next()
	c.Response.Write([]byte("Added in middleware 2 after calling Next\n"))
}

func HelloHandler(c *mv.Context) {
	c.Response.Write([]byte("Hello\n"))
}

func WorldHandler(c *mv.Context) {
	c.Response.Write([]byte("World\n"))
}

func main() {
	m := mv.New()
	m.UseFunc(Middleware1, Middleware2)
	http.Handle("/hello", m.HF(HelloHandler))
	http.Handle("/workd", m.HF(WorldHandler))
	http.ListenAndServe(":8000", nil)
}
