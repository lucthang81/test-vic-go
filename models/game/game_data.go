package game

type GameData struct {
	rooms    *Int64RoomMap
	helpText string
}

func NewGameData() *GameData {
	return &GameData{
		rooms: NewInt64RoomMap(),
	}
}

func (gameData *GameData) Rooms() *Int64RoomMap {
	return gameData.rooms
}

func (gameData *GameData) SetRooms(rooms *Int64RoomMap) {
	gameData.rooms = rooms
}

func (gameData *GameData) NumberOfOnlineRooms() int {
	counter := 0
	for _, room := range gameData.rooms.Copy() {
		if room.IsRoomOnline() {
			counter++
		}
	}
	return counter
}

func (gameData *GameData) HelpText() string {
	return gameData.helpText
}

func (gameData *GameData) SetHelpText(helpText string) {
	gameData.helpText = helpText
}

func (gameData *GameData) GetRoom(roomId int64) *Room {
	for _, room := range gameData.rooms.Copy() {
		if room.Id() == roomId {
			return room
		}
	}
	return nil
}
