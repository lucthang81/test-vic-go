package server

import (
	"fmt"
	//"io"
	//"bufio"
	"io/ioutil"
	//"net"
	"net/http"
	//"strings"
	"encoding/json"
	"time"

	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"

	"github.com/vic/vic_go/log"
	//"github.com/vic/vic_go/utils"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

func init() {
	_, _ = json.MarshalIndent(map[int]int{}, "", "    ")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ModelsInterface interface {
	HandleRequest(requestType string, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error)
	HandleAuth(data map[string]interface{}, ipAddress string) (authStatus bool, authData map[string]interface{}, err error)
	HandleOffline(playerId int64)
	HandleLogout(playerId int64)

	HandleHttp(m *martini.ClassicMartini, staticFolderAddress string, mediaFolderAddress string, staticRoot string, mediaRoot string)
}

var models ModelsInterface

func RegisterModelsInterface(registeredModels ModelsInterface) {
	models = registeredModels
}

type Server struct {
	// Registered connections.
	pendingAuthConnections *PendingAuthConnMap

	connections *ConnMap

	counter  int64
	upgrader websocket.Upgrader

	pendingAuthChan      chan *Connection
	authConnectionChan   chan *Connection
	removeConnectionChan chan *Connection
	closeChan            chan bool

	// status
	numberOfRequests         int64
	averageRequestHandleTime float64
}

func NewServer() (server *Server) {
	server = &Server{
		connections:            NewConnMap(),
		pendingAuthConnections: NewPendingAuthConnMap(),
		counter:                0,
		upgrader:               upgrader,
		pendingAuthChan:        make(chan *Connection),
		authConnectionChan:     make(chan *Connection),
		removeConnectionChan:   make(chan *Connection),
		closeChan:              make(chan bool),
	}
	go server.waitForConnectionsEvent()
	return server
}

func (server *Server) StartServingSocket(socketAddress string, sslPemPath string, sslKeyPath string) {
	go server.startHandleRequest()
	go server.startListenToRequest(socketAddress, sslPemPath, sslKeyPath)
}

func (server *Server) StartServingHttp(httpAddress string,
	sslPemPath string,
	sslKeyPath string,
	staticFolderAddress string,
	mediaFolderAddress string,
	staticRoot string,
	mediaRoot string,
	projectRoot string) {
	go server.startHandleHttpRequest(httpAddress, sslPemPath, sslKeyPath, staticFolderAddress, mediaFolderAddress, staticRoot, mediaRoot, projectRoot)
}

func (server *Server) startHandleRequest() {
	http.HandleFunc("/crossdomain.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<?xml version="1.0" ?>
<cross-domain-policy> 
  <site-control permitted-cross-domain-policies="master-only"/>
  <allow-access-from domain="*"/>
  <allow-http-request-headers-from domain="*" headers="*"/>
</cross-domain-policy>`))
	})
	http.HandleFunc("/paybnbipn", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}"))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		// fmt.Printf("http connect header: %#v\n", r.Header)
		body, _ := ioutil.ReadAll(r.Body)
		_ = fmt.Sprintf("http connect body: %#v\n", body)
		ws, err := server.upgrader.Upgrade(w, r, nil)
		// fmt.Printf("websocket err", err)
		if err != nil {
			log.Log(err.Error())
			return
		}
		// fmt.Println(net.SplitHostPort(r.RemoteAddr))
		c := &Connection{
			isAuthenticated:       false,
			ws:                    ws,
			id:                    server.counter,
			incoming:              make(chan *WebsocketMessage, 10),
			outgoing:              make(chan []byte, 10),
			auth:                  make(chan bool, 1),
			closeWaitForAuthChan:  make(chan bool, 1),
			closeWaitForReadChan:  make(chan bool, 1),
			closeWaitForWriteChan: make(chan bool, 1),
		}
		server.counter++
		server.pendingAuthChan <- c
		c.start()
	})
}

func (server *Server) startListenToRequest(url string, sslPemPath string, sslKeyPath string) {
	if sslPemPath != "" && sslKeyPath != "" {
		err := http.ListenAndServeTLS(url, sslPemPath, sslKeyPath, nil)
		if err != nil {
			log.LogSerious("ListenAndServe: ", err.Error())
		}
	} else {
		// same func on 2 ports
		go func() {
			err1 := http.ListenAndServe(":4007", nil)
			fmt.Println("haha", err1)
		}()
		err := http.ListenAndServe(url, nil)
		if err != nil {
			log.LogSerious("ListenAndServe: ", err.Error())
		}
	}
}

func (server *Server) LogoutPlayer(playerId int64) {
	go func(serverInRoutine *Server, playerIdInRoutine int64) {
		request := newRequestForClient("logout", map[string]interface{}{})
		conn := serverInRoutine.connections.get(playerIdInRoutine)
		if conn != nil {
			conn.handleRequest(request)
		}
	}(server, playerId)
	// TODO we may need to send and check for success (implement callId, reply from client, currently send and ignore)
}

// already run in a goroutine
func (server *Server) SendRequest(requestType string, data map[string]interface{}, toPlayerId int64) {
	response := newRequestForClient(requestType, data)
	message := response.getMessage()
	go func(serverInRoutine *Server, playerIdInRoutine int64) {
		conn := serverInRoutine.connections.get(playerIdInRoutine)
		if conn != nil {
			conn.incoming <- message
		}
	}(server, toPlayerId)
	// TODO we may need to send and check for success (implement callId, reply from client, currently send and ignore)
}

// already run in a goroutine
func (server *Server) sendRequestToConnection(requestType string, data map[string]interface{}, conn *Connection) {
	response := newRequestForClient(requestType, data)
	message := response.getMessage()
	go func(connInRoutine *Connection) {
		connInRoutine.incoming <- message
	}(conn)
	// TODO we may need to send and check for success (implement callId, reply from client, currently send and ignore)
}

// already run in a goroutine
func (server *Server) SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64) {
	response := newRequestForClient(requestType, data)
	message := response.getMessage()
	for _, playerId := range toPlayerIds {
		go func(serverInRoutine *Server, playerIdInRoutine int64) {
			conn := serverInRoutine.connections.get(playerIdInRoutine)
			if conn != nil {
				conn.incoming <- message
			}
		}(server, playerId)
	}
}

// already run in a goroutine
func (server *Server) SendRequestsToAll(requestType string, data map[string]interface{}) {
	response := newRequestForClient(requestType, data)
	message := response.getMessage()
	server.connections.rLock()
	for _, connection := range server.connections.coreMap {
		go func(connectionInRoutine *Connection) {
			if connectionInRoutine != nil {
				connectionInRoutine.incoming <- message
			}
		}(connection)
	}
	server.connections.rUnlock()
}

func (server *Server) DisconnectPlayer(playerId int64, data map[string]interface{}) {
	conn := server.connections.get(playerId)
	if conn != nil {
		conn.handleCloseFromErrorInReadOrWrite()
	}
}

func (server *Server) waitForConnectionsEvent() {
	for {
		select {
		case c := <-server.pendingAuthChan:
			server.pendingAuthConnections.set(c, true)
			c.server = server
		case c := <-server.authConnectionChan:
			oldConn := server.connections.get(c.playerId)
			if oldConn != nil && !oldConn.isClosing {
				// notify session timeout
				server.sendRequestToConnection("session_timeout", make(map[string]interface{}), oldConn)
			}
			server.pendingAuthConnections.delete(c)
			server.connections.set(c.playerId, c)
		case c := <-server.removeConnectionChan:
			if c.isAuthenticated {
				// check if the connection that will be remove is the same connection that active
				// there is a case when the connection that is being closed is due to session_timeout by
				// another connection login, then replace it
				oldConn := server.connections.get(c.playerId)
				if oldConn == c {
					server.connections.delete(c.playerId)
				}
			} else {
				server.pendingAuthConnections.delete(c)
			}
		case <-server.closeChan:
			return
		}
	}
}

func (server *Server) getConnection(playerId int64) *Connection {
	return server.connections.get(playerId)
}

func (server *Server) Close() {
	server.closeChan <- true
}

func (server *Server) NumberOfRequest() int64 {
	return server.numberOfRequests
}

func (server *Server) AverageRequestHandleTime() float64 {
	return server.averageRequestHandleTime
}
