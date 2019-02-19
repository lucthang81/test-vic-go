package game_config

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/log"
	"io/ioutil"
)

func writeConfigToFile(data map[string]interface{}) {
	configFilePath := fmt.Sprintf("%s/conf/game_config/config.json", projectDirectory)
	// rawBytes, err := json.Marshal(data)
	rawBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.LogSerious("err make raw bytes from default config %v", err)
		return
	}
	err = ioutil.WriteFile(configFilePath, rawBytes, 0666)
	if err != nil {
		log.LogSerious("err create file %v", err)
		return
	}
}
