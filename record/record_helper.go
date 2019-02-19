package record

import (
	"database/sql"
	"github.com/vic/vic_go/log"
)

func getInt64FromQuery(queryString string, a ...interface{}) int64 {
	row := dataCenter.Db().QueryRow(queryString, a...)
	var value sql.NullInt64
	err := row.Scan(&value)
	if err != nil {
		log.LogSerious("Error fetch general money data %v", err)
	}
	return value.Int64
}
