package webserver

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsCtxMap = new(sync.Map)
)

const (
	defaultWriteWait       = 10 * time.Second
	defaultPongWait        = 5 * time.Second
	defaultPingPeriod      = (defaultPongWait * 9) / 10
	defaultMaxMessageSize  = 2048
	defaultDeadInNoPongNum = 10
)

type WebSocketCallback interface {
	OnReciveTextMessage(string, []byte)
	OnReciveBinaryMessage(string, []byte)
	OnConnection(string, *WSRequest) bool
	OnCloseConnection(string)
	OnError(error)
}

type WebSocketController struct {
	WebSocketHandler
}

func (w *WebSocketController) SendTextMessage(to string, msg []byte) error {

	return w.WebSocketHandler.sendMessage(to, websocket.TextMessage, msg)

}

func (w *WebSocketController) SendBinaryMessage(to string, msg []byte) error {
	return w.WebSocketHandler.sendMessage(to, websocket.BinaryMessage, msg)
}

func (w *WebSocketController) Close(id string) error {

	return w.WebSocketHandler.closeConnection(id)

}

type WebSocketHandler struct {
	upgrader        websocket.Upgrader
	connections     *sync.Map
	controller      *WebSocketController
	callback        WebSocketCallback
	pongWait        time.Duration
	pingPeriod      time.Duration
	writeWait       time.Duration
	deadInNoPongNum int
	maxMessageSize  int64
}

func NewWebSocketHandler(wsCfg WebSocketConfig, callback WebSocketCallback) *WebSocketHandler {
	r := WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:    wsCfg.ReadBufferSize,
			WriteBufferSize:   wsCfg.WriteBufferSize,
			Subprotocols:      wsCfg.Subprotocols,
			EnableCompression: wsCfg.EnableCompression,
		},

		connections: new(sync.Map),
		controller:  new(WebSocketController),
		callback:    callback,

		pongWait:        defaultPongWait,
		pingPeriod:      defaultPingPeriod,
		writeWait:       defaultWriteWait,
		maxMessageSize:  defaultMaxMessageSize,
		deadInNoPongNum: defaultDeadInNoPongNum,
	}
	r.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	r.controller.WebSocketHandler = r
	if wsCfg.HandshakeTimeout != 0 {
		r.upgrader.HandshakeTimeout = time.Second * time.Duration(wsCfg.HandshakeTimeout)
	}

	if wsCfg.PongWait > 0 {
		r.pongWait = time.Second * time.Duration(wsCfg.PongWait)
		r.pingPeriod = (r.pongWait * 9) / 10
	}

	if wsCfg.WriteWait > 0 {
		r.writeWait = time.Second * time.Duration(wsCfg.WriteWait)
	}

	if wsCfg.MaxMessageSize > 0 {
		r.maxMessageSize = wsCfg.MaxMessageSize
	}

	if wsCfg.DeadInNoPongNum > 0 {
		r.deadInNoPongNum = wsCfg.DeadInNoPongNum
	}

	return &r
}

func (ws *WebSocketHandler) HandleConn(w http.ResponseWriter, r *http.Request) {

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println("Failed to set websocket upgrade: " + err.Error())
		return
	}
	err = ws.onConnection(conn, r)
	if err != nil {
		logger.Println("Open connection failed " + err.Error() + r.URL.RawQuery)
		conn.Close()
		return
	}

}

func (ws *WebSocketHandler) onCloseConnection(id string) {
	ws.callback.OnCloseConnection(id)
	ws.connections.Delete(id)
}

func (ws *WebSocketHandler) closeConnection(id string) (err error) {

	c, ok := ws.connections.Load(id)

	if !ok {
		return errors.New("connection not found")
	}

	c.(*Client).close()

	return
}

type WSRequest struct {
	request *http.Request
}

func (w *WSRequest) FormValue(key string) string {

	return w.request.FormValue(key)
}

func (w *WSRequest) Header(key string) string {
	return w.request.Header.Get(key)
}

