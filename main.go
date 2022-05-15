package main

import (
	"flag"
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
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename))) //filepath.Join パスを結合する
	})
	// テンプレートを描画
	t.templ.Execute(w, r)
}

//http.HandlerFuncを使えばServeHTTPを実装するstructを作らなくて良い

func main() {
	var addr = flag.String("addr", "8080", "アプリケーションのアドレス") //flag.String(flagname, default value, usage string) 戻り値 *sting
	flag.Parse()
	r := newRoom() //チャットルーム作成用の関数 戻り値 *room
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	//チャットルームを開始
	go r.run() //goroutineとして実行される
	//webサーバー起動
	log.Println("webサーバー開始 port : ", *addr) //*addr 間接演算子
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("LintenAndServe: ", err)
	}
}
