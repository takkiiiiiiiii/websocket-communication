package main

import (
	"flag"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

//独自のハンドラを作成

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
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
	var addr = flag.String("addr", "8080", "アプリケーションのアドレス")
	flag.Parse()

	gomniauth.SetSecurityKey("セキュリティーキー")
	gomniauth.WithProviders(
		facebook.New("クライアントID", "秘密の鍵", "http://localhost:8080/auth/callback/facebook"),
		google.New("クライアントID", "http://localhost:8080/auth/callback/google"),
		github.New("クライアントID", "秘密の鍵", "http://localhost:8080/auth/callback/github"),
	)
	r := newRoom()
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)

	http.Handle("/room", r)

	go r.run()

	log.Println("webサーバー開始 port : ", *addr)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("LintenAndServe: ", err)
	}
}
