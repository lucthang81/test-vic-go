package otp

import (
	"github.com/vic/vic_go/log"
)

func isRequestIdUnique(requestId string) bool {
	row := dataCenter.Db().QueryRow("SELECT request_id FROM otp_request WHERE request_id = $1", requestId)
	var scanRequestId string
	err := row.Scan(&scanRequestId)
	if err != nil {
		return true
	}
	return false
}

func recordRequestId(requestId string) {
	_, err := dataCenter.Db().Exec("INSERT INTO otp_request (request_id) VALUES ($1)", requestId)
	if err != nil {
		log.LogSerious("err record otp request id %v", err)
	}
}
