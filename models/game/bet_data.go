package game

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/utils"
	"html/template"
	"sort"
)

type BetEntryInterface interface {
	UpdateEntry(data map[string]interface{})
	ChipValues() []int64
	Max() int64
	Min() int64
	Step() int64
	Tax() float64
	ImageName() string
	OwnerThreshold() int64
	CheatCode() string
	SetCheatCode(cheatCode string)
	EnableBot() bool
	SerializedDataForAdmin() map[string]interface{}
	SerializedData() map[string]interface{}
	SerializedDataMinimal() map[string]interface{}
	GetHTMLForEditForm() *htmlutils.EditObject

	SetGame(gameInstance GameInterface)
	Game() GameInterface
	GameCode() string
	CurrencyType() string
}

type BetDataInterface interface {
	UpdateBetData(data []map[string]interface{})
	SerializedData() []map[string]interface{}
	SerializedDataForAdmin() []map[string]interface{}
	SerializedDataMinimal() []map[string]interface{}
	Entries() []BetEntryInterface
	SetEntries(entries []BetEntryInterface)
	GetEntry(minBet int64) BetEntryInterface
	AddEntry(entry BetEntryInterface)
	AddEntryByData(data map[string]interface{})
	DeleteEntry(minBet int64)
	GetHtmlForAdminDisplay() template.HTML
	GetHTMLForCreateForm() *htmlutils.EditObject

	Game() GameInterface
	GameCode() string
	CurrencyType() string
}

type ByMinBet []BetEntryInterface

func (a ByMinBet) Len() int      { return len(a) }
func (a ByMinBet) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByMinBet) Less(i, j int) bool {
	betEntryI := a[i]
	betEntryJ := a[j]
	return betEntryI.Min() < betEntryJ.Min()
}

type BetEntry struct {
	min            int64
	max            int64
	step           int64
	tax            float64
	ownerThreshold int64

	imageName string
	imageUrl  string

	chipValues []int64

	cheatCode string
	enableBot bool

	game GameInterface
}

func NewBetEntry(min int64,
	max int64,
	step int64,
	tax float64,
	ownerThreshold int64,
	imageName string,
	imageUrl string,
	chipValues []int64,
	enableBot bool,
	cheatCode string) *BetEntry {
	return &BetEntry{
		min:            min,
		max:            max,
		step:           step,
		tax:            tax,
		ownerThreshold: ownerThreshold,
		imageName:      imageName,
		imageUrl:       imageUrl,
		chipValues:     chipValues,
		cheatCode:      cheatCode,
		enableBot:      enableBot,
	}
}

func NewBetEntryFromData(data map[string]interface{}) *BetEntry {
	betEntry := &BetEntry{}
	betEntry.UpdateEntry(data)
	return betEntry
}

func (entry *BetEntry) UpdateEntry(data map[string]interface{}) {
	entry.min = utils.GetInt64AtPath(data, "min_bet")
	entry.max = utils.GetInt64AtPath(data, "max_bet")
	entry.tax = utils.GetFloat64AtPath(data, "tax")
	entry.chipValues = utils.GetInt64SliceAtPath(data, "chip_values")
	entry.ownerThreshold = utils.GetInt64AtPath(data, "owner_threshold")
	entry.step = utils.GetInt64AtPath(data, "step")
	entry.imageName = utils.GetStringAtPath(data, "image_name")
	entry.imageUrl = utils.GetStringAtPath(data, "image_url")
	entry.cheatCode = utils.GetStringAtPath(data, "cheat_code")
	entry.enableBot = utils.GetBoolAtPath(data, "enable_bot")
}

func (entry *BetEntry) ChipValues() []int64 {
	return entry.chipValues
}

func (entry *BetEntry) Max() int64 {
	return entry.max
}

func (entry *BetEntry) Min() int64 {
	return entry.min
}

func (entry *BetEntry) Step() int64 {
	return entry.step
}

func (entry *BetEntry) Tax() float64 {
	return entry.tax
}

func (entry *BetEntry) OwnerThreshold() int64 {
	return entry.ownerThreshold
}

func (entry *BetEntry) CheatCode() string {
	return entry.cheatCode
}

func (entry *BetEntry) ImageName() string {
	return entry.imageName
}

func (entry *BetEntry) SetCheatCode(cheatCode string) {
	entry.cheatCode = cheatCode
}

