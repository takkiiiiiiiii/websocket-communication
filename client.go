package main

import (
	"github.com/gorilla/websocket"
	"time"
)

type client struct {
	//socketはwebクライアントのためのWebsocket  //WebSocketとは、WebサーバとWebブラウザの間で双方向通信できるようにする技術
	socket *websocket.Conn
	//sendはメッセージが送られるチャネル
	send chan *message
	//roomはこのクライアントが参加しているチャットルーム
	room *room
	//userdata  ユーザーに関するデータを保持
	userData map[string]interface{}
}

//WriteMessage and ReadMessage methods to send and receive messages as a slice of bytes

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil { //ReadJSON func(v interface{}) error   message.goのmessage型をデコード
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg //チャネルへ値を 送信
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
