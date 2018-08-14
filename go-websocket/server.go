package main

import (
	"net/http"
	"PushMsgServ_WebSocket/websocket"
	"PushMsgServ_WebSocket/go-websocket/impl"
	"time"
)


var(
	upgrader = websocket.Upgrader{
		//允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter,r *http.Request){
	var(
		wsConn *websocket.Conn
		err error
		data []byte
		conn *impl.Connection
	)
	//Upgrade:websocket
	if wsConn,err = upgrader.Upgrade(w,r,nil);err != nil{
		return
	}

	if conn,err = impl.InitConnection(wsConn);err != nil{
		goto ERR
	}

	go func() {
		var(
			err error
		)
		for{
				if err = conn.WriteMessage([]byte("heartbeat"));err != nil{
					return
				}
				time.Sleep(1 * time.Second)
		}
	}()

	for{
		if data,err = conn.ReadMessage();err != nil{
			goto ERR
		}
		if err = conn.WriteMessage(data);err != nil{
			goto ERR
		}
	}

ERR:
	conn.Close()


	//websocket.Conn
	/*for{
		//Text,Binary
		if _,data,err = conn.ReadMessage();err != nil{
			goto ERR
		}
		if err = conn.WriteMessage(websocket.TextMessage,data);err != nil{
			goto ERR
		}
	}
	ERR:
		conn.Close()*/
}

func main() {
	//  接口地址为 http://localhost:7777/ws
	http.HandleFunc("/ws",wsHandler)  //配置路由

	http.ListenAndServe("0.0.0.0:7777",nil)
}
