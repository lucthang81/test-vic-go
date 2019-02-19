package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/vic/vic_go/encryption"
	"github.com/vic/vic_go/language"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
)

const (
	IS_SHOW_LOG = false
	IS_ENCRYPT  = false
)

func init() {
	_, _ = encryption.EncryptAesCbc("")
}

type Connection struct {
	id     int64
	ws     *websocket.Conn
	server *Server

	isAuthenticated bool
	playerId        int64

	callIdCounter int64
	incoming      chan *WebsocketMessage
	outgoing      chan []byte
	auth          chan bool

	closeWaitForAuthChan  chan bool
	closeWaitForReadChan  chan bool
	closeWaitForWriteChan chan bool
	isClosing             bool
}

type WebsocketMessage struct {
	payload     []byte
	messageType int
}

func (c *Connection) start() {
	go c.waitForAuthentication()
	go c.waitForRead()
	go c.waitForWrite()
}

func (c *Connection) waitForAuthentication() {
	timeout := time.After(30 * time.Second)
	select {
	case <-timeout:
		c.sendClose(websocket.CloseProtocolError, l.Get(l.M0002))
	case auth := <-c.auth:
		if !auth {
			c.sendClose(websocket.CloseProtocolError, l.Get(l.M0002))
		}
	case <-c.closeWaitForAuthChan:
		return
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) waitForRead() {
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(
		func(string) error {
			c.ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	for {
		select {
		case <-c.closeWaitForReadChan:
			return
		default:
			messageType, reader, err := c.ws.NextReader()
			if err != nil {
				c.handleCloseFromErrorInReadOrWrite()
				return
			}

			p, err := ioutil.ReadAll(reader)
			if err != nil {
				c.handleCloseFromErrorInReadOrWrite()
				return
			}

			if IS_ENCRYPT {
				temp, err := encryption.DecryptAesCbc(string(p))
				if err == nil {
					p = []byte(temp)
				}
			}

			switch messageType {
			case websocket.TextMessage:
				// translate to json, break if error
				var data map[string]interface{}
				err := json.Unmarshal(p, &data)

				if IS_SHOW_LOG {
					fmt.Println("_______________________________________________________________")
					fmt.Println(time.Now().Format(time.RFC3339), "request: pId ", c.playerId, utils.PFormat(data))
				}

				if err != nil {
					log.LogSerious("error when parse text message %s", err.Error())
					c.sendClose(websocket.CloseProtocolError, l.Get(l.M0004))
				} else {
					go c.handleData(data)
				}
			case websocket.BinaryMessage:
				// not support, will send back error
				c.sendClose(websocket.CloseProtocolError, l.Get(l.M0005))
			}
		}

	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) waitForWrite() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case message := <-c.incoming:
			var data map[string]interface{}
			_ = json.Unmarshal(message.payload, &data)

			if IS_SHOW_LOG {
				fmt.Println("_______________________________________________________________")
				fmt.Println(time.Now().Format(time.RFC3339), "response: pId ", c.playerId, utils.PFormat(data))
			}

			if err := c.write(message.messageType, message.payload); err != nil {
				c.handleCloseFromErrorInReadOrWrite()
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				c.handleCloseFromErrorInReadOrWrite()
				return
			}
		case <-c.closeWaitForWriteChan:
			return
		}

	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	if IS_ENCRYPT {
		temp, err := encryption.EncryptAesCbc(string(payload))
		if err == nil {
			payload = []byte(temp)
		}
	}

	return c.ws.WriteMessage(mt, payload)
}

func (c *Connection) sendClose(closeType int, reason string) {
	message := &WebsocketMessage{}
	message.messageType = websocket.CloseMessage
	message.payload = websocket.FormatCloseMessage(closeType, reason)
	c.incoming <- message
	// wait for close confirmation frame
}

func (c *Connection) handleData(data map[string]interface{}) {

	if data["method"] != nil { // request to server
		request := newRequest(data, c)
		c.handleRequest(request)
	} else { // response from client

	}
}

func (c *Connection) handleRequest(request *Request) {
	defer func() {
		if r := recover(); r != nil {
			// send crash back
			response := newErrorResponse(errors.New(l.Get(l.M0003)), request.callId)
			message := response.getMessage()
			c.incoming <- message
			log.SendMailWithCurrentStack("request crash")
		}
	}()
	//	fmt.Println("handleRequest cp 1")
	if request.method == "close_connection" {
		c.sendClose(websocket.CloseNormalClosure, "client_want_to_close")
	} else if request.method == "logout" {
		models.HandleLogout(c.playerId)
		c.sendClose(websocket.CloseNormalClosure, "client_want_to_close")
	} else if !c.isAuthenticated {
		//		fmt.Println("!c.isAuthenticated cp 1")
		var err error
		var authData map[string]interface{}
		var response *Response
		var message *WebsocketMessage

		ipAddressRaw := c.ws.RemoteAddr().String()
		//fmt.Println("ipAddressRaw", ipAddressRaw)
		tokens := strings.Split(ipAddressRaw, ":")
		var ipAddress string
		if len(tokens) > 0 {
			ipAddress = tokens[0]
		}
		_, authData, err = models.HandleAuth(request.data, ipAddress)
		if err != nil {
			response = newErrorResponse(err, request.callId)
			message = response.getMessage()
			c.incoming <- message
			c.isAuthenticated = false
			return
		}
		c.isAuthenticated = true
		c.playerId = utils.GetInt64AtPath(authData, "player_id")
		c.auth <- true
		c.server.authConnectionChan <- c

		response = newSuccessResponse(request.callId)
		response.data = authData
		message = response.getMessage()
		c.incoming <- message
	} else {
		var responseData map[string]interface{}
		var err error
		var response *Response
		var message *WebsocketMessage
		// log.Log("connection playerId %d handle request %s %v", c.playerId, request.method, request.data)
		now := time.Now()

		timeout := time.After(15 * time.Second)
		channel := make(chan *ModelsResponse)

		go func(channelInRoutine chan *ModelsResponse, modelsInRoutine ModelsInterface, method string, data map[string]interface{}, id int64) {

			defer func() {
				if r := recover(); r != nil {
					// send crash back
					response := newErrorResponse(errors.New(l.Get(l.M0003)), request.callId)
					message := response.getMessage()
					c.incoming <- message
					log.SendMailWithCurrentStack(fmt.Sprintf("request crash %v", r))
				}
			}()

			// fmt.Println("method", method, data)
			responseData, err := modelsInRoutine.HandleRequest(method, data, id)
			modelsResponse := &ModelsResponse{
				// data: utils.ConvertData(responseData),
				data: responseData,
				err:  err,
			}
			ta1s := time.After(1 * time.Second)
			select {
			case channelInRoutine <- modelsResponse:
				//				fmt.Println("normal case")
			case <-ta1s:
				//				fmt.Println("ERROR: too late to response")
			}
		}(channel, models, request.method, request.data, c.playerId)

		select {
		case <-timeout:
			err = errors.New(l.Get(l.M0001))
		case response := <-channel:
			responseData = response.data
			err = response.err
		}

		duration := time.Since(now)
		// log.Log("connection playerId %d got response %v error %v", c.playerId, responseData, err)
		c.server.averageRequestHandleTime = ((float64(c.server.numberOfRequests) * c.server.averageRequestHandleTime) + duration.Seconds()) / (float64(c.server.numberOfRequests) + 1)
		c.server.numberOfRequests++
		if err != nil {
			response = newErrorResponse(err, request.callId)
			message = response.getMessage()
			c.incoming <- message
			return
		}

		response = newSuccessResponse(request.callId)
		response.data = responseData
		message = response.getMessage()
		c.incoming <- message
	}
}

type ModelsResponse struct {
	data map[string]interface{}
	err  error
}

func (c *Connection) handleCloseFromErrorInReadOrWrite() {
	if !c.isClosing {
		c.isClosing = true

		if c.isAuthenticated {
			models.HandleOffline(c.playerId)
		}
		if len(c.closeWaitForWriteChan) == 0 {
			c.closeWaitForWriteChan <- true
		}
		if len(c.closeWaitForReadChan) == 0 {
			c.closeWaitForReadChan <- true
		}
		if len(c.closeWaitForAuthChan) == 0 {
			c.closeWaitForAuthChan <- true
		}
		c.ws.Close()

		if c != nil {
			if c.server != nil {
				c.server.removeConnectionChan <- c
			}
		}
	}
}