func (ws *WebSocketHandler) onConnection(conn *websocket.Conn, r *http.Request) (err error) {

	w := md5.New()
	io.WriteString(w, fmt.Sprintf("%s@%v", conn.RemoteAddr().String(), time.Now().UnixNano()))
	id := fmt.Sprintf("%x", w.Sum(nil))

	if !ws.callback.OnConnection(id, &WSRequest{r}) {
		err = errors.New("not allow")
		return
	}

	client := &Client{
		id:       id,
		conn:     conn,
		sendChan: make(chan int64),
		handler:  ws,
	}
	client.deal()

	ws.connections.Store(id, client)

	return
}

func (wsh *WebSocketHandler) onError(errMsg string) {
	logger.Println(errMsg)
}

func (ws *WebSocketHandler) sendMessage(id string, msgType int, msg []byte) error {

	c, ok := ws.connections.Load(id)

	if !ok {
		return errors.New("connection not found")
	}

	client := c.(*Client)

	client.sendMsg(msgType, msg)

	return nil

}

func (ws *WebSocketHandler) onReciveTextMessage(id string, msg []byte) {
	ws.callback.OnReciveTextMessage(id, msg)
}

func (ws *WebSocketHandler) onReciveBinaryMessage(id string, msg []byte) {
	ws.callback.OnReciveBinaryMessage(id, msg)
}

type Client struct {
	id          string
	conn        *websocket.Conn
	sendChan    chan int64
	handler     *WebSocketHandler
	closeMutex  sync.RWMutex
	isClosed    bool
	sendMsgMap  sync.Map
	sendID      int64
	noPongCount int32
}

type MessageWrapper struct {
	id      int64
	msgType int
	msg     []byte
	done    chan error
}

func (c *Client) nextSendID() int64 {

	return atomic.AddInt64(&c.sendID, 1)
}

func (c *Client) deal() {
	c.conn.SetReadLimit(c.handler.maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.handler.pongWait))
	c.conn.SetPongHandler(func(msg string) error {
		atomic.StoreInt32(&c.noPongCount, 0)
		c.conn.SetReadDeadline(time.Now().Add(c.handler.pongWait))
		return nil
	})
	c.conn.SetPingHandler(func(msg string) error {
		err := c.conn.WriteControl(websocket.PongMessage, []byte(msg), time.Now().Add(c.handler.writeWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err

	})
	go c.readPump()
	go c.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.close()
	}()

	for {
		t, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Printf("websocket error: %v", err)
			}
			break
		}
		atomic.StoreInt32(&c.noPongCount, 0)
		switch t {
		case websocket.TextMessage:
			c.handler.onReciveTextMessage(c.id, message)
		case websocket.BinaryMessage:
			c.handler.onReciveBinaryMessage(c.id, message)

		}
	}
}

func (c *Client) sendMsg(msgType int, msg []byte) error {
	c.closeMutex.RLock()
	defer c.closeMutex.RUnlock()
	if c.isClosed {
		return errors.New(fmt.Sprintf("[client.sendMsg]client %v closed", c.id))
	}
	done := make(chan error)

	mw := MessageWrapper{id: c.nextSendID(), msgType: msgType, msg: msg, done: done}
	c.sendMsgMap.Store(mw.id, &mw)
	defer func() {
		c.sendMsgMap.Delete(mw.id)
		close(mw.done)
	}()
	c.sendChan <- mw.id

	return <-done
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.handler.pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case id, ok := <-c.sendChan:

			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(c.handler.writeWait))

			msg, ok := c.sendMsgMap.Load(id)

			if !ok {
				logger.Println("get message failed : ", id)
				continue
			}
			mw := msg.(*MessageWrapper)

			err := c.conn.WriteMessage(mw.msgType, mw.msg)

			mw.done <- err
			if err != nil {
				return
			}

		case <-ticker.C:

			if int(atomic.LoadInt32(&c.noPongCount)) > c.handler.deadInNoPongNum {
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(c.handler.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			atomic.AddInt32(&c.noPongCount, 1)
		}
	}

}

func (c *Client) close() (err error) {
	c.closeMutex.Lock()
	if c.isClosed {
		c.closeMutex.Unlock()
		return
	} else {
		c.isClosed = true
		c.closeMutex.Unlock()
	}

	defer func() {
		c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		err = c.conn.Close()
	}()
	close(c.sendChan)
	c.handler.onCloseConnection(c.id)

	return
}
