package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stretchr/objx"
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
		switch provider {
		case "google":
			url := GoogleOAuthConfig.AuthCodeURL(OAuthStateString)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			fmt.Println("redirecting : ", url)
		default:
			http.Error(w, "Unsupported provider", http.StatusBadRequest)
		}
	case "callback":
		if provider != "google" {
			http.Error(w, "Unsupported provider", http.StatusBadRequest)
			return
		}

		if r.FormValue("state") != OAuthStateString {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}

		// Exchange code for token with google
		code := r.FormValue("code")
		token, err := GoogleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Code exchange failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch user info
		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Parse JSON
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed reading user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var userInfo map[string]interface{}
		if err := json.Unmarshal(body, &userInfo); err != nil {
			http.Error(w, "Invalid user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create cookie to save name
		name := fmt.Sprintf("%v", userInfo["name"])
		cookieValue := objx.New(map[string]interface{}{
			"name": name,
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: cookieValue,
			Path:  "/",
		})

		http.Redirect(w, r, "/chat", http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
