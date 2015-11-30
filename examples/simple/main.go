package main

import (
	"net/http"

	mv "github.com/delicb/mezvaro"
)

func HelloHandler(c *mv.Context) {
	c.Response.Write([]byte("Hello"))
}

func main() {
	var m mv.Mezvaro
	http.Handle("/", m.HF(HelloHandler))
	http.ListenAndServe(":8000", nil)
}