func (entry *BetEntry) EnableBot() bool {
	return entry.enableBot
}

func (entry *BetEntry) Game() GameInterface {
	return entry.game
}

func (entry *BetEntry) GameCode() string {
	return entry.game.GameCode()
}

func (entry *BetEntry) CurrencyType() string {
	return entry.game.CurrencyType()
}

func (entry *BetEntry) SetGame(gameInstance GameInterface) {
	entry.game = gameInstance
}

func (entry *BetEntry) SerializedDataForAdmin() map[string]interface{} {
	data := entry.SerializedData()
	data["cheat_code"] = entry.cheatCode
	data["enable_bot"] = entry.enableBot
	data["tax"] = entry.tax
	data["image_name"] = entry.imageName
	return data
}

func (entry *BetEntry) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["min_bet"] = entry.min
	data["max_bet"] = entry.max
	data["chip_values"] = entry.chipValues
	data["step"] = entry.step
	data["owner_threshold"] = entry.ownerThreshold
	return data
}

func (entry *BetEntry) SerializedDataMinimal() map[string]interface{} {
	data := make(map[string]interface{})
	data["min_bet"] = entry.min
	return data
}

func (betEntry *BetEntry) GetHTMLForEditForm() *htmlutils.EditObject {
	row1 := htmlutils.NewStringHiddenField("game_code", betEntry.GameCode())
	row2 := htmlutils.NewInt64HiddenField("min_bet_params", betEntry.Min())
	row3 := htmlutils.NewStringField("Cheat code", "cheat_code", "Cheat code", betEntry.CheatCode())
	row4 := htmlutils.NewRadioField("Enable bot", "enable_bot", fmt.Sprintf("%v", betEntry.EnableBot()), []string{"true", "false"})
	row5 := htmlutils.NewInt64Field("Min bet", "min_bet", "Min bet", betEntry.Min())
	row6 := htmlutils.NewInt64Field("Max bet", "max_bet", "Max bet", betEntry.Max())
	row7 := htmlutils.NewInt64Field("Step", "step", "Step", betEntry.Step())
	row8 := htmlutils.NewFloat64Field("Tax", "tax", "Tax", betEntry.Tax())
	row9 := htmlutils.NewInt64SliceField("Chip values", "chip_values", "Chip values", betEntry.ChipValues())
	row10 := htmlutils.NewImageRadioField("Image name",
		"image_name",
		betEntry.ImageName(),
		[]string{"macau.png", "atlantic.png", "sydney.png", "vegas.png", "paris.png", "monaco.png", "dubai.png", "singapore.png", "london.png", "phuquoc.png", "tokyo.png"})

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6, row7, row8, row9, row10},
		fmt.Sprintf("/admin/game/%s/bet_data/edit", betEntry.GameCode()))
	return editObject
}

type BetData struct {
	entries []BetEntryInterface
	game    GameInterface
}

func (betData *BetData) Game() GameInterface {
	return betData.game
}

func (betData *BetData) GameCode() string {
	return betData.game.GameCode()
}

func (betData *BetData) CurrencyType() string {
	return betData.game.CurrencyType()
}

func NewBetData(gameInstance GameInterface, entries []BetEntryInterface) *BetData {
	for _, entry := range entries {
		entry.SetGame(gameInstance)
	}
	return &BetData{
		game:    gameInstance,
		entries: entries,
	}
}

func NewBetDataFromData(gameInstance GameInterface, data []map[string]interface{}) *BetData {
	betData := &BetData{
		game:    gameInstance,
		entries: make([]BetEntryInterface, 0),
	}
	for _, entryData := range data {
		entry := NewBetEntryFromData(entryData)
		entry.SetGame(gameInstance)
		betData.entries = append(betData.entries, entry)
	}
	return betData
}

func (betData *BetData) UpdateBetData(data []map[string]interface{}) {
	newEntries := make([]BetEntryInterface, 0)
	for _, betEntryData := range data {
		min := utils.GetInt64AtPath(betEntryData, "min_bet")
		var didUpdate bool
		for _, entry := range betData.entries {
			if entry.Min() == min {
				entry.UpdateEntry(betEntryData)
				newEntries = append(newEntries, entry)
				didUpdate = true
				break
			}
		}

		if !didUpdate {
			// create new
			betEntry := NewBetEntryFromData(betEntryData)
			betEntry.SetGame(betData.game)
			newEntries = append(newEntries, betEntry)
		}
	}

	sort.Sort(ByMinBet(newEntries))
	betData.entries = newEntries
}

