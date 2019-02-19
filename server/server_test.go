package server

import (
	"encoding/json"
	"errors"
	"github.com/go-martini/martini"
	"io"
	"testing"
	"time"
	// "database/sql"
	"fmt"
	// "github.com/vic/vic_go/log"
	"github.com/gorilla/websocket"
	"github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	// "log"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	models   *TestModels
	portTest int
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {
	s.models = &TestModels{}
	s.portTest = 40005
	RegisterModelsInterface(s.models)
}

func (s *TestSuite) TearDownSuite(c *C) {

}

func (s *TestSuite) SetUpTest(c *C) {
	// Use s.dir to prepare some data.
	fmt.Printf("start test %s \n", c.TestName())
}

func (s *TestSuite) TearDownTest(c *C) {

}

/*



THE ACTUAL TESTS




*/

const waitTimeForRequest time.Duration = 1000 * time.Millisecond

// func (s *TestSuite) TestRunnable(c *C) {
// 	c.Assert(false, Equals, false)
// }

// func (s *TestSuite) TestConnectToServer(c *C) {
// 	server := NewServer()
// 	address, testCode, addressUrl := s.genNewAddress()
// 	fmt.Println(address)
// 	server.startServingSocketForTest(address, testCode, "", "")
// 	timeout := time.After(waitTimeForRequest)
// 	<-timeout

// 	// wait for about 1s
// 	c.Assert(len(server.pendingAuthConnections), Equals, 0)
// 	chatClient := NewChatClient(c, addressUrl)
// 	timeout = time.After(waitTimeForRequest)
// 	<-timeout
// 	c.Assert(len(server.pendingAuthConnections), Equals, 1)

// 	chatClient.Close(c)
// 	timeout = time.After(waitTimeForRequest)
// 	<-timeout
// 	c.Assert(len(server.pendingAuthConnections), Equals, 0)
// }

func (s *TestSuite) TestAuthToServer(c *C) {

	server := NewServer()
	address, testCode, addressUrl := s.genNewAddress()
	fmt.Println(address)
	server.startServingSocketForTest(address, testCode, "", "")
	timeout := time.After(waitTimeForRequest)
	<-timeout
	var chatClient *ChatClient

	chatClient = NewChatClient(c, addressUrl)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "success",
		"player_id": 125,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 0)
	c.Assert(len(server.connections.coreMap), Equals, 1)
	connection := server.getConnection(125)
	c.Assert(connection, NotNil)
	c.Assert(connection.isAuthenticated, Equals, true)
	c.Assert(connection.playerId, Equals, int64(125))
	c.Assert(connection.server, Equals, server)
	c.Assert(server.getConnection(connection.playerId), NotNil)
	chatClient.Close(c)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 0)
	c.Assert(len(server.connections.coreMap), Equals, 0)

	// auth fail
	chatClient = NewChatClient(c, addressUrl)

	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "fail",
		"player_id": 129,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 1) // still keep connection so server can send error back
	c.Assert(len(server.connections.coreMap), Equals, 0)
	chatClient.Close(c)

	// auth 2 time same player id
	chatClient = NewChatClient(c, addressUrl)
	chatClient.tag = "chatClient"
	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "success",
		"player_id": 157,
	})
	response := chatClient.ReadData(c)
	fmt.Println(response)
	c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 0)
	c.Assert(len(server.connections.coreMap), Equals, 1)

	chatClient1 := NewChatClient(c, addressUrl)
	chatClient1.tag = "chatClient1"
	timeout = time.After(waitTimeForRequest)
	<-timeout

	chatClient1.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "success",
		"player_id": 157,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout

	response = chatClient1.ReadData(c)
	fmt.Println(response)

	go func() {
		fmt.Println("try to read", chatClient.tag)
		response = chatClient.ReadData(c)
		fmt.Println(response)
		c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 0)
		c.Assert(len(server.connections.coreMap), Equals, 1)
	}()

	chatClient.SendRequest(c, "close_connection", make(map[string]interface{}))
	chatClient1.Close(c)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(true, Equals, true)

	// connect and send request without auth
	chatClient = NewChatClient(c, addressUrl)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth_weird", map[string]interface{}{
		"type": "create_new_player",
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(len(server.pendingAuthConnections.coreMap), Equals, 1) // still keep connection so server can send error back
	c.Assert(len(server.connections.coreMap), Equals, 0)            // connections did not increase
	chatClient.Close(c)
}

