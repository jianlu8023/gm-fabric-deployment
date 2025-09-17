package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jianlu8023/golang-example/pkg/json"

	"github.com/gorilla/websocket"
)

type MSG interface {
	bytes() []byte
}

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	clients = make(map[string]*ClientManager)

	lock = sync.Mutex{}
)

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return wsUpgrader.Upgrade(w, r, nil)
}

type Message struct {
	Type    int
	Content []byte
}

type Response struct {
	Obj interface{} `json:"obj"`
}

func SuccessResponse() Response {
	return Response{
		Obj: "success",
	}
}

func CustomResponse(obj string) Response {
	return Response{
		Obj: obj,
	}
}

func (r Response) bytes() []byte {
	return []byte(r.String())
}

func (r Response) String() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}

type ClientManager struct {
	addr        string
	Conn        *websocket.Conn
	sendMsgChan chan Message
	sync.Mutex
}

func (c *ClientManager) broadcast() {
	select {
	case msg, ok := <-c.sendMsgChan:
		if !ok {
			return
		}
		_ = c.Conn.WriteMessage(msg.Type, msg.Content)
	default:
		fmt.Print("broadcast default")
	}
}

func (c *ClientManager) SendMsg(mType int, msg MSG) {
	switch mType {
	case websocket.TextMessage:
		c.sendText(msg.bytes())
	case websocket.PingMessage:
		c.sendPing([]byte("ping"))
	case websocket.PongMessage:
		c.sendPong([]byte("pong"))
	case websocket.BinaryMessage:
		c.sendBinary(msg.bytes())
	case websocket.CloseMessage:
		c.sendClose("bye")
		return
	default:
		fmt.Println("未知消息类型")
	}
	c.broadcast()
}

func (c *ClientManager) Close() error {
	c.SendMsg(websocket.CloseMessage, nil)
	return c.Conn.Close()
}

func (c *ClientManager) sendText(msg []byte) {
	c.sendMsgChan <- Message{Type: websocket.TextMessage, Content: msg}
}

func (c *ClientManager) sendPing(msg []byte) {
	c.sendMsgChan <- Message{Type: websocket.PingMessage, Content: msg}
}

func (c *ClientManager) sendPong(msg []byte) {
	c.sendMsgChan <- Message{Type: websocket.PongMessage, Content: msg}
}

func (c *ClientManager) sendBinary(msg []byte) {
	c.sendMsgChan <- Message{Type: websocket.BinaryMessage, Content: msg}
}

func (c *ClientManager) sendClose(msg string) {
	deadline := time.Now().Add(time.Duration(5) * time.Second)
	err := c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg),
		deadline,
	)
	if err != nil {
		fmt.Printf("发送关闭消息失败: %v", err)
	}
	err = c.Conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		fmt.Printf("设置读取超时失败: %v", err)
	}
	for {
		_, _, err = c.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			break
		}
		if err != nil {
			break
		}
	}
}

func NewClientManager(addr string, conn *websocket.Conn) *ClientManager {
	return &ClientManager{
		addr:        addr,
		Conn:        conn,
		sendMsgChan: make(chan Message, 1024),
	}
}

func SetClient(client *ClientManager) bool {
	lock.Lock()
	defer lock.Unlock()
	if clients == nil {
		clients = make(map[string]*ClientManager)
	}
	c, ok := clients[client.addr]
	if ok {
		_ = c.Close()
	}
	// time.Sleep(time.Duration(rand.IntN(7-5)+5) * time.Second)
	clients[client.addr] = client
	// fmt.Printf("clients %v\n", clients)

	return true
}

func GetClient(addr string) *ClientManager {
	lock.Lock()
	defer lock.Unlock()
	if clients == nil {
		return nil
	}
	if client, ok := clients[addr]; ok {
		return client
	}
	return nil
}

func CloseClient(addr string) {
	lock.Lock()
	defer lock.Unlock()
	if client, ok := clients[addr]; ok {
		_ = client.Close()
		delete(clients, addr)
		fmt.Printf("bye %s\n", addr)
	}
}

func BroadcastMsg(mType int, msg MSG) {
	lock.Lock()
	defer lock.Unlock()
	for _, client := range clients {
		client.SendMsg(mType, msg)
	}
}

func BroadcastTextMsg(msg MSG) {
	lock.Lock()
	defer lock.Unlock()
	for _, client := range clients {
		client.SendMsg(websocket.TextMessage, msg)
	}
}

func CloseAll() {
	lock.Lock()
	defer lock.Unlock()
	for _, client := range clients {
		_ = client.Close()
	}
}
