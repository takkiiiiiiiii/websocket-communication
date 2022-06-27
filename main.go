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

//ServeHttpはHTTPリクエストを処理します
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() { //sync package の once.Doは一度飲み実行される
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename))) //filepath.Join パスを結合する
	})
	data := map[string]interface{}{ //Host, UserDataの2つのフィールドが含まれる 　連想配列
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil { //r.Cookie https://pkg.go.dev/net/http#Request.Cookie
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	// テンプレートを描画
	t.templ.Execute(w, data)
}

//http.HandlerFuncを使えばServeHTTPを実装するstructを作らなくて良い

func main() {
	var addr = flag.String("addr", "8080", "アプリケーションのアドレス") //flag.String(flagname, default value, usage string) 戻り値 *sting
	flag.Parse()
	//Gomniauthのセットアップ
	gomniauth.SetSecurityKey("セキュリティーキー")
	gomniauth.WithProviders(
		facebook.New("クライアントID", "秘密の鍵", "http://localhost:8080/auth/callback/facebook"),
		google.New("749056294155-ihl05dq8rps4r8u8gv75sb6jjoncdspi.apps.googleusercontent.com", "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", "http://localhost:8080/auth/callback/google"), //http://localhost:8080/auth/callback/googleにアクセスするとgoogleによるログイン処理の画面にリダイレクトされる
		github.New("クライアントID", "秘密の鍵", "http://localhost:8080/auth/callback/github"),
	)
	r := newRoom()                                                          //チャットルーム作成用の関数 戻り値 *room
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"})) //func Handle(pattern string, handler Handler) handlerは独自ハンドラ(interface or 構造体のポインタ)
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler) //loginHandlerは内部状態を保持する
	//MustAuth関数を使って、*templateHandlerをラップ -> http.Handlerをラップした*authHandlerを生成
	//まず、*authHandlerのServeHTTPメソッドが実行されて、認証に成功した時のみ*templateHandlerのServeHTTPメソッドが実行される

	http.Handle("/room", r)
	//チャットルームを開始
	go r.run() //goroutineとして実行される
	//webサーバー起動
	log.Println("webサーバー開始 port : ", *addr) //*addr 間接演算子
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("LintenAndServe: ", err)
	}
}
