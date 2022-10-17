package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
)

func NewWebsocket(w http.ResponseWriter, r *http.Request) {
	// upgrader := websocket.Upgrader{}
	// conn, _ := upgrader.Upgrade(w, r, nil)
	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	cmdline := make([]string, 0)
	// defer conn.Close()
	go func(conn *websocket.Conn, cmdline []string) {
		for {
			// 读取信息
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			// fmt.Println(string(msg))
			if string(msg) == "\r" { // 回车事件
				sendCmd := strings.Join(cmdline, "")
				if len(strings.TrimSpace(sendCmd)) != 0 { // 命令去除空格后不等于0
					cmd := exec.Command("bash", "-c", sendCmd)
					result, err := cmd.CombinedOutput()
					// fmt.Println(string(result))
					if err != nil {
						fmt.Println(err)
					}
					// 发送信息
					err = conn.WriteMessage(msgType, result)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				cmdline = cmdline[0:0] // 清空切片
			} else if string(msg) == "\177" { // 删除事件
				if len(cmdline) > 0 {
					cmdline = cmdline[0 : len(cmdline)-1]
				}
			} else { // 输入事件
				cmdline = append(cmdline, string(msg))
			}
		}
	}(conn, cmdline)
}

func main() {
	http.HandleFunc("/webssh", NewWebsocket)
	http.ListenAndServe(":6080", nil)
}
