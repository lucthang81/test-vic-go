package server

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func init() {
	fmt.Print("")
}

type Request struct {
	callId           string
	method           string
	requiredCallback bool
	data             map[string]interface{}
	payload          string
	connection       *Connection
}

func newRequest(data map[string]interface{}, connection *Connection) *Request {
	request := &Request{}
	request.callId, _ = data["callId"].(string)
	request.method, _ = data["method"].(string)
	request.data, _ = data["data"].(map[string]interface{})

	request.connection = connection
	return request
}

func newRequestForClient(method string, data map[string]interface{}) *Request {
	request := &Request{}
	request.method = method
	request.data = data
	return request
}

func (request *Request) getPayload() []byte {
	data_map := make(map[string]interface{})

	if request.callId != "" {
		data_map["callId"] = request.callId
	}
	data_map["requiredCallback"] = request.requiredCallback
	data_map["method"] = request.method
	data_map["data"] = request.data
	payload, _ := json.Marshal(data_map)
	return payload
}

func (request *Request) getMessage() *WebsocketMessage {
	payload := request.getPayload()
	message := &WebsocketMessage{}
	message.messageType = websocket.TextMessage
	message.payload = payload
	return message
}
