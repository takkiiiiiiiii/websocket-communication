package main

import (
	"chat/trace"
	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
	"log"
	"net/http"
)

type room struct {
	//forwardは他のクライアントに転送するためのメッセージを保持するためのチャネル
	forward chan *message
	//joinはチャットルームに参加しようとしているクライアントのためのチャネル
	join chan *client
	//leaveはチャットルームから退室しようとしているクライアントのためのチャネル
	leave chan *client
	//clientsに在室しているすべてのクライアントが保持されている
	clients map[*client]bool
	//tracerはチャットルーム上で行われた操作のログを受け取ります
	tracer trace.Tracer //traceパッケージのTrace型(interface)
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join: //<-channel 構文  チャネルから値を受信
			//参加
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")
		case client := <-r.leave:
			//退室
			delete(r.clients, client) //map rooo型のclientsからclientを削除
			close(client.send)
			r.tracer.Trace("クライアントは退室しました")
		case msg := <-r.forward:
			r.tracer.Trace("メッセージを受信しました: ", msg.Message)
			//すべてのクライアントにメッセージを送信
			for client := range r.clients {
				select {
				case client.send <- msg:
					//メッセージを送信
					r.tracer.Trace(" -- クライアントに送信しました")
				default:
					//送信に失敗
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- 送信失敗しました。クライアントをクリーンアップします。")
				}
			}
		}
	}
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(), //niltrace構造体とともに定義   trace.Off() 戻り値 *niltrace  newRoom生成したら　traceパッケージのOffメソッドも実行される1
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil) //UpgraderのUpgradeはHTTP通信からWebSocket通信に更新してくれる
	if err != nil {
		log.Fatal("ServeHttp:", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("クッキーの取得に失敗しました: ", err)
		return
	}
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value), //MustFromBase64の戻り値 map[string]interface{}  エンコードされたクッキーの値をマップのオブジェクトへ復元
	}

	r.join <- client //チャネルへ値を 送信
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
