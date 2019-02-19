package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var ListeningPort = ":2052"

// var BackendAddr = "ws://localhost:2053/ws"
var BackendAddr = "ws://45.119.213.88:2053/ws"

const (
	BLANK_LINE = "________________________________________________________________________________\n"
	IS_TESTING = false
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Connection struct {
	WsConn        *websocket.Conn
	ConjugateConn *Connection
	ChanWrite     chan []byte
	ChanClose     chan []byte
}

func (c *Connection) ReadPump() {
	defer fmt.Println("ReadPump ended", c.WsConn.LocalAddr(), c.WsConn.RemoteAddr())
	c.WsConn.SetReadLimit(int64(65536))
	c.WsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.WsConn.SetPongHandler(func(string) error {
		c.WsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		messageType, message, err := c.WsConn.ReadMessage()
		_ = messageType //
		if err != nil {
			fmt.Println("WsConn.ReadMessage err", err)
			c.WsConn.Close()
			c.ConjugateConn.Close("ConjugateConn.Close")
			return
		} else { // forward to ConjugateConn
			if IS_TESTING {
				fmt.Printf(
					BLANK_LINE+"ReadPump %v local %v remote %v:\n%v\n",
					time.Now(), c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(),
					string(message))
			}
			c.ConjugateConn.Write(message)
		}
	}
}

func (c *Connection) WritePump() {
	defer fmt.Println("WritePump ended", c.WsConn.LocalAddr(), c.WsConn.RemoteAddr())
	ticker := time.NewTicker(54 * time.Second)
	defer func() { ticker.Stop() }()
	for {
		var writeErr error
		var msg []byte
		var caseName string
		select {
		case msg = <-c.ChanWrite:
			c.WsConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			writeErr = c.WsConn.WriteMessage(websocket.TextMessage, msg)
			if writeErr == nil {
				if IS_TESTING {
					fmt.Printf(BLANK_LINE+"WritePump %v local %v remote %v:\n%v\n",
						time.Now(), c.WsConn.LocalAddr(), c.WsConn.RemoteAddr(),
						string(msg))
				}
			}
			caseName = "0"
		case <-ticker.C:
			c.WsConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			writeErr = c.WsConn.WriteMessage(websocket.PingMessage, nil)
			caseName = "1"
		case msg = <-c.ChanClose:
			c.WsConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
			writeErr = c.WsConn.WriteMessage(websocket.CloseMessage, msg)
			caseName = "2"
		}
		if writeErr != nil {
			c.WsConn.Close()
			c.ConjugateConn.Close("ConjugateConn.Close")
			fmt.Printf("WsConn.WriteMessage writeErr %v msg %v caseName %v\n",
				writeErr, string(msg), caseName)
			return
		}
	}
}

// send close control message
func (c *Connection) Close(reason string) {
	payload := websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason)
	timeout := time.After(1 * time.Second)
	select {
	case c.ChanClose <- payload:
	case <-timeout:
	}
}

// run a goroutine to send the message to peer
func (c *Connection) Write(message []byte) {
	go func(c *Connection) {
		timeout := time.After(1 * time.Second)
		select {
		case c.ChanWrite <- message:
		case <-timeout:
			fmt.Println("Write timeout", time.Now(), string(message))
		}
	}(c)
}

// listen to clients
func main() {
	go func() {
		fmt.Println("Listening websocket on port", ListeningPort)
		err := http.ListenAndServe(ListeningPort, nil)
		if err != nil {
			fmt.Printf("Fail to listen websocket on port %v:\n%v\n",
				ListeningPort, err.Error())
		}
	}()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		cws, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("upgrader.Upgrade err", err.Error())
			return
		}
		bws, _, err := websocket.DefaultDialer.Dial(BackendAddr, nil)
		if err != nil {
			fmt.Println("Dial BackendAddr err", err.Error())
			return
		}
		clientConn := &Connection{WsConn: cws,
			ChanWrite: make(chan []byte), ChanClose: make(chan []byte)}
		backendConn := &Connection{WsConn: bws, ConjugateConn: clientConn,
			ChanWrite: make(chan []byte), ChanClose: make(chan []byte)}
		clientConn.ConjugateConn = backendConn
		go clientConn.ReadPump()
		go clientConn.WritePump()
		go backendConn.ReadPump()
		go backendConn.WritePump()
	})
	select {}
}
