// Package mezvaro is simple implementation for middleware management for Go services.
// Mezvaro does not follow http.Handler interface conventions (although it is
// compatible with http.Handler). Instead Mezvaro handlers receive instance
// context that can be used to obtain ResponseWriter and *Request that is
// usually provided to http.Handler. ResponseWriter and *Request can be obtained
// from context provided, of course, but context allows communication between
// middlewares.
//
// Also, context that Mezvaro uses is fully compatible with x/net/context
// library.
//
// All existing Middleware handlers can be used with Mezvaro without hassle,
// but those middlewares can not use Context, since they only have access
// to ResponseWriter and *Request objects.
//
// Example of hello world handler:
//
//     func MyHandler(c *mezvaro.Context) {
//         c.Response.Write([]byte("hello world"))
//     }
package mezvaro
