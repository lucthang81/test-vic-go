package models

import (
	"github.com/vic/vic_go/models/game_config"
	"github.com/vic/vic_go/utils"
)

func (models *Models) IsAlreadyRegisterUsingThisIpAddress(ipAddress string) bool {
	return models.registerIpAddressMap[ipAddress]
}

func (models *Models) RegisterIpAddress(ipAddress string) {
	models.registerIpAddressMap[ipAddress] = true
}

func (models *Models) actuallyStartScheduleToCleanIpAddressMap() {
	defer models.startScheduleToCleanIpAddressMap()

	models.registerIpAddressMap = make(map[string]bool)
	timeout := utils.NewTimeOut(game_config.RegisterAgainMinDuration())
	timeout.Start()
}
