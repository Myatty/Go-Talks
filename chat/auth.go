package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type authHandler struct {
	next http.Handler
}

// When a request comes in, the authHandler checks if the user is authenticated by
// looking for a cookie named "auth"
// If cookie exists, it passes control to the wrapped handler
// (the page the user actually requested)
// Decorator pattern !!!
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		// success, so call the next handler
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")

	if len(segs) < 4 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalid auth format!")
	}

	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":
		log.Println("TODO: Handle login for ", provider)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s is not supported", action)
	}

}
