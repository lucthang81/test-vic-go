// Package rank provides functionality for ranking players, measuring by a key
//
package rank

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	//	"github.com/lib/pq"
	"github.com/daominah/livestream/misc"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/zconfig"
)

const (
	LEADERBOARD_LIMIT         = int(100)
	LEADERBOARD_UPDATE_PERIOD = 120 * time.Second

	RANK_TEST1 = int64(1)
	RANK_TEST2 = int64(2)

	RANK_NUMBER_OF_WINS = int64(4)
)

var MapRankIdToLeaderboard map[int64][]TopRow
var GMutex sync.Mutex

func updateLeaderboards() {
	for {
		for i := int64(1); i <= 4; i++ {
			lb, _ := LoadLeaderboard(i)
			GMutex.Lock()
			MapRankIdToLeaderboard[i] = lb
			GMutex.Unlock()
		}
		time.Sleep(LEADERBOARD_UPDATE_PERIOD)
	}
}

func init() {
	_ = fmt.Println
	_ = datacenter.NewDataCenter
	_ = zconfig.PostgresAddress
	//
	dataCenterInstance := datacenter.NewDataCenter(
		zconfig.PostgresUsername, zconfig.PostgresPassword,
		zconfig.PostgresAddress, zconfig.PostgresDatabaseName,
		zconfig.RedisAddress)
	record.RegisterDataCenter(dataCenterInstance)
	//	fmt.Println("record.DataCenter", record.DataCenter)
	//
	MapRankIdToLeaderboard = make(map[int64][]TopRow)
	go updateLeaderboards()
}

// update rank.started_time,
// insert into rank_archive,
// delete all rows in rank_key with the rankId
func Reset(rankId int64) error {
	var rank_name string
	var started_time time.Time
	row := record.DataCenter.Db().QueryRow(
		`SELECT rank_name, started_time FROM rank  WHERE rank_id = $1`, rankId)
	err := row.Scan(&rank_name, &started_time)
	if err != nil {
		return err
	}
	//
	_, err = record.DataCenter.Db().Exec(
		`UPDATE rank SET started_time = $1 WHERE rank_id = $2`, time.Now(), rankId)
	if err != nil {
		return err
	}
	//
	leaderboard, err := LoadLeaderboard(rankId)
	if err != nil {
		return err
	}
	full_order, err := json.Marshal(leaderboard)
	if err != nil {
		return err
	}
	var top10Slice []TopRow
	if len(leaderboard) >= 10 {
		top10Slice = leaderboard[0:10]
	} else {
		top10Slice = leaderboard
	}
	top10, err := json.Marshal(top10Slice)
	if err != nil {
		return err
	}
	_, err = record.DataCenter.Db().Exec(
		`INSERT INTO rank_archive
            (rank_id, rank_name, started_time, finished_time, top_10, full_order)
        VALUES ($1, $2, $3, $4, $5, $6) `,
		rankId, rank_name, started_time, time.Now(),
		string(top10), string(full_order))
	if err != nil {
		return err
	}
	//
	_, err = record.DataCenter.Db().Exec(
		`DELETE FROM rank_key WHERE rank_id = $1`, rankId)
	return err
}

func ChangeKey(rankId int64, userId int64, amount float64) error {
	r, err := record.DataCenter.Db().Exec(
		`UPDATE rank_key
		SET rkey = rkey + $1, last_modified = $2
		WHERE rank_id = $3 AND user_id = $4`,
		amount, time.Now(), rankId, userId)
	if err != nil {
		return err
	}
	nRowsAffected, _ := r.RowsAffected()
	if nRowsAffected != 0 {
		return nil
	}
	_, err = record.DataCenter.Db().Exec(
		`INSERT INTO rank_key (rank_id, user_id, rkey, last_modified)
		VALUES ($3, $4, $1, $2)`,
		0, time.Now(), rankId, userId)
	if err != nil {
		return err
	}
	_, err = record.DataCenter.Db().Exec(
		`UPDATE rank_key
    		SET rkey = rkey + $1, last_modified = $2
    		WHERE rank_id = $3 AND user_id = $4`,
		amount, time.Now(), rankId, userId)
	return err

}

//
func LoadKeyAndPosition(rankId int64, userId int64) (float64, int, error) {
	GMutex.Lock()
	leaderboard := MapRankIdToLeaderboard[rankId]
	GMutex.Unlock()
	if leaderboard == nil {
		return 0, 0, errors.New("leaderboard == nil")
	}
	keys := make([]float64, 0)
	for _, row := range leaderboard {
		keys = append(keys, row.RKey)
	}
	row := record.DataCenter.Db().QueryRow(
		`SELECT rkey FROM rank_key WHERE rank_id = $1 AND user_id = $2 `,
		rankId, userId)
	var rkey float64
	e := row.Scan(&rkey)
	if e != nil {
		return 0, 0, e
	}
	position, e := misc.BisectRight(keys, rkey)
	if e != nil {
		return 0, 0, e
	}
	return rkey, position, nil
}

type TopRow struct {
	RankId int64
	UserId int64
	RKey   float64
}

// always return non-nil rows.
func LoadLeaderboard(rankId int64) ([]TopRow, error) {
	result := make([]TopRow, 0)
	rows, e := record.DataCenter.Db().Query(
		`SELECT rank_id, user_id, rkey FROM rank_key
	    WHERE rank_id = $1
	    ORDER BY rkey DESC, last_modified DESC LIMIT $2`,
		rankId, LEADERBOARD_LIMIT)
	if e != nil {
		return result, e
	}
	defer rows.Close()
	for rows.Next() {
		var rank_id, user_id int64
		var rkey float64
		e := rows.Scan(&rank_id, &user_id, &rkey)
		if e != nil {
			return result, e
		}
		result = append(result,
			TopRow{RankId: rank_id, UserId: user_id, RKey: rkey})
	}
	return result, nil
}

func GetLeaderboard(rankId int64) []TopRow {
	var leaderboard []TopRow
	GMutex.Lock()
	leaderboard = MapRankIdToLeaderboard[rankId]
	GMutex.Unlock()
	return leaderboard
}
