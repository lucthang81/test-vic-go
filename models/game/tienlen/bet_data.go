package tienlen

import (
	"fmt"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/utils"
	"html/template"
	"sort"
)

type BetEntry struct {
	game.BetEntryInterface

	ownerTax            float64
	numberOfSystemRooms int
}

func NewBetEntry(
	min int64,
	tax float64,
	ownerTax float64,
	numberOfSystemRooms int,
	imageName string,
	enableBot bool,
	cheatCode string) *BetEntry {
	entry := &BetEntry{
		BetEntryInterface: game.NewBetEntry(min, 0, 0, tax, 0, imageName, "", nil, enableBot, cheatCode),
	}
	entry.ownerTax = ownerTax
	entry.numberOfSystemRooms = numberOfSystemRooms
	return entry
}

func NewBetEntryFromData(data map[string]interface{}) *BetEntry {
	betEntry := &BetEntry{BetEntryInterface: game.NewBetEntryFromData(data)}
	betEntry.UpdateEntry(data)
	return betEntry
}

func (entry *BetEntry) UpdateEntry(data map[string]interface{}) {
	entry.BetEntryInterface.UpdateEntry(data)
	entry.ownerTax = utils.GetFloat64AtPath(data, "owner_tax")
	entry.numberOfSystemRooms = utils.GetIntAtPath(data, "number_of_system_rooms")
}

func (betEntry *BetEntry) GetHTMLForEditForm() *htmlutils.EditObject {
	row1 := htmlutils.NewStringHiddenField("game_code", betEntry.GameCode())
	row2 := htmlutils.NewStringHiddenField("currency_type", betEntry.CurrencyType())
	row3 := htmlutils.NewInt64Field("Number of System Rooms", "number_of_system_rooms", "Number of System Rooms", int64(betEntry.numberOfSystemRooms))
	row4 := htmlutils.NewInt64HiddenField("min_bet_params", betEntry.Min())
	row5 := htmlutils.NewStringField("Cheat code", "cheat_code", "Cheat code", betEntry.CheatCode())
	row6 := htmlutils.NewRadioField("Enable bot", "enable_bot", fmt.Sprintf("%v", betEntry.EnableBot()), []string{"true", "false"})
	row7 := htmlutils.NewInt64Field("Min bet", "min_bet", "Min bet", betEntry.Min())
	row8 := htmlutils.NewInt64Field("Max bet", "max_bet", "Max bet", betEntry.Max())
	row9 := htmlutils.NewInt64Field("Step", "step", "Step", betEntry.Step())
	row10 := htmlutils.NewFloat64Field("Tax", "tax", "Tax", betEntry.Tax())
	row11 := htmlutils.NewFloat64Field("Owner Tax", "owner_tax", "Owner Tax", betEntry.ownerTax)
	row12 := htmlutils.NewImageRadioField("Image name",
		"image_name",
		betEntry.ImageName(),
		[]string{"macau.png", "atlantic.png", "sydney.png", "vegas.png", "paris.png", "monaco.png", "dubai.png", "singapore.png", "london.png", "phuquoc.png", "tokyo.png"})

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6, row7, row8, row9, row10, row11, row12},
		fmt.Sprintf("/admin/game/%s/bet_data/edit", betEntry.GameCode()))
	return editObject
}

func (entry *BetEntry) SerializedDataForAdmin() map[string]interface{} {
	data := entry.BetEntryInterface.SerializedDataForAdmin()
	data["owner_tax"] = entry.ownerTax
	data["number_of_system_rooms"] = entry.numberOfSystemRooms
	data["cheat_code"] = entry.CheatCode()
	data["enable_bot"] = entry.EnableBot()
	return data
}

type BetData struct {
	game.BetDataInterface
}

func NewBetData(gameInstance game.GameInterface, entries []game.BetEntryInterface) *BetData {
	return &BetData{
		BetDataInterface: game.NewBetData(gameInstance, entries),
	}
}

func (betData *BetData) UpdateBetData(data []map[string]interface{}) {
	newEntries := make([]game.BetEntryInterface, 0)
	for _, betEntryData := range data {
		min := utils.GetInt64AtPath(betEntryData, "min_bet")
		var didUpdate bool
		for _, entry := range betData.Entries() {
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
			newEntries = append(newEntries, betEntry)
		}
	}

	sort.Sort(game.ByMinBet(newEntries))
	betData.BetDataInterface.SetEntries(newEntries)
}

