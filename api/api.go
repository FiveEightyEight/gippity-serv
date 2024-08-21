package api

import (
	"log"
	"net/http"

	"github.com/rs/cors"
)

type APIServer struct {
	addr string
}

func CreateAPIServer(addr string) *APIServer {
	return &APIServer{
		addr: addr,
	}
}

type Middleware func(http.Handler) http.HandlerFunc

func HandleMiddleWareChain(middleware ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next.ServeHTTP
	}
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("method %s, path: %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func (s *APIServer) Run() error {

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	})

	router := http.NewServeMux()
	middleWareChain := HandleMiddleWareChain(
		RequestLoggerMiddleware,
	)

	server := http.Server{
		Addr:    s.addr,
		Handler: c.Handler(middleWareChain(router)),
	}
	log.Printf("Server has started %s", s.addr)
	return server.ListenAndServe()
}
