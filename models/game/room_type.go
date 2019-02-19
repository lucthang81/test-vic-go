package game

type RoomType struct {
	code   string
	minBet int64
	maxBet int64

	chipValues []int64
}

func (roomType *RoomType) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	data["code"] = roomType.code
	data["min_bet"] = roomType.minBet
	data["max_bet"] = roomType.maxBet
	data["chip_values"] = roomType.chipValues
	return data
}
func GetRoomTypeData(roomTypeList []*RoomType) (data []map[string]interface{}) {
	data = make([]map[string]interface{}, 0)
	for _, roomType := range roomTypeList {
		data = append(data, roomType.SerializedData())
	}
	return data
}

func GetRoomType(roomTypeList []*RoomType, code string) *RoomType {
	for _, roomType := range roomTypeList {
		if roomType.code == code {
			return roomType
		}
	}
	return nil
}
