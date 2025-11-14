package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"chatapp.myatty.net/trace"
	"github.com/joho/godotenv"
	"github.com/stretchr/objx"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

var GoogleOAuthConfig *oauth2.Config
var OAuthStateString = "sixSevensixOne"

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ =
			template.Must(template.ParseFiles(filepath.Join("templates",
				t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)

}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading the .env file!")
	}

	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/callback/google",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	OAuthStateString = "sixSevensixOne"

	fmt.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse()

	r := newRoom()

	// Comment this to remove tracing
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.tmpl.html"}))
	http.Handle("/login", &templateHandler{filename: "login.tmpl.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Error encountered: ", err)
	}
}
