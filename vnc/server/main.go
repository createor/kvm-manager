package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	guac "vnc-server/guac"

	"context"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

// 定义请求参数
type ReqArg struct {
	GuacadAddr    string `from:"guacad_addr"`    // 访问地址
	AssetProtocol string `from:"asset_protocol"` // 访问协议
	AssetHost     string `from:"asset_host"`     // win主机地址
	AssetPort     string `from:"asset_port"`     // win端口
	AssetUser     string `from:"asset_user"`     // win用户
	AssetPassword string `from:"asset_password"` // win密码
	ScreenWidth   int    `from:"screen_width"`   // 屏幕宽度
	ScreenHeight  int    `from:"screen_height"`  // 屏幕高度
	ScreenDpi     int    `from:"screen_dpi"`     // 屏幕深度
}

// ws协议转guacamole协议
func Ws2Guacamole() func(w http.ResponseWriter, r *http.Request) { // 返回http的handler句柄事件
	wsReadBufferSize := guac.MaxGuacMessage      // ws读取的buffer大小
	wsWriteBufferSize := guac.MaxGuacMessage * 2 // ws写入的buffer大小
	// 初始化ws设置
	upgrader := websocket.Upgrader{
		ReadBufferSize:  wsReadBufferSize,
		WriteBufferSize: wsWriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// 限定websocket被其他的域名访问
			return true
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// 1.接收连接参数并解析
		r.ParseForm() // 解析url参数
		width, _ := strconv.Atoi(r.FormValue("width"))
		height, _ := strconv.Atoi(r.FormValue("height"))
		dpi, _ := strconv.Atoi(r.FormValue("dpi"))
		arg := ReqArg{
			GuacadAddr:    r.FormValue("addr"),
			AssetProtocol: r.FormValue("protocol"),
			AssetHost:     r.FormValue("host"),
			AssetPort:     r.FormValue("port"),
			AssetUser:     r.FormValue("user"),
			AssetPassword: r.FormValue("password"),
			ScreenWidth:   width,
			ScreenHeight:  height,
			ScreenDpi:     dpi,
		}
		// 2.创建ws连接
		head := http.Header{} // 设置返回头部信息
		head.Add("Sec-Websocket-Protocol", r.Header.Get("Sec-Websocket-Protocol"))
		conn, err := upgrader.Upgrade(w, r, head) // 创建ws连接
		if err != nil {
			fmt.Println("创建ws失败")
		}
		defer func() { // 关闭ws连接
			if err = conn.Close(); err != nil {
				fmt.Println("关闭ws失败")
			}
		}()
		// 3.ws转guacamole
		uuid := ""
		pipTunnel, err := guac.NewGuacamoleTunnel(arg.GuacadAddr, arg.AssetProtocol, arg.AssetHost, arg.AssetPort, arg.AssetUser, arg.AssetPassword, uuid, arg.ScreenWidth, arg.ScreenHeight, arg.ScreenDpi) // 创建连接管道
		if err != nil {
			fmt.Println("创建rdp连接失败")
		}
		defer func() { // 关闭rdp连接
			if err = pipTunnel.Close(); err != nil {
				fmt.Println("关闭rdp连接失败")
			}
		}()
		ioCopy(conn, pipTunnel)
	}
}

// ws转guacamole实际处理逻辑
func ioCopy(ws *websocket.Conn, tunl *guac.SimpleTunnel) {
	// 输出
	writer := tunl.AcquireWriter()
	defer tunl.ReleaseWriter() // 释放
	// 输入
	reader := tunl.AcquireReader()
	defer tunl.ReleaseReader() // 释放
	// 使用sync的errgroup处理goroutine协程
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error { // 写数据
		// 设置缓存buffer
		buffer := bytes.NewBuffer(make([]byte, 0, guac.MaxGuacMessage*2))
		for {
			inStream, err := reader.ReadSome()
			if err != nil {
				fmt.Println(err)
				return err
			}
			if bytes.HasPrefix(inStream, guac.InternalOpcodeIns) { // 判断前几个字节
				continue
			}
			if _, err = buffer.Write(inStream); err != nil { // 将读取到的数据写入buffer
				fmt.Println(err)
				return err
			}
			if !reader.Available() || buffer.Len() >= guac.MaxGuacMessage { // 当读取到数据或者buffer里数据大于设置值时，发送数据并清空buffer
				if err = ws.WriteMessage(1, buffer.Bytes()); err != nil {
					// 判断错误类型
					fmt.Println(err)
					if err == websocket.ErrCloseSent {
						return fmt.Errorf("websocket:%v", err)
					}
					return err
				}
				buffer.Reset() // 重置
			}
		}
	})
	eg.Go(func() error { // 读数据
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return err
			}
			if bytes.HasPrefix(data, guac.InternalOpcodeIns) {
				continue
			}
			if _, err = writer.Write(data); err != nil {
				fmt.Println(err)
				return err
			}
		}
	})
	if err := eg.Wait(); err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	http.HandleFunc("/webssh", Ws2Guacamole())
	http.ListenAndServe(":6080", nil)
}