func (betData *BetData) AddEntryByData(data map[string]interface{}) {
	entry := NewBetEntryFromData(data)
	betData.AddEntry(entry)
}

func (betData *BetData) GetHtmlForAdminDisplay() template.HTML {
	gameCode := betData.GameCode()
	currencyType := betData.CurrencyType()
	headers := []string{"MinBet", "MaxBet", "Step", "Tax", "Owner Tax", "Number of System Rooms", "CheatCode", "EnableBot", "Image", "Action", ""}
	columns := make([][]*htmlutils.TableColumn, 0)
	for _, entry := range betData.Entries() {
		betEntry := entry.(*BetEntry)
		c1 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Min()))
		c2 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Max()))
		c3 := htmlutils.NewStringTableColumn(utils.FormatWithComma(entry.Step()))
		c4 := htmlutils.NewStringTableColumn(fmt.Sprintf("%.2f", entry.Tax()))
		c5 := htmlutils.NewStringTableColumn(fmt.Sprintf("%.2f", betEntry.ownerTax))
		c6 := htmlutils.NewStringTableColumn(fmt.Sprintf("%d", betEntry.numberOfSystemRooms))
		c7 := htmlutils.NewStringTableColumn(entry.CheatCode())
		c8 := htmlutils.NewStringTableColumn(fmt.Sprintf("%v", entry.EnableBot()))
		c9 := htmlutils.NewImageTableColumn(fmt.Sprintf("/images/%s", entry.ImageName()))
		c10 := htmlutils.NewActionTableColumn("primary",
			"Edit",
			fmt.Sprintf("/admin/game/%s/bet_data/edit?min_bet=%d&currency_type=%s", gameCode, entry.Min(), currencyType))
		c11 := htmlutils.NewActionTableColumn("danger",
			"Delete",
			fmt.Sprintf("/admin/game/%s/bet_data/delete?min_bet=%d&currency_type=%s", gameCode, entry.Min(), currencyType))

		row := []*htmlutils.TableColumn{c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11}
		columns = append(columns, row)
	}
	table := htmlutils.NewTableObject(headers, columns)
	return table.SerializedData()
}

func (betData *BetData) GetHTMLForCreateForm() *htmlutils.EditObject {
	gameCode := betData.GameCode()
	row1 := htmlutils.NewStringHiddenField("game_code", gameCode)
	row2 := htmlutils.NewStringHiddenField("currency_type", betData.CurrencyType())
	row3 := htmlutils.NewStringField("Cheat code", "cheat_code", "Cheat code", "")
	row4 := htmlutils.NewRadioField("Enable bot", "enable_bot", "false", []string{"true", "false"})
	row5 := htmlutils.NewInt64Field("Number of system rooms", "number_of_system_rooms", "Number of system rooms", 0)
	row6 := htmlutils.NewInt64Field("Min bet", "min_bet", "Min bet", 0)
	row7 := htmlutils.NewInt64Field("Max bet", "max_bet", "Max bet", 0)
	row8 := htmlutils.NewInt64Field("Step", "step", "Step", 0)
	row9 := htmlutils.NewFloat64Field("Tax", "tax", "Tax", 0)
	row10 := htmlutils.NewFloat64Field("Owner Tax", "owner_tax", "Owner Tax", 0)
	row11 := htmlutils.NewImageRadioField("Image name",
		"image_name",
		"",
		[]string{"macau.png", "atlantic.png", "sydney.png", "vegas.png", "paris.png", "monaco.png", "dubai.png", "singapore.png", "london.png", "phuquoc.png", "tokyo.png"})

	editObject := htmlutils.NewEditObject([]*htmlutils.EditEntry{row1, row2, row3, row4, row5, row6, row7, row8, row9, row10, row11},
		fmt.Sprintf("/admin/game/%s/bet_data/add", gameCode))
	return editObject
}

func (betData *BetData) SerializedDataForAdmin() []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for _, entry := range betData.Entries() {
		betEntry := entry.(*BetEntry)
		data = append(data, betEntry.SerializedDataForAdmin())
	}
	return data
}
