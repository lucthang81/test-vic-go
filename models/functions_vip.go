package models

import (
	"github.com/vic/vic_go/models/player"
)

func getVipDataList(models *Models, data map[string]interface{}, playerId int64) (responseData map[string]interface{}, err error) {
	responseData = make(map[string]interface{})
	responseData["vip_data_list"] = player.GetVipDataList()
	return responseData, nil
}
