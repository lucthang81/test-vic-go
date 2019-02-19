package player

import (
	"fmt"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"sort"
)

var vipDataList map[string]*VipData

const VipDataDatabaseTableName string = "vip_data"
const VipRecordDatabaseTableName string = "vip_record"

type VipData struct {
	id   int64
	code string

	name                        string
	requirementScore            int64
	timeBonusMultiplier         float64
	megaTimeBonusMultiplier     float64
	leaderboardRewardMultiplier float64
	purchaseMultiplier          float64
}

func (vipData *VipData) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["id"] = vipData.id
	data["code"] = vipData.code
	data["name"] = vipData.name
	data["requirement_score"] = vipData.requirementScore
	data["time_bonus_multiplier"] = vipData.timeBonusMultiplier
	data["mega_time_bonus_multiplier"] = vipData.megaTimeBonusMultiplier
	data["leaderboard_reward_multiplier"] = vipData.leaderboardRewardMultiplier
	data["purchase_multiplier"] = vipData.purchaseMultiplier
	return data
}

type ByVipCode []map[string]interface{}

func (a ByVipCode) Len() int      { return len(a) }
func (a ByVipCode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVipCode) Less(i, j int) bool {
	data1 := a[i]
	data2 := a[j]
	code1 := utils.GetStringAtPath(data1, "code")
	code2 := utils.GetStringAtPath(data2, "code")
	return code1 < code2
}

func refreshVipData() {
	if vipDataList == nil {
		vipDataList = make(map[string]*VipData)
	}
	queryString := fmt.Sprintf("SELECT id, code,name, requirement_score, time_bonus_multiplier,mega_time_bonus_multiplier, leaderboard_reward_multiplier, purchase_multiplier FROM %s", VipDataDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString)
	if err != nil {
		log.LogSerious("error fetch vip data %s %v", queryString, err)
		return
	}

	for rows.Next() {
		var id int64
		var code string
		var name string
		var requirementScore int64
		var megaTimeBonusMultiplier float64
		var timeBonusMultiplier float64
		var leaderboardRewardMultiplier float64
		var purchaseMultiplier float64
		err = rows.Scan(&id, &code, &name, &requirementScore, &timeBonusMultiplier, &megaTimeBonusMultiplier, &leaderboardRewardMultiplier, &purchaseMultiplier)
		if err != nil {
			rows.Close()
			log.LogSerious("error fetch vip data %s %v", queryString, err)
			return
		}
		var vipData *VipData
		if vipDataList[code] == nil {
			vipData = &VipData{}
			vipDataList[code] = vipData
		} else {
			vipData = vipDataList[code]
		}
		vipData.id = id
		vipData.code = code
		vipData.name = name
		vipData.requirementScore = requirementScore
		vipData.timeBonusMultiplier = timeBonusMultiplier
		vipData.leaderboardRewardMultiplier = leaderboardRewardMultiplier
		vipData.purchaseMultiplier = purchaseMultiplier
		vipData.megaTimeBonusMultiplier = megaTimeBonusMultiplier
		vipData.name = name
	}
	rows.Close()
}

func getVipDataList() []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, vipData := range vipDataList {
		results = append(results, vipData.SerializedData())
	}
	sort.Sort(ByVipCode(results))
	return results
}

func editVipData(data map[string]interface{}) error {
	queryString := fmt.Sprintf("UPDATE %s SET name = $1, requirement_score = $2, time_bonus_multiplier = $3,"+
		" mega_time_bonus_multiplier = $4, leaderboard_reward_multiplier = $5, purchase_multiplier = $6", VipDataDatabaseTableName)
	_, err := dataCenter.Db().Exec(queryString, data["name"], data["requirement_score"], data["time_bonus_multiplier"],
		data["mega_time_bonus_multiplier"], data["leaderboard_reward_multiplier"], data["purchase_multiplier"])
	if err != nil {
		log.LogSerious("err edit vip data %s %v", queryString, err)
		return err
	}
	code := utils.GetStringAtPath(data, "code")
	vipData := vipDataList[code]
	vipData.name = utils.GetStringAtPath(data, "name")
	vipData.requirementScore = utils.GetInt64AtPath(data, "requirement_score")
	vipData.timeBonusMultiplier = utils.GetFloat64AtPath(data, "time_bonus_multiplier")
	vipData.megaTimeBonusMultiplier = utils.GetFloat64AtPath(data, "mega_time_bonus_multiplier")
	vipData.leaderboardRewardMultiplier = utils.GetFloat64AtPath(data, "leaderboard_reward_multiplier")
	vipData.purchaseMultiplier = utils.GetFloat64AtPath(data, "purchase_multiplier")
	return nil
}

func (player *Player) increaseVipScore(score int64) (newVipScore int64, currentVipCode string, err error) {
	newVipScore = player.vipScore + score
	var biggestValidScore int64
	var newVipData *VipData
	for _, vipData := range vipDataList {
		if vipData.requirementScore <= newVipScore {
			if vipData.requirementScore >= biggestValidScore {
				newVipData = vipData
				biggestValidScore = vipData.requirementScore
			}
		}
	}
	newVipCode := newVipData.code
	// insert to vip record
	queryString := fmt.Sprintf("UPDATE %s SET vip_code = $1, vip_score = $2 WHERE player_id = $3", VipRecordDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, newVipCode, newVipScore, player.Id())
	if err != nil {
		log.LogSerious("err add vip score %s %s", err, queryString)
		return 0, "", err
	}
	var willNotify bool
	if player.vipCode != newVipCode {
		willNotify = true
		// notify change to feature depend on vip
	}

	player.vipScore = newVipScore
	player.vipCode = newVipCode
	player.notifyPlayerDataChange()
	if willNotify {
		player.notifyTimeBonusChange()
	}
	return newVipScore, newVipCode, nil
}

func (player *Player) createVipRecord() (err error) {
	queryString := fmt.Sprintf("INSERT INTO %s (player_id, vip_score, vip_code) VALUES ($1,$2,$3)", VipRecordDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, player.Id(), 0, "vip_1")
	if err != nil {
		log.LogSerious("error create new vip record %s", err)
		return err
	}
	player.vipCode = "vip_1"
	player.vipScore = 0
	return nil
}

func (player *Player) getVipTimeBonusMultiplier() float64 {
	if feature.IsVipAvailable() {
		return vipDataList[player.vipCode].timeBonusMultiplier
	}
	return 1
}

func (player *Player) getVipLeaderboardRewardMultiplier() float64 {
	if feature.IsVipAvailable() {
		return vipDataList[player.vipCode].leaderboardRewardMultiplier
	}
	return 1
}

func (player *Player) getVipPurchaseMultiplier() float64 {
	if feature.IsVipAvailable() {
		return vipDataList[player.vipCode].purchaseMultiplier
	}
	return 1
}

func (player *Player) getVipMegaTimeBonusMultiplier() float64 {
	if feature.IsVipAvailable() {
		return vipDataList[player.vipCode].megaTimeBonusMultiplier
	}
	return 1
}
