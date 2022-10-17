package utils

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func NewWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, _ := websocket.Upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	go func(conn *websocket.Conn) {
		for{
			// 读取信息
			msgType,msg,_ := conn.ReadMessage()
			// 发送信息
			conn.WriteMessage(msgType,msg)
		}
	}(conn)
}

http.HandleFunc("/webssh",NewWebsocket)
http.ListenAndServe(":6080",nil)