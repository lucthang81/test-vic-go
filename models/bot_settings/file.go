package bot_settings

import (
	"fmt"
	"github.com/vic/vic_go/log"
	"io/ioutil"
)

func writeConfigToFile(data string) {
	configFilePath := fmt.Sprintf("%s/conf/game_config/bot_config.json", projectDirectory)
	err := ioutil.WriteFile(configFilePath, []byte(data), 0666)
	if err != nil {
		log.LogSerious("err create file %v", err)
		return
	}
}
