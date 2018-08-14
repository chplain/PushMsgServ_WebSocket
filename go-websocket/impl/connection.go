package impl

import (
	"PushMsgServ_WebSocket/websocket"
	"sync"
	"github.com/pkg/errors"
)

type Connection struct {
	wsConn *websocket.Conn
	inChan chan []byte
	outChan chan []byte
	closeChan chan byte

	mutex sync.Mutex  //加锁
	isClosed bool

}

func InitConnection(wsConn *websocket.Conn)(conn *Connection,err error){  //把wensocket长链接做一个封装
	conn = &Connection{
		wsConn:wsConn,
		inChan:make(chan []byte,1000),
		outChan:make(chan []byte,1000),
		closeChan:make(chan byte,1),
	}

	//启动读协程
	go conn.readLoop()

	//启动写协程
	go conn.writeLoop()

	return
}

//封装API
func (conn *Connection)ReadMessage()(data []byte,err error){
	select {
		case data = <- conn.inChan:
		case <- conn.closeChan:
			err = errors.New("connection is closed")
	}
	return
}

func (conn *Connection)WriteMessage(data []byte)(err error){
	select {
	case conn.outChan <- data:
	case <- conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func (conn *Connection)Close(){
	//线程安全，可重入的CLose
	conn.wsConn.Close()

	//这一行代码只执行一次
	conn.mutex.Lock()
	if !conn.isClosed{  //判断close标记_如果没有被关闭过
		close(conn.closeChan)
		conn.isClosed = true  //已经关闭
	}
	conn.mutex.Unlock()  //释放掉这把锁

}

//内部实现
func (conn *Connection)readLoop(){
	var(
		data []byte
		err error
	)
	//读长链接的消息
	for{
		if _,data,err = conn.wsConn.ReadMessage();err != nil{
			goto ERR
		}
		//阻塞在这里，等待inChan有空闲的位置
		select {
			case conn.inChan <- data:
			case <- conn.closeChan:
				//closeChan关闭的时候
				goto ERR
		}

	}
	ERR:
		conn.Close()
}

func (conn *Connection)writeLoop(){
	var(
		data []byte
		err error
	)
	//如果消息发送成功，继续取消息发送
	for{
		select {
		case data = <- conn.outChan:
		case <- conn.closeChan:
			goto ERR

		}

		if err = conn.wsConn.WriteMessage(websocket.TextMessage,data);err != nil{
			goto ERR
		}
	}
	ERR:
		conn.Close()
}