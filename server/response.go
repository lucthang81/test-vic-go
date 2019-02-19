package server

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/lib/pq"

	"github.com/vic/vic_go/details_error"
)

type Response struct {
	callId  string                 `json:"callId"`
	data    map[string]interface{} `json:"data"`
	error   map[string]interface{} `json:"error"`
	payload string
	success bool
}

func newResponse(data map[string]interface{}) *Response {
	response := &Response{}
	response.data = data

	return response
}

func newErrorResponse(err error, callId string) *Response {
	response := &Response{}
	response.callId = callId
	response.data = make(map[string]interface{})
	errorData := make(map[string]interface{})
	response.error = errorData
	if val, ok := err.(*pq.Error); ok {
		errorData["code"] = val.Code
		errorData["message"] = "Database error"
		detail := make(map[string]interface{})
		detail["second_message"] = err.Error()
		errorData["details"] = detail
	} else if val, ok := err.(*details_error.DetailsError); ok {
		errorData["message"] = val.Error()
		errorData["details"] = val.Details()
	} else {
		errorData["message"] = err.Error()
	}
	return response
}

func newSuccessResponse(callId string) *Response {
	response := &Response{}
	response.callId = callId
	response.data = make(map[string]interface{})
	response.success = true
	return response
}

func (response *Response) getPayload() []byte {
	data_map := make(map[string]interface{})
	data_map["callId"] = response.callId
	if len(response.data) > 0 {
		data_map["data"] = response.data
	}
	if len(response.error) > 0 {
		data_map["error"] = response.error
	}
	data_map["success"] = response.success

	payload, _ := json.Marshal(data_map)
	return payload
}

func (response *Response) getMessage() *WebsocketMessage {
	payload := response.getPayload()
	message := &WebsocketMessage{}
	message.messageType = websocket.TextMessage
	message.payload = payload
	return message
}
