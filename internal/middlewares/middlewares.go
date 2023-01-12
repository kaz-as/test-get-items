package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r != nil {
				log.Printf("new request: %v \"%v\"", r.Method, r.RequestURI)
			} else {
				log.Printf("logger: request is nil")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Recoverer() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic: %v", rec)
					jsonBody, _ := json.Marshal(map[string]string{
						"error": "There was an internal server error",
					})

					if w.Header() != nil {
						w.Header().Set("Content-Type", "application/json")
					}
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write(jsonBody)
					if err != nil {
						log.Printf("cannot write internal server error: %s", err)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
