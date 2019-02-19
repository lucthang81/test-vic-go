package event

import (
	"fmt"
	"testing"
	"time"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models"
)

var _, _ = fmt.Println("")

func init() {
	//
	_ = time.Now()
	//
	dataCenterInstance := datacenter.NewDataCenter(
		"vic_user", "123qwe",
		"127.0.0.1:5432", "casino_vic_db",
		":6379",
	)
	models.RegisterDataCenter(dataCenterInstance)
}

func TestHaha(t *testing.T) {
	//	event := EventTop{
	//		EventName:     "TestEvent",
	//		StartingTime:  time.Now(),
	//		FinishingTime: time.Now().Add(60 * time.Minute),
	//		MapPositionToPrize: map[int]int64{
	//			1: 1000000,
	//			2: 100000,
	//			3: 20000,
	//		},
	//		MapPlayerIdToValue: make(map[int64]int64),
	//	}
	//	event.SetNewValue(2987, 4)
	//	event.SetNewValue(3976, 6)
	//	event.SetNewValue(0, 2)
	//	SaveEvent(event)
	event := LoadEvent("TestEvent")
	fmt.Print("event", event)
	event.start()
	select {}
}
