package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type room struct {
	//forwardは他のクライアントに転送するためのメッセージを保持するためのチャネル
	forward chan []byte
	//joinはチャットルームに参加しようとしているクライアントのためのチャネル
	join chan *client
	//leaveはチャットルームから退室しようとしているクライアントのためのチャネル
	leave chan *client
	//clientsに在室しているすべてのクライアントが保持されている
	clients map[*client]bool
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join: //<-channel 構文  チャネルから値を受信
			//参加
			r.clients[client] = true
		case client := <-r.leave:
			//退室
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			//すべてのクライアントにメッセージを送信
			for client := range r.clients {
				select {
				case client.send <- msg:
					//メッセージを送信
				default:
					//送信に失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
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
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client //チャネルへ値を 送信
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