func (betData *BetData) SerializedData() []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, entry := range betData.entries {
		data = append(data, entry.SerializedData())
	}
	return data

}

func (betData *BetData) SerializedDataForAdmin() []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, entry := range betData.entries {
		data = append(data, entry.SerializedDataForAdmin())
	}
	return data
}

func (betData *BetData) SerializedDataMinimal() []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, entry := range betData.entries {
		data = append(data, entry.SerializedDataMinimal())
	}
	return data
}

func (betData *BetData) GetHtmlForAdminDisplay() template.HTML {
	gameCode := betData.game.GameCode()

	headers := []string{"MinBet", "MaxBet", "Step", "CheatCode", "EnableBot", "Image", "Action", ""}
	columns := make([][]*htmlutils.TableColumn, 0)
	for _, entry := range betData.Entries() {
		c1 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Min()))
		c2 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Max()))
		c3 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Step()))
		c4 := htmlutils.NewStringTableColumn(entry.CheatCode())
		c5 := htmlutils.NewStringTableColumn(fmt.Sprintf("%v", entry.EnableBot()))
		c6 := htmlutils.NewImageTableColumn(fmt.Sprintf("/images/%s", entry.ImageName()))
		c7 := htmlutils.NewActionTableColumn("primary",
			"Edit",
			fmt.Sprintf("/admin/game/%s/bet_data/edit?min_bet=%d", gameCode, entry.Min()))
		c8 := htmlutils.NewActionTableColumn("danger",
			"Delete",
			fmt.Sprintf("/admin/game/%s/bet_data/delete?min_bet=%d", gameCode, entry.Min()))

		row := []*htmlutils.TableColumn{c1, c2, c3, c4, c5, c6, c7, c8}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	return table.SerializedData()
}

func (betData *BetData) GetHTMLForCreateForm() *htmlutils.EditObject {
	gameCode := betData.game.GameCode()

	row1 := htmlutils.NewStringHiddenField("game_code", gameCode)
	row3 := htmlutils.NewStringField("Cheat code", "cheat_code", "Cheat code", "")
	row4 := htmlutils.NewRadioField("Enable bot", "enable_bot", "false", []string{"true", "false"})
	row5 := htmlutils.NewInt64Field("Min bet", "min_bet", "Min bet", 0)
	row6 := htmlutils.NewInt64Field("Max bet", "max_bet", "Max bet", 0)
	row7 := htmlutils.NewInt64Field("Step", "step", "Step", 0)
	row8 := htmlutils.NewInt64Field("Tax", "tax", "Tax", 0)
	row9 := htmlutils.NewInt64SliceField("Chip values", "chip_values", "Chip values", nil)
	row10 := htmlutils.NewImageRadioField("Image name",
		"image_name",
		"",
		[]string{"macau.png", "atlantic.png", "sydney.png", "vegas.png", "paris.png", "monaco.png", "dubai.png", "singapore.png", "london.png", "phuquoc.png", "tokyo.png"})

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row3, row4, row5, row6, row7, row8, row9, row10}, fmt.Sprint("/admin/game/%s/bet_data/add", gameCode))
	return editObject
}

func (betData *BetData) Entries() []BetEntryInterface {
	return betData.entries
}

func (betData *BetData) SetEntries(entries []BetEntryInterface) {
	for _, entry := range entries {
		entry.SetGame(betData.game)
	}
	betData.entries = entries
}

func (betData *BetData) GetEntry(minBet int64) BetEntryInterface {
	for _, entry := range betData.entries {
		if entry.Min() == minBet {
			return entry
		}
	}
	return nil
}

func (betData *BetData) AddEntry(entry BetEntryInterface) {
	entry.SetGame(betData.game)
	betData.entries = append(betData.entries, entry)
	sort.Sort(ByMinBet(betData.entries))
}

func (betData *BetData) AddEntryByData(data map[string]interface{}) {
	betData.AddEntry(NewBetEntryFromData(data))
}

func (betData *BetData) DeleteEntry(minBet int64) {
	newEntries := make([]BetEntryInterface, 0)
	for _, entry := range betData.entries {
		if entry.Min() != minBet {
			newEntries = append(newEntries, entry)
		}
	}
	betData.entries = newEntries
}
