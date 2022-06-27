package main

import (
	"time"
)

//messageは1つのメッセージを表す
type message struct {
	Name    string    //ユーザー名
	Message string    //contents
	When    time.Time //送信された時刻
}
