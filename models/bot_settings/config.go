package bot_settings

import (
	"fmt"
	"io/ioutil"
)

var projectDirectory string

var fullData string

func GetConfigString() string {
	return fullData
}

func LoadGameConfig(theProjectDirectory string) {
	projectDirectory = theProjectDirectory
	configFilePath := fmt.Sprintf("%s/conf/game_config/bot_config.json", projectDirectory)

	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("err??", err)
		// missing file, just create new and put default value in
		contentString := loadDefault()
		writeConfigToFile(contentString)

		update(contentString)
	} else {

		update(string(content))
	}
}

func loadDefault() string {
	rawString := `
		{
	"bot_settings": [
{
			"currency_type": "money",
			"game_code": "bacay",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					1000,
					10000
				],
				[
					5000,
					50000
				],
				[
					10000,
					100000
				],
				[
					20000,
					200000
				],
				[
					50000,
					500000
				],
				[
					26000,
					520000
				],
				[
					130000,
					2600000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "bacay",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					40000,
					100000
				],
				[
					100000,
					500000
				],
				[
					400000,
					1000000
				],
				[
					800000,
					2000000
				],
				[
					500000,
					5000000
				],
				[
					260000,
					520000
				],
				[
					1300000,
					2600000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "baicao",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					1000,
					10000
				],
				[
					5000,
					50000
				],
				[
					10000,
					100000
				],
				[
					20000,
					200000
				],
				[
					50000,
					500000
				],
				[
					260000,
					520000
				],
				[
					130000,
					2600000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "baicao",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					40000,
					100000
				],
				[
					100000,
					500000
				],
				[
					400000,
					1000000
				],
				[
					800000,
					2000000
				],
				[
					500000,
					5000000
				],
				[
					260000,
					520000
				],
				[
					130000,
					2600000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "xidach",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					1000,
					10000
				],
				[
					5000,
					50000
				],
				[
					10000,
					100000
				],
				[
					20000,
					200000
				],
				[
					50000,
					500000
				],
				[
					1300000,
					2600000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "xidach",
			"number_of_bots":35,
			"leave_when_normal_count":3,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					10000,
					100000
				],
				[
					50000,
					500000
				],
				[
					100000,
					1000000
				],
				[
					200000,
					2000000
				],
				[
					500000,
					5000000
				],
				[
					130000,
					2600000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "tienlen_solo",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					3500,
					35000
				],
				[
					3500,
					35000
				],
				[
					14000,
					140000
				],
				[
					32000,
					300000
				],
				[
					80000,
					500000
				],
				[
					400000,
					800000
				],
				[
					2000000,
					4000000
				]
			]
		},
{
			"currency_type": "test_money",
			"game_code": "tienlen_solo",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					7000,
					70000
				],
				[
					35000,
					350000
				],
				[
					140000,
					1400000
				],
				[
					320000,
					3000000
				],
				[
					800000,
					5000000
				],
				[
					400000,
					800000
				],
				[
					2000000,
					4000000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "tienlen",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					7000,
					70000
				],
				[
					35000,
					350000
				],
				[
					140000,
					1400000
				],
				[
					320000,
					3000000
				],
				[
					800000,
					5000000
				],
				[
					260000,
					520000
				],
				[
					1300000,
					2600000
				]
			]
		},

		{
			"currency_type": "money",
			"game_code": "tienlen",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					3500,
					10000
				],
				[
					10000,
					35000
				],
				[
					40000,
					140000
				],
				[
					32000,
					300000
				],
				[
					80000,
					500000
				],
				[
					260000,
					520000
				],
				[
					1300000,
					2600000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "maubinh",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					3500,
					10000
				],
				[
					10000,
					35000
				],
				[
					40000,
					140000
				],
				[
					32000,
					300000
				],
				[
					80000,
					500000
				],
				[
					25000,
					500000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "maubinh",
			"number_of_bots":35,
			"leave_when_normal_count":2,
			"min_match_plays":5,
			"max_match_plays":10,
			"money_range": [
				[
					7000,
					70000
				],
				[
					35000,
					350000
				],
				[
					140000,
					1400000
				],
				[
					320000,
					3000000
				],
				[
					800000,
					5000000
				],
				[
					250000,
					5000000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "roulette",
			"number_of_bots":35,
			"leave_when_normal_count":4,
			"min_match_plays":6,
			"max_match_plays":10,
			"money_range": [
				[
					1000,
					40000
				],
				[
					10000,
					400000
				],
				[
					80000,
					2000000
				]
			]
		},
		{
			"currency_type": "test_money",
			"game_code": "roulette",
			"number_of_bots":35,
			"leave_when_normal_count":4,
			"min_match_plays":6,
			"max_match_plays":10,
			"money_range": [
				[
					10000,
					400000
				],
				[
					100000,
					4000000
				],
				[
					800000,
					20000000
				],
				[
					800000,
					20000000
				]
			]
		},
		{
			"currency_type": "money",
			"game_code": "xocdia",
			"number_of_bots":35,
			"leave_when_normal_count":4,
			"min_match_plays":6,
			"max_match_plays":10,
			"money_range": [
				[
					1000,
					30000
				],
				[
					5000,
					150000
				],
				[
					50000,
					1500000
				],
				[
					100000,
					7000000
				]

			]
		},
		{
			"currency_type": "test_money",
			"game_code": "xocdia",
			"number_of_bots":35,
			"leave_when_normal_count":4,
			"min_match_plays":6,
			"max_match_plays":10,
			"money_range": [
				[
					10000,
					300000
				],
				[
					50000,
					1500000
				],
				[
					500000,
					15000000
				],
				[
					1000000,
					70000000
				]
			]
		}
	]
}
	`

	return rawString
}

func update(updatedString string) {
	fullData = updatedString

}