func (s *TestSuite) TestSendRequestToServer(c *C) {

	server := NewServer()
	address, testCode, addressUrl := s.genNewAddress()
	fmt.Println(address)
	server.startServingSocketForTest(address, testCode, "", "")
	utils.DelayInDuration(waitTimeForRequest)

	chatClient := NewChatClient(c, addressUrl)
	defer chatClient.Close(c)
	timeout := time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "success",
		"player_id": 157,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.ReadData(c)
	chatClient.SendRequest(c, "echo", map[string]interface{}{
		"type":      "test",
		"want":      "ok",
		"player_id": 125,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	response := chatClient.ReadData(c)
	c.Assert(utils.GetStringAtPath(response, "data/type"), Equals, "test")
	c.Assert(utils.GetStringAtPath(response, "data/want"), Equals, "ok")

	chatClient.SendRequest(c, "want_fail", map[string]interface{}{
		"type":      "test",
		"want":      "ok",
		"player_id": 125,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	response = chatClient.ReadData(c)
	fmt.Println(response)
	c.Assert(utils.GetStringAtPath(response, "error/message"), Equals, "intent_fail")
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(chatClient.AvailableDataCount(c), Equals, 0)
}

func (s *TestSuite) TestSendRequestToClient(c *C) {
	server := NewServer()
	address, testCode, addressUrl := s.genNewAddress()
	fmt.Println(address)
	server.startServingSocketForTest(address, testCode, "", "")
	utils.DelayInDuration(waitTimeForRequest)

	chatClient := NewChatClient(c, addressUrl)
	defer chatClient.Close(c)
	timeout := time.After(waitTimeForRequest)
	<-timeout
	chatClient.SendRequest(c, "auth", map[string]interface{}{
		"type":      "create_new_player",
		"want":      "success",
		"player_id": 157,
	})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	chatClient.ReadData(c)

	server.SendRequest("send_test", map[string]interface{}{"foo": "bar"}, 157)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	response := chatClient.ReadData(c)
	c.Assert(utils.GetStringAtPath(response, "data/foo"), Equals, "bar")
	c.Assert(utils.GetStringAtPath(response, "method"), Equals, "send_test")

	//send to not exist conn
	server.SendRequest("send_test", map[string]interface{}{"foo": "bar"}, 999)
	timeout = time.After(waitTimeForRequest)
	<-timeout
	c.Assert(chatClient.AvailableDataCount(c), Equals, 0)

	// send many, include 1 hit
	server.SendRequests("send_test", map[string]interface{}{"foo": "bar"}, []int64{999, 777, 157})
	timeout = time.After(waitTimeForRequest)
	<-timeout
	response = chatClient.ReadData(c)
	c.Assert(utils.GetStringAtPath(response, "data/foo"), Equals, "bar")
	c.Assert(utils.GetStringAtPath(response, "method"), Equals, "send_test")
}

/*

Helper

*/

var cstDialer = websocket.Dialer{
	Subprotocols:    []string{"p1", "p2"},
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

type ChatClient struct {
	ws            *websocket.Conn
	callIdCounter int64
	dataChannel   chan map[string]interface{}
	ignoreMethods []string
	tag           string
}

func NewChatClient(c *C, url string) *ChatClient {
	ws, _, err := cstDialer.Dial(url, nil)
	c.Assert(err, IsNil)
	chatClient := &ChatClient{
		ws:            ws,
		callIdCounter: 0,
		dataChannel:   make(chan map[string]interface{}, 10),
	}

	ws.SetPongHandler(
		func(string) error {
			return nil
		})
	go chatClient.waitToRead()
	return chatClient
}

func (chatClient *ChatClient) Close(c *C) {
	params := map[string]interface{}{}
	payload := chatClient.GenerateRequest(c, "close_connection", params)
	chatClient.ws.WriteMessage(websocket.TextMessage, []byte(payload))
}

// func (chatClient *ChatClient) Auth(c *C, gameAccount *GameAccount) map[string]interface{} {
// 	token := GetToken(c, gameAccount.id)
// 	apiKey, secretKey := GetApiKeySecretKey(c, gameAccount.gameId)
// 	c.Assert(token != "", Equals, true)
// 	params := map[string]interface{}{
// 		"token":      token,
// 		"api_key":    apiKey,
// 		"secret_key": secretKey,
// 	}
// 	chatClient.SendRequest(c, "auth", params)
// 	return chatClient.ReadData(c)
// }

func (chatClient *ChatClient) waitToRead() {
	for {
		messageType, p, err := chatClient.ws.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("error in read %v \n", err)
			}
			chatClient.ws.Close()
			break
			return
		}
		switch messageType {
		case websocket.TextMessage:
			// translate to json, break if error
			var data map[string]interface{}
			err := json.Unmarshal(p, &data)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(data)
			willIgnore := false
			for _, ignoreMethod := range chatClient.ignoreMethods {
				if ignoreMethod == utils.GetStringAtPath(data, "method") {
					willIgnore = true
					break
				}
			}
			if !willIgnore {
				fmt.Println("send data to channel", data, chatClient.tag)
				chatClient.dataChannel <- data
			}
		case websocket.BinaryMessage:
			// just ignore
			break
		case websocket.CloseMessage:
			// close frame, just close
			fmt.Printf("close %v", string(p))
			break
		case websocket.PingMessage:
			//just ignore
			fmt.Println("ping message")
		case websocket.PongMessage:
			// just ignore
			fmt.Println("pong message")
		}

	}
}

func (chatClient *ChatClient) ReadData(c *C) (data map[string]interface{}) {
	fmt.Println("read data from channel", chatClient.tag)
	return <-chatClient.dataChannel
}

func (chatClient *ChatClient) AvailableDataCount(c *C) (count int) {
	return len(chatClient.dataChannel)
}

func (chatClient *ChatClient) GenerateRequest(c *C, method string, params map[string]interface{}) (payload string) {
	data_map := make(map[string]interface{})
	data_map["method"] = method
	data_map["data"] = params
	data_map["callId"] = fmt.Sprintf("callId_%d", chatClient.callIdCounter)
	chatClient.callIdCounter = chatClient.callIdCounter + 1

	bytePayload, err := json.Marshal(data_map)
	c.Assert(err, IsNil)
	return string(bytePayload)
}

func (chatClient *ChatClient) SendRequest(c *C, method string, params map[string]interface{}) {
	payload := chatClient.GenerateRequest(c, method, params)
	err := chatClient.ws.WriteMessage(websocket.TextMessage, []byte(payload))
	c.Assert(err, IsNil)
}

// func IsSuccessResponse(response map[string]interface{}) bool {
// 	if val, ok := response["success"]; ok {
// 		return val.(bool)
// 	}
// 	return false
// }

// func GetMethodFromResponse(response map[string]interface{}) string {
// 	if val, ok := response["method"]; ok {
// 		return val.(string)
// 	}
// 	return ""
// }

// fake the models

type TestModels struct {
}

func (testModels *TestModels) HandleRequest(requestType string, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	if requestType == "echo" {
		return data, nil
	} else if requestType == "want_fail" {
		return nil, errors.New("intent_fail")
	}
	return map[string]interface{}{}, nil
}

func (testModels *TestModels) HandleAuth(data map[string]interface{}, ipAddress string) (authStatus bool, authData map[string]interface{}, err error) {
	wantSuccess := utils.GetStringAtPath(data, "want")
	playerId := utils.GetInt64AtPath(data, "player_id")
	authData = make(map[string]interface{})
	authData["player_id"] = playerId
	if wantSuccess == "success" {
		fmt.Println("success auth")
		return true, authData, nil
	} else {
		fmt.Println("fail auth")
		return false, map[string]interface{}{}, errors.New("intent_fail")
	}
}

func (testModels *TestModels) HandleOffline(playerId int64) {

}

func (testModels *TestModels) HandleLogout(playerId int64) {

}

func (testModels *TestModels) HandleHttp(m *martini.ClassicMartini, staticFolderAddress string, mediaFolderAddress string, staticRoot string, mediaRoot string) {

}

func (s *TestSuite) genNewAddress() (address string, testCode string, addressSocketUrl string) {
	s.portTest++
	address = fmt.Sprintf("0.0.0.0:%d", s.portTest)
	testCode = fmt.Sprintf("/ws%d", s.portTest)
	addressSocketUrl = fmt.Sprintf("ws://%s%s", address, testCode)
	return address, testCode, addressSocketUrl
}
