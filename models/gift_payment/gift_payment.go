package gift_payment

import (
	"github.com/vic/vic_go/datacenter"
)

var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
	loadData()
	loadGiftPaymentTypes()
}

type PlayerInterface interface {
	Id() int64

	GetMoney(currencyType string) int64
	LockMoney(currencyType string)
	UnlockMoney(currencyType string)
	IncreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error)
	DecreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error)

	Bet() int64

	// commu
	GetUnreadCountOfInboxMessages() (total int64, err error)
	CreateRawMessage(title string, content string) (err error)
	AppType() string
	DeviceType() string
	APNSDeviceToken() string
	GCMDeviceToken() string
}

func VipPointFromBet(bet int64) int64 {
	return bet / vipPointRate
}
