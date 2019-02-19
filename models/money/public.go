package money

import (
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models/player"
)

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	fetchCardTypes()
	fetchPurchaseTypes()
	fetchPaymentRequirement()
}

type Version interface {
	GetVersion() string
}

var version Version

func RegisterVersion(registeredVersion Version) {
	version = registeredVersion
}

func RequestPayment(playerInstance *player.Player, cardCode string) (paymentId int64, err error) {
	action := NewActionContext()
	action.actionType = "requestPayment"
	action.playerInstance = playerInstance
	action.cardCode = cardCode

	response := sendAction(action)
	return response.paymentId, response.err
}

func AcceptPayment(adminId int64, paymentId int64) (err error) {
	action := NewActionContext()
	action.actionType = "acceptPayment"
	action.adminId = adminId
	action.paymentId = paymentId

	response := sendAction(action)
	return response.err
}

func DeclinePayment(adminId int64, paymentId int64) (err error) {
	action := NewActionContext()
	action.actionType = "declinePayment"
	action.adminId = adminId
	action.paymentId = paymentId

	response := sendAction(action)
	return response.err
}
