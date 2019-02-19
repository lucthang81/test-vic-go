package zmisc

import (
	"github.com/vic/vic_go/datacenter"
)

const (
	GLOBAL_TEXT_LOWER_BOUND = int64(100000)

	GLOBAL_TEXT_TYPE_BIG_WIN      = "GLOBAL_TEXT_TYPE_BIG_WIN"
	GLOBAL_TEXT_TYPE_SPACE_HOLDER = "GLOBAL_TEXT_TYPE_SPACE_HOLDER"
	GLOBAL_TEXT_TYPE_OTHER        = "GLOBAL_TEXT_TYPE_OTHER"
)

// global database connection
var dataCenter *datacenter.DataCenter

func RegisterDataCenter(registeredDataCenter *datacenter.DataCenter) {
	dataCenter = registeredDataCenter
}
