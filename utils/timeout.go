package utils

import (
	"time"
)

type TimeOut struct {
	shouldHandle bool
	outFor       time.Duration
	cancelChan   chan bool
	startTime    time.Time
}

func NewTimeOut(outFor time.Duration) *TimeOut {
	return &TimeOut{
		shouldHandle: true,
		outFor:       outFor,
		cancelChan:   make(chan bool),
	}
}

func (timeOut *TimeOut) Start() bool {
	realTimeOut := time.After(timeOut.outFor)
	timeOut.startTime = time.Now()
	select {
	case <-realTimeOut:
		return timeOut.shouldHandle
	case <-timeOut.cancelChan:
		return timeOut.shouldHandle
	}
	return true
}

func (timeOut *TimeOut) SetShouldHandle(shouldHandle bool) {
	timeOut.shouldHandle = shouldHandle
	if !shouldHandle {
		go timeOut.sendCancelEvent()
	}
}

func (timeOut *TimeOut) sendCancelEvent() {
	timeOut.cancelChan <- true
}

func (timeOut *TimeOut) GetSecondsUntilsTimerEnd() float64 {
	timeDuration := timeOut.outFor - time.Now().Sub(timeOut.startTime)
	return timeDuration.Seconds()
}
