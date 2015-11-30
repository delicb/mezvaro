package main
import (
	"net/http"
	mv "github.com/delicb/mezvaro"
	"log"
	"math/rand"
	"time"
)

func LoggingMiddleware(c *mv.Context) {
	log.Println("Simluate real logging here")
	c.Next()
}

func AuthMiddleware(c *mv.Context) {
	log.Println("Simulate user authentication.")
	rand.Seed(time.Now().Unix())
	if rand.Int() % 2 == 0 {
		c.Response.Write([]byte("User authenticated.\n"))
		c.Next()
	} else {
		c.Response.WriteHeader(http.StatusUnauthorized)
		c.Response.Write([]byte("Use not authenticated.\n"))
		c.Abort()
	}
}

func PubliclyAvailable(c *mv.Context) {
	c.Response.Write([]byte("This page is available for all users."))
}

func PrivatelyAvailable(c *mv.Context) {
	c.Response.Write([]byte("This page is only for authorized users."))
}

func main() {
	m := mv.New(mv.HandlerFunc(LoggingMiddleware)) // all calls go through logging middleware
	authOnly := m.Fork(mv.HandlerFunc(AuthMiddleware))
	http.Handle("/public", m.HF(PubliclyAvailable))
	http.Handle("/auth", authOnly.HF(PrivatelyAvailable))
	http.ListenAndServe(":8000", nil)
}
