package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

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
)

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	fmt.Printf("[Upgrade] uprgade websocket ...\n")
	return wsUpgrader.Upgrade(w, r, nil)
}

type Message struct {
	Type    int
	Content []byte
}

type ClientManager struct {
	addr        string
	Conn        *websocket.Conn
	sendMsgChan chan Message
	recvMsgChan chan Message
}

func (c *ClientManager) broadcast() {
	fmt.Printf("[broadcast] broadcast message to %s ...\n", c.addr)
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

func (c *ClientManager) SendMsg(mType int, msg []byte) {
	fmt.Printf("[SendMsg] send message to %s messageType %v msg %v...\n", c.addr, mType, string(msg))
	switch mType {
	case websocket.TextMessage:
		c.sendText(msg)
	case websocket.PingMessage:
		c.sendPing([]byte("ping"))
	case websocket.PongMessage:
		c.sendPong([]byte("pong"))
	case websocket.BinaryMessage:
		c.sendBinary(msg)
	case websocket.CloseMessage:
		c.sendClose("bye")
		return
	default:
		fmt.Println("未知消息类型")
	}
	c.broadcast()
}

func (c *ClientManager) Close() error {
	fmt.Printf("[Close] close websocket %s ...\n", c.addr)
	c.SendMsg(websocket.CloseMessage, nil)
	return c.Conn.Close()
}

func (c *ClientManager) sendText(msg []byte) {
	fmt.Printf("[sendText] send text message to %s ...\n", c.addr)
	c.sendMsgChan <- Message{Type: websocket.TextMessage, Content: msg}
}

func (c *ClientManager) sendPing(msg []byte) {
	fmt.Printf("[sendPing] send ping message to %s ...\n", c.addr)
	c.sendMsgChan <- Message{Type: websocket.PingMessage, Content: msg}
}

func (c *ClientManager) sendPong(msg []byte) {
	fmt.Printf("[sendPong] send pong message to %s ...\n", c.addr)
	c.sendMsgChan <- Message{Type: websocket.PongMessage, Content: msg}
}

func (c *ClientManager) sendBinary(msg []byte) {
	fmt.Printf("[sendBinary] send binary message to %s ...\n", c.addr)
	c.sendMsgChan <- Message{Type: websocket.BinaryMessage, Content: msg}
}

func (c *ClientManager) sendClose(msg string) {
	fmt.Printf("[sendClose] send close message to %s ...\n", c.addr)
	deadline := time.Now().Add(time.Minute)
	err := c.Conn.WriteControl(
		websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg),
		deadline,
	)
	if err != nil {
		fmt.Printf("[sendClose] send close message to %s error: %v\n", c.addr, err)
	}
	err = c.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		fmt.Printf("[sendClose] set read deadline to %s error: %v\n", c.addr, err)
	}
	for {
		_, _, err = c.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			fmt.Printf("[sendClose] close by mornal ...\n")
			break
		}
		if err != nil {
			fmt.Printf("[sendClose] read message error: %v\n", err)
			break
		}
	}
}

func NewClientManager(addr string, conn *websocket.Conn) *ClientManager {
	return &ClientManager{
		addr:        addr,
		Conn:        conn,
		sendMsgChan: make(chan Message, 1024),
		recvMsgChan: make(chan Message, 1024),
	}
}

func SetClient(client *ClientManager) bool {
	if clients == nil {
		clients = make(map[string]*ClientManager)
	}
	_, ok := clients[client.addr]
	if ok {
		return false
	}
	clients[client.addr] = client
	return true
}

func GetClient(addr string) *ClientManager {
	if clients == nil {
		return nil
	}
	if client, ok := clients[addr]; ok {
		return client
	}
	return nil
}

func CloseClient(addr string) {
	if client, ok := clients[addr]; ok {
		_ = client.Close()
		delete(clients, addr)
		fmt.Printf("bye %s\n", addr)
	}
}

func BroadcastMsg(mType int, msg []byte) {
	for _, client := range clients {
		client.SendMsg(mType, msg)
	}
}

func CloseAll() {
	for _, client := range clients {
		_ = client.Close()
	}
}
