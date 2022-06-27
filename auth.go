package main

import (
	"fmt"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"strings"
)

type authHandler struct {
	next http.Handler //ラップ対象のハンドラ ここでは、*templateHandler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie { //auth = cookieのチェック
		//未認証   //ログイン画面へリダイレクト
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		panic(err.Error())
	} else {
		//成功 ラップする
		h.next.ServeHTTP(w, r) //*templateHandlerのServeHTTPメソッド
	}
}

//http.Handlerをラップした*authHandlerを生成
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler} //ServeHTTPメソッドを定義することで、*authHandlerはhttp.Handlerに適合できる
}

//loginHandlerはサードパーティーへのログインの処理を受けつける
//パス形式 : /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/") //r.URL.Path = /auth/{action}/{provider}
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider) //urlで指定している認証プロバイダのオブジェクトを指定  func Provider(name string) (common.Provider, error)
		if err != nil {
			log.Fatalln("認証プロバイダのログインに失敗 :", provider, "-", err)
		}
		loginUrl, err := provider.GetBeginAuthURL(nil, nil) //プロバイダ毎の認証ページへのurlを取得
		if err != nil {
			log.Fatalln("GetBeginAuthURLの呼び出し中にエラー発生:", provider, "-", err)
		}
		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider) //
		if err != nil {
			log.Fatalln("認証プロバイダーの取得に失敗しました", provider, "-", err)
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery)) //CompleteAuth 提供されたURLから、認証に必要な情報を抜き出す
		/*
			(provider *GoogleProvider(googleの場合))CompleteAuth func(data Map) (*Credentials, error)  メソッド
			func MustFromURLQuery(query string) Map   指定されたクエリを解析することによって新しいオブジェクトを生成
		*/
		if err != nil {
			log.Fatalln("認証を完了できませんでした", provider, "-", err)
		}

		user, err := provider.GetUser(creds) //プロバイダからコールバック結果（User interface）を取得
		if err != nil {
			log.Fatalln("ユーザーの取得に失敗しました", provider, "-", err)
		}
		authCookieValue := objx.New(map[string]interface{}{ //user interface  https://pkg.go.dev/github.com/stretchr/gomniauth/common#User
			"name": user.Name(),
		}).MustBase64() //含まれているオブジェクトをjsonの文字列表現のBase64に変換
		http.SetCookie(w, &http.Cookie{ //func SetCookie(w ResponseWriter, cookie *Cookie)
			Name:  "auth", //authというクッキーに保存
			Value: authCookieValue,
			Path:  "/"})
		/*
			httpPackageのCookie型 構造体 https://pkg.go.dev/net/http#Cookie
		*/
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound) //ステータスコード404
		fmt.Fprintf(w, "アクション%sには非対応です", action)
	}
}
