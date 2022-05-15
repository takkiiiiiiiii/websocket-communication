package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	//socketはwebクライアントのためのWebsocket
	socket *websocket.Conn
	//sendはメッセージが送られるチャネル
	send chan []byte
	//roomはこのクライアントが参加しているチャットルーム
	room *room
}

//WriteMessage and ReadMessage methods to send and receive messages as a slice of bytes

func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg //チャネルへ値を 送信
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
