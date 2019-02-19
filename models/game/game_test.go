package game

import (
	"errors"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/details_error"
	"github.com/vic/vic_go/htmlutils"
	"github.com/vic/vic_go/models/congrat_queue"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"

	// "github.com/vic/vic_go/log"
	"github.com/vic/vic_go/test"
	// "github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter      *datacenter.DataCenter
	server          *TestServer
	dbName          string
	playerIdCounter int64
}

var _ = Suite(&TestSuite{
	dbName: "casino_game_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../sql/init_schema.sql", "../../sql/test_schema/player_test.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	RegisterDataCenter(s.dataCenter)
	currency.RegisterDataCenter(s.dataCenter)
	record.RegisterDataCenter(s.dataCenter)
	congrat_queue.LoadCongratQueue()
	s.server = NewTestServer()
	RegisterServer(s.server)
}

func (s *TestSuite) TearDownSuite(c *C) {
	s.dataCenter.Db().Close()
	test.DropTestDatabase(s.dbName)
	s.dataCenter.FlushCache()

}

func (s *TestSuite) SetUpTest(c *C) {
	// Use s.dir to prepare some data.

	fmt.Printf("start test %s \n", c.TestName())
}

func (s *TestSuite) TearDownTest(c *C) {
}

/*



THE ACTUAL TESTS




*/

func (s *TestSuite) TestRunnable(c *C) {
	c.Assert(false, Equals, false)
}

func (s *TestSuite) Test1Money(c *C) {
	betEntry := &BetEntry{
		tax: 0.3,
	}

	moneyAfter := MoneyAfterTax(1, betEntry)
	c.Assert(moneyAfter, Equals, int64(1))

	tax := TaxFromMoney(1, betEntry)
	c.Assert(tax, Equals, int64(0))
}

func (s *TestSuite) TestCreateRoom(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	var room *Room
	var err error

	testGame := NewTestGame(currencyType)
	_, err = CreateRoom(testGame, player, 1000, 4, "")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")

	_, err = CreateRoom(testGame, player, 100, 10, "")
	c.Assert(err.Error(), Equals, "err:invalid_max_number_of_players_in_room")

	room, err = CreateRoom(testGame, player, 10, 0, "")
	c.Assert(err, IsNil)
	c.Assert(room.MaxNumberOfPlayers(), Equals, 4)
	c.Assert(room.Owner().Id(), Equals, player.id)
	c.Assert(len(room.Players().coreMap), Equals, 1)
	data := room.SerializedData()
	c.Assert(utils.GetIntAtPath(data, "number_of_players"), Equals, 1)
	c.Assert(utils.GetStringAtPath(room.SerializedData(), "game_code"), Equals, testGame.gameCode)

	player1 := s.newPlayer()
	player1.setMoney(50, currencyType)
	testGame.requirementMultiplier = 10
	_, err = CreateRoom(testGame, player1, 10, 0, "")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")
}

func (s *TestSuite) TestJoinCreateRoomQuick(c *C) {
	currencyType := currency.TestMoney

	player1 := s.newPlayer()
	testGame := NewTestGame(currencyType)
	testGame.roomType = RoomTypeQuick
	room, err := QuickJoinRoom(testGame, player1)
	c.Assert(err, IsNil)
	c.Assert(room, NotNil)

	fmt.Println("111")

	player2 := s.newPlayer()
	player2.setMoney(2, currencyType)
	room2, err := QuickJoinRoom(testGame, player2)
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")
	c.Assert(room2, IsNil)

	fmt.Println("222")

	player3 := s.newPlayer()
	room3, err := QuickJoinRoom(testGame, player3)
	c.Assert(err, IsNil)
	c.Assert(room3, NotNil)

	fmt.Println("3333")

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)
	room4, err := JoinRoomByRequirement(testGame, player4, 10000)
	c.Assert(err, IsNil)
	c.Assert(room4.Requirement(), Equals, int64(10000))

	fmt.Println("4444")

	_, err = JoinRoomByRequirement(testGame, player4, 1000)
	c.Assert(err.Error(), Equals, l.Get(l.M0038))

	player6 := s.newPlayer()
	player6.setMoney(1000, currencyType)
	_, err = JoinRoomByRequirement(testGame, player6, 10000)
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")

}

func (s *TestSuite) TestQuickTypeRoomPriority(c *C) {
	currencyType := currency.TestMoney

	testGame := NewTestGame(currencyType)
	testGame.roomType = RoomTypeQuick
	players := make([]*TestPlayer, 0)
	for i := 0; i < 30; i++ {
		player := s.newPlayer()
		player.playerType = "bot"

		room, err := JoinRoomByRequirement(testGame, player, 100)
		c.Assert(err, IsNil)
		c.Assert(room, NotNil)
		players = append(players, player)
	}

	normal1 := s.newPlayer()
	normal1.playerType = "normal"
	normal2 := s.newPlayer()
	normal2.playerType = "normal"

	room, err := JoinRoomByRequirement(testGame, normal1, 100)
	c.Assert(err, IsNil)
	c.Assert(room, NotNil)

	room2, err := JoinRoomByRequirement(testGame, normal2, 100)
	c.Assert(err, IsNil)
	c.Assert(room2, NotNil)
	c.Assert(room.id, Equals, room2.id)
}

func (s *TestSuite) TestRegisterUnregisterLeaveNoMorePlayer(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")
	JoinRoomById(testGame, player3, room.Id(), "")
	c.Assert(len(room.Players().coreMap), Equals, 4)

	c.Assert(room.GetMoneyOnTable(player.id), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player1.id), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.id), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player3.id), Equals, int64(100))

	_, err := room.StartGame(player)
	c.Assert(err, IsNil)
	c.Assert(room.Session(), NotNil)

	err = RegisterLeaveRoom(testGame, player)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 4)
	c.Assert(len(room.registerLeaveRoom), Equals, 1)

	err = RegisterLeaveRoom(testGame, player1)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 4)
	c.Assert(len(room.registerLeaveRoom), Equals, 2)

	// just free and then remove it
	player2.FreezeMoney(10000, currencyType, room.GetRoomIdentifierString(), true)
	room.DecreaseMoney(player2, 10000, true) // player2 run out of money
	c.Assert(room.GetTotalPlayerMoney(player2.id), Equals, int64(0))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(0))

	player3.FreezeMoney(10000, currencyType, room.GetRoomIdentifierString(), true)
	room.DecreaseMoney(player3, 10000, true) // player2 run out of money
	c.Assert(room.GetTotalPlayerMoney(player3.id), Equals, int64(0))
	c.Assert(player3.GetAvailableMoney(currencyType), Equals, int64(0))

	room.DidEndGame(room.Session().ResultSerializedData(), 0)
	c.Assert(len(room.Players().coreMap), Equals, 0)
	c.Assert(len(room.registerLeaveRoom), Equals, 0)
}

func (s *TestSuite) TestHelperMethod(c *C) {
	currencyType := currency.Money
	betEntry := &BetEntry{
		tax: 0.05,
	}
	testGame := NewTestGame(currencyType)
	testGame.leavePenalty = 2.5
	c.Assert(MoneyAfterTax(1000, betEntry), Equals, int64(1000*(1-betEntry.Tax())))
}

func (s *TestSuite) TestRoomFunction(c *C) {
	currencyType := currency.Money
	player := s.newPlayer()
	player.setMoney(100, currencyType)

	// player1 := s.newPlayer()
	// player1.GetAvailableMoney(currencyType) = 10000

	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()
	player5 := s.newPlayer()
	player6 := s.newPlayer()
	player7 := s.newPlayer()
	player8 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	var roomData []map[string]interface{}
	var err error

	testGame := NewTestGame(currencyType)
	fmt.Println("test in test", testGame.RequirementMultiplier())
	room1 := s.justCreateRoomWithOwnerName(testGame, "a", currencyType)
	room1.requirement = 100
	room2 := s.justCreateRoomWithOwnerName(testGame, "d", currencyType)
	room2.requirement = 80
	room3 := s.justCreateRoomWithOwnerName(testGame, "b", currencyType)
	room3.requirement = 90
	room4 := s.justCreateRoomWithOwnerName(testGame, "c", currencyType)
	room4.requirement = 70

	roomData, _ = GetRoomList(testGame, "", 10, 0)
	c.Assert(len(roomData), Equals, 4)

	// test offline
	room1.HandlePlayerOffline(room1.owner)
	roomData, _ = GetRoomList(testGame, "", 10, 0)
	c.Assert(len(roomData), Equals, 3)

	// test get room
	room1.HandlePlayerOnline(room1.owner)
	roomData, _ = GetRoomList(testGame, "", 10, 0)
	c.Assert(len(roomData), Equals, 4)

	roomData, _ = GetRoomList(testGame, "requirement", 10, 0)
	c.Assert(len(roomData), Equals, 4)

	var requirement int64

	for _, room := range roomData {
		if utils.GetInt64AtPath(room, "requirement") > requirement {
			requirement = utils.GetInt64AtPath(room, "requirement")
		} else {
			c.Assert(false, Equals, true)
		}
	}

	roomData, _ = GetRoomList(testGame, "-requirement", 10, 0)
	c.Assert(len(roomData), Equals, 4)

	requirement = 10000
	for _, room := range roomData {
		if utils.GetInt64AtPath(room, "requirement") < requirement {
			requirement = utils.GetInt64AtPath(room, "requirement")
		} else {
			c.Assert(false, Equals, true)
		}
	}

	roomData, _ = GetRoomList(testGame, "-owner", 10, 0)
	c.Assert(len(roomData), Equals, 4)
	fmt.Println(roomData)
	c.Assert(utils.GetStringAtPath(roomData[0], "owner/username"), Equals, "d")
	c.Assert(utils.GetStringAtPath(roomData[1], "owner/username"), Equals, "c")
	c.Assert(utils.GetStringAtPath(roomData[2], "owner/username"), Equals, "b")
	c.Assert(utils.GetStringAtPath(roomData[3], "owner/username"), Equals, "a")

	roomData, _ = GetRoomList(testGame, "owner", 10, 0)
	c.Assert(len(roomData), Equals, 4)
	c.Assert(utils.GetStringAtPath(roomData[0], "owner/username"), Equals, "a")
	c.Assert(utils.GetStringAtPath(roomData[1], "owner/username"), Equals, "b")
	c.Assert(utils.GetStringAtPath(roomData[2], "owner/username"), Equals, "c")
	c.Assert(utils.GetStringAtPath(roomData[3], "owner/username"), Equals, "d")

	// num players
	// room1 4/4
	JoinRoomById(testGame, player3, room1.id, "")
	JoinRoomById(testGame, player4, room1.id, "")
	JoinRoomById(testGame, player5, room1.id, "")

	_, err = JoinRoomById(testGame, player6, room1.id, "")
	c.Assert(err.Error(), Equals, l.Get(l.M0039))
	_, err = JoinRoomById(testGame, player7, room1.id, "")
	c.Assert(err.Error(), Equals, l.Get(l.M0039))

	// room2 3/4
	JoinRoomById(testGame, player6, room2.id, "")
	JoinRoomById(testGame, player7, room2.id, "")
	// room3 1/4
	//room 4 2/4
	JoinRoomById(testGame, player8, room4.id, "")

	roomData, _ = GetRoomList(testGame, "-numPlayers", 10, 0)
	c.Assert(len(roomData), Equals, 4)
	fmt.Println(utils.GetIntAtPath(roomData[0], "max_number_of_players") - utils.GetIntAtPath(roomData[0], "number_of_players"))
	fmt.Println(utils.GetIntAtPath(roomData[1], "max_number_of_players") - utils.GetIntAtPath(roomData[1], "number_of_players"))
	fmt.Println(utils.GetIntAtPath(roomData[2], "max_number_of_players") - utils.GetIntAtPath(roomData[2], "number_of_players"))
	fmt.Println(utils.GetIntAtPath(roomData[3], "max_number_of_players") - utils.GetIntAtPath(roomData[3], "number_of_players"))
	c.Assert(utils.GetInt64AtPath(roomData[0], "id"), Equals, room2.id)
	c.Assert(utils.GetInt64AtPath(roomData[1], "id"), Equals, room4.id)
	c.Assert(utils.GetInt64AtPath(roomData[2], "id"), Equals, room3.id)
	c.Assert(utils.GetInt64AtPath(roomData[3], "id"), Equals, room1.id)

	roomData, _ = GetRoomList(testGame, "", 10, 1)
	c.Assert(len(roomData), Equals, 3)

	roomData, _ = GetRoomList(testGame, "", 10, 2)
	c.Assert(len(roomData), Equals, 2)

	roomData, _ = GetRoomList(testGame, "", 10, 4)
	c.Assert(len(roomData), Equals, 0)

	roomData, _ = GetRoomList(testGame, "", 10, 3)
	c.Assert(len(roomData), Equals, 1)

	roomData, _ = GetRoomList(testGame, "", 10, 2)
	c.Assert(len(roomData), Equals, 2)

	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	s.justCreateRoom(testGame, currencyType)
	// 12 rooms

	roomData, _ = GetRoomList(testGame, "", 10, 2)
	c.Assert(len(roomData), Equals, 10)

	roomData, _ = GetRoomList(testGame, "", 8, 2)
	c.Assert(len(roomData), Equals, 8)

	roomData, _ = GetRoomList(testGame, "", 8, 6)
	c.Assert(len(roomData), Equals, 6)

	for i := 0; i < 50; i++ {
		s.justCreateRoom(testGame, currencyType)
	}

	_, err = JoinRoomById(testGame, player2, int64(1000), "")
	c.Assert(err.Error(), Equals, l.Get(l.M0092))

	room, _ := CreateRoom(testGame, player2, 1000, 9, "")
	c.Assert(room, NotNil)

	_, err = JoinRoomById(testGame, player, room.Id(), "")
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")

	player.setMoney(1000, currencyType)
	testGame.requirementMultiplier = 2
	_, err = JoinRoomById(testGame, player, room.Id(), "")
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")

}

func (s *TestSuite) TestAddRemoveKeepOrder(c *C) {
	currencyType := currency.TestMoney

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()
	player5 := s.newPlayer()
	player6 := s.newPlayer()
	player7 := s.newPlayer()

	// var roomData []map[string]interface{}
	// var err error

	testGame := NewTestGame(currencyType)

	room, _ := CreateRoom(testGame, player1, 1000, 9, "")
	c.Assert(room, NotNil)

	c.Assert(room.players.get(0).Id(), Equals, player1.id)

	JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)

	JoinRoomById(testGame, player3, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player3.id)

	JoinRoomById(testGame, player4, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player3.id)
	c.Assert(room.players.get(3).Id(), Equals, player4.id)

	RegisterLeaveRoom(testGame, player3)
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(3).Id(), Equals, player4.id)

	JoinRoomById(testGame, player5, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	fmt.Println(room.players)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)
	c.Assert(room.players.get(3).Id(), Equals, player4.id)

	RegisterLeaveRoom(testGame, player4)
	c.Assert(room.players.get(0).Id(), Equals, player1.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)

	RegisterLeaveRoom(testGame, player1)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)

	JoinRoomById(testGame, player6, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player6.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)

	JoinRoomById(testGame, player7, room.Id(), "")
	c.Assert(room.players.get(0).Id(), Equals, player6.id)
	c.Assert(room.players.get(1).Id(), Equals, player2.id)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)
	c.Assert(room.players.get(3).Id(), Equals, player7.id)

	RegisterLeaveRoom(testGame, player2)
	c.Assert(room.players.get(0).Id(), Equals, player6.id)
	c.Assert(room.players.get(2).Id(), Equals, player5.id)
	c.Assert(room.players.get(3).Id(), Equals, player7.id)

	playersInOrder := room.getPlayersSliceInOrder()
	c.Assert(len(playersInOrder), Equals, 3)
	for index, player := range playersInOrder {
		if index == 0 {
			c.Assert(player.Id(), Equals, player6.Id())
		} else if index == 1 {
			c.Assert(player.Id(), Equals, player5.Id())
		} else if index == 2 {
			c.Assert(player.Id(), Equals, player7.Id())
		}
	}
}

func (s *TestSuite) TestJoinRoomWithPassword(c *C) {
	currencyType := currency.Money
	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "password")
	c.Assert(player.room.Id(), Equals, room.Id())

	room1, _ := QuickJoinRoom(testGame, player1)
	c.Assert(room1, NotNil)
	c.Assert(room1.Id() != room.Id(), Equals, true)

	room, err := JoinRoomById(testGame, player3, room.Id(), "")
	c.Assert(err.Error(), Equals, "err:invalid_password")
}

func (s *TestSuite) TestFindAndJoinRoom(c *C) {
	currencyType := currency.Money
	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "")
	c.Assert(player.room.Id(), Equals, room.Id())
	room, _ = QuickJoinRoom(testGame, player1)
	c.Assert(room, NotNil)
	room, _ = QuickJoinRoom(testGame, player2)
	c.Assert(room, NotNil)
	room, _ = QuickJoinRoom(testGame, player3)
	c.Assert(room, NotNil)
	noRoomHereSoCreateNew, _ := QuickJoinRoom(testGame, player4)
	c.Assert(noRoomHereSoCreateNew, NotNil)
	c.Assert(noRoomHereSoCreateNew.Id() != room.Id(), Equals, true)

	// currently in room but trying to join other room
	CreateRoom(testGame, player4, 10, 4, "")
	fmt.Println(room.Id())
	c.Assert(player.room.Id(), Equals, room.Id())
	_, err := QuickJoinRoom(testGame, player)
	c.Assert(err.Error(), Equals, l.Get(l.M0038))
	errVal, ok := err.(*details_error.DetailsError)
	c.Assert(ok, Equals, true)
	c.Assert(utils.GetInt64AtPath(errVal.Details(), "id"), Equals, room.Id())

	_, err = QuickJoinRoom(testGame, player1)
	c.Assert(err.Error(), Equals, l.Get(l.M0038))
}

func (s *TestSuite) TestResponseNewPlayerJoinRoom(c *C) {
	currencyType := currency.Money
	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "")
	s.server.cleanupAllResponse()
	_, err := JoinRoomById(testGame, player1, room.Id(), "")
	c.Assert(err, IsNil)
	var isValid bool
	for i := 0; i < 2; i++ {
		response := s.server.getAndRemoveResponse(player.Id())
		if utils.GetStringAtPath(response, "method") == "new_player_join_room" {
			c.Assert(utils.GetIntAtPath(response, "data/index"), Equals, 1)
			isValid = true
			break
		}
		fmt.Println(utils.GetStringAtPath(response, "method"))
	}
	c.Assert(isValid, Equals, true)

	s.server.cleanupAllResponse()
	_, err = JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(err, IsNil)

	for i := 0; i < 2; i++ {
		response := s.server.getAndRemoveResponse(player1.Id())
		if utils.GetStringAtPath(response, "method") == "new_player_join_room" {
			c.Assert(utils.GetStringAtPath(response, "method"), Equals, "new_player_join_room")
			c.Assert(utils.GetIntAtPath(response, "data/index"), Equals, 2)
			isValid = true
			break
		}
	}
	c.Assert(isValid, Equals, true)

	for i := 0; i < 2; i++ {
		response := s.server.getAndRemoveResponse(player.Id())
		if utils.GetStringAtPath(response, "method") == "new_player_join_room" {
			c.Assert(utils.GetStringAtPath(response, "method"), Equals, "new_player_join_room")
			c.Assert(utils.GetIntAtPath(response, "data/index"), Equals, 2)
			isValid = true
			break
		}
	}
	c.Assert(isValid, Equals, true)
}

func (s *TestSuite) TestInviteNewPlayerToRoom(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "")

	err := room.InvitePlayerToRoom(player1, player1)
	c.Assert(err.Error(), Equals, "err:player_not_in_room")

	err = room.InvitePlayerToRoom(player, player)
	c.Assert(err.Error(), Equals, "err:player_already_in_room")

	player2.setMoney(1, currencyType)
	err = room.InvitePlayerToRoom(player, player2)
	c.Assert(err.Error(), Equals, "err:requirement_not_meet")

	s.server.cleanupAllResponse()
	err = room.InvitePlayerToRoom(player, player1)
	c.Assert(err, IsNil)

	response := s.server.getAndRemoveResponse(player1.Id())
	c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_receive_invitation_to_room")
	c.Assert(utils.GetInt64AtPath(response, "data/room/id"), Equals, room.Id())

	player2.setMoney(1000, currencyType)
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	s.server.cleanupAllResponse()
	err = room.InvitePlayerToRoom(player1, player3)
	c.Assert(err, IsNil)
	response = s.server.getAndRemoveResponse(player3.Id())
	c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_receive_invitation_to_room")
	c.Assert(utils.GetInt64AtPath(response, "data/room/id"), Equals, room.Id())

	room.password = "dddd"
	s.server.cleanupAllResponse()
	err = room.InvitePlayerToRoom(player1, player3)
	c.Assert(err, IsNil)
	response = s.server.getAndRemoveResponse(player3.Id())
	c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_receive_invitation_to_room")
	c.Assert(utils.GetInt64AtPath(response, "data/room/id"), Equals, room.Id())
	c.Assert(utils.GetStringAtPath(response, "data/room/password"), Equals, "dddd")
	_, err = JoinRoomById(testGame, player3, room.Id(), "")
	c.Assert(err, NotNil)

	_, err = JoinRoomById(testGame, player3, room.Id(), room.password)
	c.Assert(err, IsNil)

	roomData := room.SerializedData()
	c.Assert(utils.GetStringAtPath(roomData, "password"), Equals, "")
	room.password = ""

	// full
	_, err = JoinRoomById(testGame, player4, room.Id(), "")
	c.Assert(err.Error(), Equals, l.Get(l.M0039))

	err = room.InvitePlayerToRoom(player1, player4)
	c.Assert(err.Error(), Equals, l.Get(l.M0039))

}

func (s *TestSuite) TestReadyRoom(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "")
	_, err := room.StartGame(player)
	c.Assert(err.Error(), Equals, "err:not_enough_player")

	JoinRoomById(testGame, player1, room.Id(), "")
	_, err = room.StartGame(player1)
	c.Assert(err.Error(), Equals, "err:no_permission")
	s.server.cleanupAllResponse()
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(9990)) // move the money to the table
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(10))

	_, err = room.StartGame(player)
	c.Assert(err, IsNil)
	c.Assert(room.IsPlaying(), Equals, true)
	c.Assert(len((room.session.(*TestGameSession)).players), Equals, 2)

	JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9990)) // move the money to the table
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(10))

	room.DidEndGame(room.Session().ResultSerializedData(), 0)
	s.server.cleanupAllResponse()

	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9990)) // move the money to the table
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(10))

	RegisterLeaveRoom(testGame, player1)

	s.server.cleanupAllResponse()
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9990)) // already deduce when join room, mark ready not ready will not change this anymore
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(10))

	err = room.KickPlayer(player2, player)
	c.Assert(err.Error(), Equals, "err:no_permission")

	err = room.KickPlayer(player, player)
	c.Assert(err.Error(), Equals, "err:cannot_kick_owner")

	_, err = room.StartGame(player)
	c.Assert(err, IsNil)

	err = room.KickPlayer(player, player2)
	c.Assert(err.Error(), Equals, "err:already_start_playing")

	JoinRoomById(testGame, player1, room.Id(), "")

	room.DidEndGame(room.Session().ResultSerializedData(), 0)
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9990)) // game end, nothing change if money on table not touch
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(10))
	err = room.KickPlayer(player, player2)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 2)

	for _, player := range room.players.copy() {
		c.Assert(player.Id() != player2.Id(), Equals, true)
	}

	// test got kick and got money on table back
	JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9990)) // move the money to the table, same amount, so money did not change
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(10))
	room.KickPlayer(player, player2)
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(10000)) // move the money back
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(0))
}

func (s *TestSuite) TestReadyRoomAndAutoStart(c *C) {
	currencyType := currency.Money
	player := s.newPlayer()
	player.setMoney(100000, currencyType)

	player1 := s.newPlayer()
	player1.GetAvailableMoney(currencyType)

	player2 := s.newPlayer()
	player2.GetAvailableMoney(currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 10, 4, "")
	_, err := room.StartGame(player)
	c.Assert(err.Error(), Equals, "err:not_enough_player")

	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")
	s.server.cleanupAllResponse()
	utils.DelayInDuration(10 * time.Millisecond)
	c.Assert(room.getTimeUntilsAutoStartGame() != 0, Equals, true)

}

func (s *TestSuite) TestCreateAndLeaveRightAway(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100000, currencyType)

	testGame := NewTestGameWithMultiplier(currencyType)
	testGame.isPlayedAgainstOwner = false
	testGame.moneyOnTableMultiplier = 4
	CreateRoom(testGame, player, 10, 4, "")
	RegisterLeaveRoom(testGame, player)

	utils.DelayInDuration(10 * time.Second)
	c.Assert(true, Equals, true)
}

func (s *TestSuite) TestMoneyOnTableMultiplier(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGameWithMultiplier(currencyType)
	testGame.isPlayedAgainstOwner = false
	testGame.moneyOnTableMultiplier = 4
	room, _ := CreateRoom(testGame, player, 10, 4, "")

	JoinRoomById(testGame, player1, room.Id(), "")

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(9960)) // move the money to the table
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(40))

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(9960)) // move the money to the table
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(40))

	room.KickPlayer(player, player1)
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(10000))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(0))
}

func (s *TestSuite) TestKickFromRoomIfNotEnoughMoney(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(400, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	_, err := room.StartGame(player)
	c.Assert(err.Error(), Equals, "err:not_enough_player")

	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	_, err = room.StartGame(player)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	// manually remove player money
	player2.FreezeMoney(10000, currencyType, room.GetRoomIdentifierString(), true)
	room.DecreaseMoney(player2, 10000, true)

	room.DidEndGame(room.session.ResultSerializedData(), 0)
	c.Assert(len(room.Players().coreMap), Equals, 2)

	_, err = room.StartGame(player)
	c.Assert(err, IsNil)
	// manually remove player money (he has 400 at the money, will now remove everything)
	room.DecreaseMoney(player, 100, true)
	player.DecreaseMoney(300, currencyType, true)
	room.DidEndGame(room.session.ResultSerializedData(), 0)
	c.Assert(len(room.Players().coreMap), Equals, 1)
	fmt.Println(room.players.get(1).Id())
	fmt.Println(room.owner.Id())
	c.Assert(room.owner.Id(), Equals, player1.Id())
}

func (s *TestSuite) TestRegisterUnregisterLeave(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")

	_, err := room.StartGame(player)
	c.Assert(err, IsNil)
	c.Assert(room.Session(), NotNil)
	JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(len(room.Players().coreMap), Equals, 3)

	err = RegisterLeaveRoom(testGame, player2)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 2)
	c.Assert(len(room.registerLeaveRoom), Equals, 0)

	err = RegisterLeaveRoom(testGame, player)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 2)
	c.Assert(len(room.registerLeaveRoom), Equals, 1)

	room.DidEndGame(room.Session().ResultSerializedData(), 0)
	c.Assert(len(room.Players().coreMap), Equals, 1)
	c.Assert(len(room.registerLeaveRoom), Equals, 0)

	JoinRoomById(testGame, player3, room.Id(), "")
	JoinRoomById(testGame, player4, room.Id(), "")
	_, err = room.StartGame(player1)
	c.Assert(err, IsNil)

	gameSession := room.Session().(*TestGameSession)

	err = RegisterLeaveRoom(testGame, player3)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 3)
	c.Assert(len(room.registerLeaveRoom), Equals, 1)

	gameSession.isPlaying = false
	err = RegisterLeaveRoom(testGame, player4)
	c.Assert(err, IsNil)
	c.Assert(len(room.Players().coreMap), Equals, 2)
	c.Assert(len(room.registerLeaveRoom), Equals, 1)
}

func (s *TestSuite) TestAlwaysHasOwner(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(4000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(2000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(2000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(2000, currencyType)

	testGame1 := NewTestGame(currencyType)
	fmt.Println("p", player.Id(), player1.Id(), player2.Id(), player3.Id(), player4.Id())
	testGame1.ownerMultiplier = 30

	room, err := CreateSystemRoom(testGame1, 100, 5, "")
	c.Assert(err, IsNil)
	c.Assert(room, NotNil)

	_, err = JoinRoomById(testGame1, player, room.Id(), "")
	c.Assert(err, NotNil)

	c.Assert(room.Owner(), IsNil)

	_, err = JoinRoomById(testGame1, player1, room.Id(), "")
	c.Assert(err, IsNil)

	c.Assert(room.Owner().Id(), Equals, player1.Id())

	_, err = JoinRoomById(testGame1, player, room.Id(), "")
	c.Assert(err, IsNil)

	_, err = JoinRoomById(testGame1, player2, room.Id(), "")
	c.Assert(err, IsNil)

	_, err = JoinRoomById(testGame1, player3, room.Id(), "")
	c.Assert(err, IsNil)

	_, err = JoinRoomById(testGame1, player4, room.Id(), "")
	c.Assert(err, IsNil)

	// remove player1 money and end game, no one will be able to be owner next round, so kick all
	_, err = room.StartGame(player1)
	c.Assert(err, IsNil)

	room.DecreaseMoney(player1, 3000, true)

	room.DidEndGame(nil, 0)
	utils.DelayInDuration(100 * time.Millisecond)
	c.Assert(room.players.Len(), Equals, 0)
	c.Assert(room.Owner(), IsNil)
}

func (s *TestSuite) TestRegisterUnregisterOwner(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(10000, currencyType)

	player4 := s.newPlayer()
	player4.setMoney(10000, currencyType)

	testGame1 := NewTestGame(currencyType)

	room, _ := CreateRoom(testGame1, player, 100, 4, "")
	JoinRoomById(testGame1, player1, room.Id(), "")
	err := room.RegisterToBeOwner(player1)
	c.Assert(err.Error(), Equals, "err:not_implemented")
	RegisterLeaveRoom(testGame1, player)
	RegisterLeaveRoom(testGame1, player1)

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(10000))

	testGame := NewTestGame(currencyType)
	testGame.properties = []string{GamePropertyRegisterOwner}
	room, _ = CreateSystemRoom(testGame, 100, 4, "")
	JoinRoomById(testGame, player, room.Id(), "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(len(room.Players().coreMap), Equals, 3)

	err = room.RegisterToBeOwner(player)
	c.Assert(err, IsNil)
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(0))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(100))

	session, err := testGame.StartGame(room, player, []GamePlayer{player, player1, player2}, 100, nil, nil)
	room.session = session

	c.Assert(room.IsPlaying(), Equals, true)
	err = room.UnregisterToBeOwner(player)
	c.Assert(err, IsNil)
	c.Assert(room.owner, NotNil)
	c.Assert(room.owner.Id(), Equals, player.Id())
	c.Assert(len(room.registerToBeOwner), Equals, 0)
	c.Assert(room.willNotBeOwnerNextRound, Equals, true)

	err = room.RegisterToBeOwner(player)
	c.Assert(len(room.registerToBeOwner), Equals, 0)
	c.Assert(room.willNotBeOwnerNextRound, Equals, false)

	err = room.RegisterToBeOwner(player1)
	c.Assert(len(room.registerToBeOwner), Equals, 1)
	c.Assert(room.willNotBeOwnerNextRound, Equals, false)

	err = room.UnregisterToBeOwner(player)
	c.Assert(err, IsNil)
	c.Assert(len(room.registerToBeOwner), Equals, 1)
	c.Assert(room.willNotBeOwnerNextRound, Equals, true)

	err = room.RegisterToBeOwner(player)
	c.Assert(len(room.registerToBeOwner), Equals, 1)
	c.Assert(room.willNotBeOwnerNextRound, Equals, false)

	err = room.RegisterToBeOwner(player2)
	c.Assert(len(room.registerToBeOwner), Equals, 2)

	err = room.RegisterToBeOwner(player2)
	c.Assert(len(room.registerToBeOwner), Equals, 2)

	err = room.UnregisterToBeOwner(player1)
	c.Assert(len(room.registerToBeOwner), Equals, 1)

	err = room.UnregisterToBeOwner(player)
	c.Assert(len(room.registerToBeOwner), Equals, 1)

	room.DidEndGame(session.ResultSerializedData(), 0)
	c.Assert(room.owner.Id(), Equals, player2.id)
	c.Assert(room.GetTotalPlayerMoney(player2.id), Equals, int64(10000))
	c.Assert(room.GetTotalPlayerMoney(player1.id), Equals, int64(10000))
	c.Assert(room.GetTotalPlayerMoney(player.id), Equals, int64(100))
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(0))
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(9900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(9900))

}

func (s *TestSuite) TestRemoveOwnerAndOwnerBlankForCasino(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(1000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(1000, currencyType)

	testGame := NewTestGame(currencyType)
	testGame.properties = []string{GamePropertyRegisterOwner}
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))

	RegisterLeaveRoom(testGame, player)
	c.Assert(room.owner, IsNil)

}

func (s *TestSuite) TestEndGameAndCheckOwnerCasino(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(1000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(1000, currencyType)

	testGame := NewTestGameCasinoOwner(currencyType)
	testGame.becomeOwnerRequirement = 1000
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")
	gameSession := room.session.(*TestGameSession)
	gameSession.isPlaying = false
	c.Assert(room.IsPlaying(), Equals, false)
	err := room.RegisterToBeOwner(player)
	c.Assert(err, IsNil)
	c.Assert(room.owner, NotNil)
	gameSession.isPlaying = true

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(1000))
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(0))

	room.DecreaseMoney(player, 1000, true)
	room.DidEndGame(room.session.ResultSerializedData(), 0)

	c.Assert(len(room.Players().coreMap), Equals, 2)
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(0))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(0))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.Owner(), IsNil)
}

func (s *TestSuite) TestRemoveOwnerAndChangeToAnotherPlayer(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(1000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(1000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))

	room.removePlayer(player)
	newOwner := room.owner
	c.Assert(newOwner.Id() != player.Id(), Equals, true)
	c.Assert(newOwner.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(newOwner.Id()), Equals, int64(100))
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(1000))
}

func (s *TestSuite) TestRemoveOwnerTwoTimeAndChangeToAnotherPlayer(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(1000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(1000, currencyType)

	player3 := s.newPlayer()
	player3.setMoney(1000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")
	JoinRoomById(testGame, player3, room.Id(), "")

	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(player2.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))

	go RegisterLeaveRoom(testGame, player)
	go RegisterLeaveRoom(testGame, player1)
	utils.Delay(1)
	newOwner := room.owner
	c.Assert(newOwner, NotNil)
	c.Assert(len(room.Players().coreMap), Equals, 2)
	fmt.Println(newOwner.Id(), player1.Id(), player.Id())
	c.Assert(newOwner.Id() != player1.Id(), Equals, true)
	c.Assert(newOwner.Id() != player.Id(), Equals, true)
	c.Assert(newOwner.GetAvailableMoney(currencyType), Equals, int64(900))
	c.Assert(room.GetMoneyOnTable(newOwner.Id()), Equals, int64(100))
}

func (s *TestSuite) TestOnlineOfflineResponse(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	s.server.cleanupAllResponse()
	room.HandlePlayerOffline(player2)
	for _, player := range []GamePlayer{player, player1, player2} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_status_change")
		c.Assert(utils.GetStringAtPath(response, "data/status"), Equals, "offline")
	}

	room.HandlePlayerOnline(player2)
	for _, player := range []GamePlayer{player, player1, player2} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_status_change")
		c.Assert(utils.GetStringAtPath(response, "data/status"), Equals, "online")
	}
}

func (s *TestSuite) TestRecordLastWin(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(100, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	_, err := room.StartGame(player)
	c.Assert(err, IsNil)
	s.server.cleanupAllResponse()

	room.DidEndGame(room.session.ResultSerializedData(), 0)
	c.Assert(len(room.lastMatchResults), Equals, 3) // hard code it to 3 under, in ResultSerializedData
	c.Assert(room.lastMatchResults[1].result, Equals, "win")
	c.Assert(room.lastMatchResults[2].result, Equals, "lose")
	c.Assert(room.lastMatchResults[2].rank, Equals, 10)
	c.Assert(room.lastMatchResults[3].result, Equals, "win")
}

func (s *TestSuite) TestPropertiesInGame(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(10000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")
	JoinRoomById(testGame, player2, room.Id(), "")

	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(100))

	testGame.properties = []string{GamePropertyCards}

	err := room.KickPlayer(player, player1)
	c.Assert(err.Error(), Equals, "err:cannot_kick_in_this_game")

	testGame.properties = []string{GamePropertyCards, GamePropertyAlwaysHasOwner}

	_, err = room.StartGame(player)
	// owner when start game will got his money on the table, since this is play against everyone
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(100))

	room.DidEndGame(room.session.ResultSerializedData(), 0)
	room.session = nil
	utils.DelayInDuration(1 * time.Second)
	testGame.properties = []string{GamePropertyCards, GamePropertyAlwaysHasOwner, GamePropertyRaiseBet}
	fmt.Println("here", testGame.Properties())
	_, err = room.StartGame(player)
	c.Assert(err, IsNil)
	utils.DelayInDuration(1 * time.Second)
	c.Assert(room.GetMoneyOnTable(player1.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player2.Id()), Equals, int64(100))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(100))
}

func (s *TestSuite) TestChat(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(10000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(1000, currencyType)

	player2 := s.newPlayer()
	player2.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")

	err := room.Chat(player2, "hihi")
	c.Assert(err.Error(), Equals, "err:player_not_in_room")

	s.server.cleanupAllResponse()
	err = room.Chat(player1, "haha")
	c.Assert(err, IsNil)

	for _, playerInstance := range []GamePlayer{player, player1} {
		response := s.server.getAndRemoveResponse(playerInstance.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "room_receive_message")
		c.Assert(utils.GetStringAtPath(response, "data/message"), Equals, "haha")
	}

	JoinRoomById(testGame, player2, room.Id(), "")
	s.server.cleanupAllResponse()

	err = room.Chat(player2, "hoho")
	c.Assert(err, IsNil)

	for _, playerInstance := range []GamePlayer{player, player1, player2} {
		response := s.server.getAndRemoveResponse(playerInstance.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "room_receive_message")
		c.Assert(utils.GetStringAtPath(response, "data/message"), Equals, "hoho")
	}

}

func (s *TestSuite) TestIncreaseMoney(c *C) {
	currencyType := currency.Money

	player := s.newPlayer()
	player.setMoney(10000, currencyType)

	player1 := s.newPlayer()
	player1.setMoney(10000, currencyType)

	testGame := NewTestGame(currencyType)
	room, _ := CreateRoom(testGame, player, 100, 4, "")
	JoinRoomById(testGame, player1, room.Id(), "")

	room.IncreaseMoney(player, 1000, true)
	money := room.GetTotalPlayerMoney(player.Id())
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(10900)) // lose money for put on table
	c.Assert(room.GetTotalPlayerMoney(player.Id()), Equals, int64(11000))
	c.Assert(money, Equals, int64(11000))

	room.IncreaseMoney(player1, 900, true)
	money = room.GetTotalPlayerMoney(player1.Id())
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(10800)) // lose money for put on table
	c.Assert(room.GetTotalPlayerMoney(player1.Id()), Equals, int64(10900))
	c.Assert(money, Equals, int64(10900))

	player.FreezeMoney(5000, currencyType, room.GetRoomIdentifierString(), true)
	room.DecreaseMoney(player, 5000, true)
	money = room.GetTotalPlayerMoney(player.Id())
	c.Assert(player.GetAvailableMoney(currencyType), Equals, int64(6000))
	c.Assert(room.GetTotalPlayerMoney(player.Id()), Equals, int64(6000))
	c.Assert(room.GetMoneyOnTable(player.Id()), Equals, int64(0))
	c.Assert(money, Equals, int64(6000))

	player1.FreezeMoney(10900, currencyType, room.GetRoomIdentifierString(), true)
	room.DecreaseMoney(player1, 10900, true)
	money = room.GetTotalPlayerMoney(player1.Id())
	c.Assert(player1.GetAvailableMoney(currencyType), Equals, int64(0)) // lose money for put on table
	c.Assert(room.GetTotalPlayerMoney(player1.Id()), Equals, int64(0))
	c.Assert(money, Equals, int64(0))

	err := room.IncreaseMoney(player1, -100000, true)
	c.Assert(err, NotNil)
}

/*
helper
*/

type TestModels struct {
	players []GamePlayer
}

func NewTestModels(players []GamePlayer) *TestModels {
	return &TestModels{
		players: players,
	}
}

func (models *TestModels) GetGamePlayer(playerId int64) (playerInstance GamePlayer, err error) {
	for _, playerInstance := range models.players {
		if playerInstance.Id() == playerId {
			return playerInstance, nil
		}
	}
	return nil, nil
}

type TestPlayer struct {
	currencyGroup *currency.CurrencyGroup

	id       int64
	name     string
	room     *Room
	isOnline bool
	exp      int64
	bet      int64

	playerType string
}

func (player *TestPlayer) LockMoney(currencyType string) {
	player.currencyGroup.Lock(currencyType)
}

func (player *TestPlayer) IpAddress() string {
	return ""
}

func (player *TestPlayer) UnlockMoney(currencyType string) {
	player.currencyGroup.Unlock(currencyType)
}

func (player *TestPlayer) IncreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	return player.currencyGroup.IncreaseMoney(money, currencyType, shouldLock)
}

func (player *TestPlayer) DecreaseMoney(money int64, currencyType string, shouldLock bool) (newMoney int64, err error) {
	return player.currencyGroup.DecreaseMoney(money, currencyType, shouldLock)
}

func (player *TestPlayer) FreezeMoney(money int64, currencyType string, reasonString string, shouldLock bool) (err error) {
	return player.currencyGroup.FreezeValue(currencyType, reasonString, money, shouldLock)
}

func (player *TestPlayer) IncreaseFreezeMoney(increaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	return player.currencyGroup.IncreaseFreezeValue(increaseAmount, currencyType, reasonString, shouldLock)
}

func (player *TestPlayer) DecreaseFromFreezeMoney(decreaseAmount int64, currencyType string, reasonString string, shouldLock bool) (newValue int64, err error) {
	return player.currencyGroup.DecreaseFromFreezeValue(decreaseAmount, currencyType, reasonString, shouldLock)
}

func (player *TestPlayer) GetFreezeValue(currencyType string) int64 {
	return player.currencyGroup.TotalFreezeValue(currencyType)
}

func (player *TestPlayer) GetFreezeValueForReason(currencyType string, reasonString string) int64 {
	return player.currencyGroup.GetFreezeValue(currencyType, reasonString)
}

func (player *TestPlayer) GetAvailableMoney(currencyType string) int64 {
	return player.currencyGroup.TotalAvailableValue(currencyType)
}

func (player *TestPlayer) GetMoney(currencyType string) int64 {
	return player.currencyGroup.GetValue(currencyType)
}

func (player *TestPlayer) setMoney(money int64, currencyType string) {
	player.currencyGroup.IncreaseMoney(money-player.GetAvailableMoney(currencyType), currencyType, true)
}

func (player *TestPlayer) Id() int64 {
	return player.id
}
func (player *TestPlayer) Name() string {
	return player.name
}

func (player *TestPlayer) Room() *Room {
	return player.room
}

func (player *TestPlayer) SetRoom(room *Room) {
	player.room = room
}

func (player *TestPlayer) SerializedData() map[string]interface{} {
	return map[string]interface{}{
		"id":       player.id,
		"username": player.name,
	}
}

func (player *TestPlayer) SerializedDataMinimal() map[string]interface{} {
	return map[string]interface{}{
		"id":       player.id,
		"username": player.name,
	}
}
func (player *TestPlayer) IncreaseVipPointForMatch(bet int64, matchId int64, gameCode string) {
}

func (player *TestPlayer) RecordGameResult(gameCode string, result string, change int64, currencyType string) (err error) {
	return nil
}

func (player *TestPlayer) IsOnline() bool {
	return player.isOnline
}

func (player *TestPlayer) IncreaseBet(bet int64) {
	player.bet += bet
}

func (player *TestPlayer) SetIsOnline(isOnline bool) {
	player.isOnline = isOnline
}

func (player *TestPlayer) IncreaseExp(exp int64) (newExp int64, err error) {
	player.exp = player.exp + exp
	return player.exp, nil
}

func (player *TestPlayer) PlayerType() string {
	return player.playerType
}

type TestGame struct {
	gameCode                 string
	currencyType             string
	minNumberOfPlayers       int
	maxNumberOfPlayers       int
	defaultNumberOfPlayers   int
	maxInactiveTimeInSeconds time.Duration
	gameData                 *GameData

	roomType   string
	properties []string

	tax          float64
	leavePenalty float64

	requirementMultiplier  float64
	ownerMultiplier        float64
	moneyOnTableMultiplier int64

	betData BetDataInterface
}

func NewTestGame(currencyType string) *TestGame {

	testGame := &TestGame{
		gameCode:                 "test_game",
		currencyType:             currencyType,
		gameData:                 NewGameData(),
		minNumberOfPlayers:       4,
		maxNumberOfPlayers:       9,
		defaultNumberOfPlayers:   4,
		requirementMultiplier:    1,
		ownerMultiplier:          1,
		moneyOnTableMultiplier:   1,
		roomType:                 RoomTypeList,
		maxInactiveTimeInSeconds: time.Duration(1000) * time.Second,
		tax:          0.05,
		leavePenalty: 2,
		properties:   []string{GamePropertyCards, GamePropertyAlwaysHasOwner, GamePropertyCanKick, GamePropertyRaiseBet},
	}

	betData := NewBetData(testGame, []BetEntryInterface{
		NewBetEntry(10, 800, 10, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(200, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(1000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(10000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100000, 200000, 10000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(500000, 1000000, 50000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(2500000, 5000000, 250000, 0.03, 0, "", "", nil, true, ""),
	})
	testGame.betData = betData

	return testGame
}

func (game *TestGame) CurrencyType() string {
	return game.currencyType
}

func (game *TestGame) Load() {

}

func (game *TestGame) GameCode() string {
	return game.gameCode
}

func (game *TestGame) CheatCode() string {
	return ""
}

func (game *TestGame) GameData() *GameData {
	return game.gameData
}

func (game *TestGame) DefaultNumberOfPlayers() int {
	return game.defaultNumberOfPlayers
}

func (game *TestGame) MinNumberOfPlayers() int {
	return game.minNumberOfPlayers
}

func (game *TestGame) MaxNumberOfPlayers() int {
	return game.maxNumberOfPlayers
}

func (game *TestGame) MaxInactiveTimeInSeconds() time.Duration {
	return game.maxInactiveTimeInSeconds
}

func (game *TestGame) SerializedData() map[string]interface{} {
	return nil
}

func (game *TestGame) UpdateData(data map[string]interface{}) {

}

func (game *TestGame) VipThreshold() int64 {
	return 0
}

func (game *TestGame) LeavePenalty() float64 {
	return game.leavePenalty
}

func (game *TestGame) Tax() float64 {
	return game.tax
}

func (game *TestGame) RequirementMultiplier() float64 {
	return game.requirementMultiplier
}

func (game *TestGame) Version() string {
	return "1.0"
}

func (game *TestGame) BetData() BetDataInterface {
	return game.betData
}

func (game *TestGame) MoneyOnTableMultiplier() int64 {
	return game.moneyOnTableMultiplier
}

func (game *TestGame) ShouldSaveRoom() bool {
	return true
}

func (game *TestGame) RoomType() string {
	return game.roomType
}

func (game *TestGame) Properties() []string {
	return game.properties
}

func (session *TestGame) SerializedDataForAdmin() map[string]interface{} {
	data := make(map[string]interface{})
	return data
}

func (game *TestGame) NewSession(models ModelsInterface, finishCallback ActivityGameSessionCallback, data map[string]interface{}) (session GameSessionInterface, err error) {
	return &TestGameSession{}, nil
}
func (game *TestGame) StartGame(finishCallback ActivityGameSessionCallback,
	owner GamePlayer,
	players []GamePlayer,
	bet int64,
	bets map[int64]int64,
	lastMatchResults map[int64]*GameResult) (session GameSessionInterface, err error) {

	return &TestGameSession{
		currencyType: game.currencyType,
		isPlaying:    true,
		players:      players,
		callback:     finishCallback,
	}, nil
}

func (game *TestGame) HandleRoomCreated(room *Room) {

}

func (game *TestGame) ConfigEditObject() *htmlutils.EditObject {
	return nil
}

func (gameInstance *TestGame) IsRoomRequirementValid(requirement int64) bool {
	return IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *TestGame) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}
func (gameInstance *TestGame) IsPlayerMoneyValidToJoinRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToJoinRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGame) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGame) IsPlayerMoneyValidToCreateRoom(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int) (err error) {
	return IsPlayerMoneyValidToCreateRoom(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers)
}
func (gameInstance *TestGame) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	if playerMoney < gameInstance.MoneyOnTableForOwner(roomRequirement, maxNumberOfPlayers, numberOfPlayers) {
		return errors.New(l.Get(l.M0016))
	}
	return nil
}

func (gameInstance *TestGame) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.requirementMultiplier)
}

func (gameInstance *TestGame) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, gameInstance.ownerMultiplier)
}

type TestServer struct {
	receiveDataMap map[int64][]map[string]interface{}
}

func NewTestServer() *TestServer {
	server := &TestServer{
		receiveDataMap: make(map[int64][]map[string]interface{}),
	}
	return server
}

func (server *TestServer) getAndRemoveResponse(toPlayerId int64) map[string]interface{} {
	fullData := server.receiveDataMap[toPlayerId][0]
	server.receiveDataMap[toPlayerId] = server.receiveDataMap[toPlayerId][1:]
	return fullData
}

func (server *TestServer) numberOfResponses(toPlayerId int64) int {
	return len(server.receiveDataMap[toPlayerId])
}

func (server *TestServer) cleanupAllResponse() {
	server.receiveDataMap = make(map[int64][]map[string]interface{})
}

func (server *TestServer) SendRequest(requestType string, data map[string]interface{}, toPlayerId int64) {
	responseList := server.receiveDataMap[toPlayerId]
	if responseList == nil {
		responseList = make([]map[string]interface{}, 0)
	}
	fullData := make(map[string]interface{})
	fullData["method"] = requestType
	fullData["data"] = data
	server.receiveDataMap[toPlayerId] = append(responseList, fullData)
}
func (server *TestServer) SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64) {
	for _, toPlayerId := range toPlayerIds {
		responseList := server.receiveDataMap[toPlayerId]
		if responseList == nil {
			responseList = make([]map[string]interface{}, 0)
		}
		fullData := make(map[string]interface{})
		fullData["method"] = requestType
		fullData["data"] = data
		server.receiveDataMap[toPlayerId] = append(responseList, fullData)
	}
}

func (server *TestServer) SendRequestsToAll(requestType string, data map[string]interface{}) {

}

func (s *TestSuite) genPlayerId() int64 {
	s.playerIdCounter = s.playerIdCounter + 1
	return s.playerIdCounter
}

func (s *TestSuite) newPlayer() *TestPlayer {
	username := utils.RandSeq(20)
	row := s.dataCenter.Db().QueryRow("INSERT INTO player (username, player_type,identifier) VALUES ($1,$2,$3) RETURNING id",
		username, "", utils.RandSeq(15))
	var id int64
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("err create new player", err)
	}
	testPlayer := &TestPlayer{
		currencyGroup: currency.NewCurrencyGroup(id),
		id:            id,
		name:          username,
		playerType:    "normal",
		isOnline:      true,
	}
	testPlayer.IncreaseMoney(100000, currency.Money, true)
	testPlayer.IncreaseMoney(100000, currency.TestMoney, true)
	return testPlayer
}

func (s *TestSuite) newPlayerWithName(name string) *TestPlayer {
	username := name
	row := s.dataCenter.Db().QueryRow("INSERT INTO player (username, player_type,identifier) VALUES ($1,$2,$3) RETURNING id",
		username, "", utils.RandSeq(15))
	var id int64
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("err create new player", err)
	}
	testPlayer := &TestPlayer{
		currencyGroup: currency.NewCurrencyGroup(id),
		id:            id,
		name:          username,
		playerType:    "normal",
		isOnline:      true,
	}
	testPlayer.IncreaseMoney(100000, currency.Money, true)
	testPlayer.IncreaseMoney(100000, currency.TestMoney, true)
	return testPlayer
}

func (s *TestSuite) justCreateRoom(game GameInterface, currencyType string) *Room {
	player := s.newPlayer()
	room, _ := CreateRoom(game, player, 10, 4, "")
	return room
}

func (s *TestSuite) justCreateRoomWithOwnerName(game GameInterface, name string, currencyType string) *Room {
	player := s.newPlayerWithName(name)
	room, _ := CreateRoom(game, player, 10, 4, "")
	return room
}

type TestGameSession struct {
	currencyType   string
	players        []GamePlayer
	callback       ActivityGameSessionCallback
	didHandleLeave bool
	isPlaying      bool
}

func (gameSession *TestGameSession) CleanUp() {

}
func (session *TestGameSession) CalculateResultsAndEndSession() {
	session.callback.DidEndGame(session.ResultSerializedData(), 0)
}
func (session *TestGameSession) SerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	return data
}
func (session *TestGameSession) ResultSerializedData() map[string]interface{} {
	data := make(map[string]interface{})
	// fake this
	data["results"] = []map[string]interface{}{
		map[string]interface{}{
			"id":     1,
			"result": "win",
			"change": 4,
		},
		map[string]interface{}{
			"id":     2,
			"result": "lose",
			"change": 4,
			"rank":   10,
		},
		map[string]interface{}{
			"id":     3,
			"result": "win",
			"change": 4,
		},
	}
	return data
}
func (session *TestGameSession) SerializedDataForPlayer(player GamePlayer) map[string]interface{} {
	data := make(map[string]interface{})
	return data
}
func (gameSession *TestGameSession) HandlePlayerRemovedFromGame(leavePlayer GamePlayer) {
	gameSession.didHandleLeave = true
}

func (gameSession *TestGameSession) HandlePlayerOffline(player GamePlayer) {

}
func (gameSession *TestGameSession) HandlePlayerOnline(player GamePlayer) {

}
func (gameSession *TestGameSession) HandlePlayerAddedToGame(player GamePlayer) {

}
func (gameSession *TestGameSession) IsDelayingForNewGame() bool {
	return false
}

func (gameSession *TestGameSession) IsPlaying() bool {
	return gameSession.isPlaying
}

func (gameSession *TestGameSession) GetPlayer(playerId int64) (player GamePlayer) {
	for _, player := range gameSession.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (gameSession *TestGameSession) Owner() (player GamePlayer) {
	return nil
}

type TestGameWithMultiplier struct {
	gameCode                 string
	currencyType             string
	minNumberOfPlayers       int
	maxNumberOfPlayers       int
	defaultNumberOfPlayers   int
	maxInactiveTimeInSeconds time.Duration
	gameData                 *GameData

	roomType   string
	properties []string

	tax          float64
	leavePenalty float64

	requirementMultiplier    float64
	moneyOnTableMultiplier   int64
	needOwner                bool
	ownerCanRaiseRequirement bool
	isPlayedAgainstOwner     bool

	betData BetDataInterface
}

func NewTestGameWithMultiplier(currencyType string) *TestGameWithMultiplier {

	testGame := &TestGameWithMultiplier{
		gameCode:                 "test_game",
		currencyType:             currencyType,
		gameData:                 NewGameData(),
		minNumberOfPlayers:       4,
		maxNumberOfPlayers:       9,
		defaultNumberOfPlayers:   4,
		requirementMultiplier:    1,
		moneyOnTableMultiplier:   1,
		needOwner:                true,
		ownerCanRaiseRequirement: false,
		isPlayedAgainstOwner:     true,
		roomType:                 RoomTypeList,
		maxInactiveTimeInSeconds: time.Duration(1000) * time.Second,
		tax:          0.05,
		leavePenalty: 2,
		properties:   []string{GamePropertyCards, GamePropertyAlwaysHasOwner, GamePropertyRaiseBet, GamePropertyCanKick},
	}

	betData := NewBetData(testGame, []BetEntryInterface{
		NewBetEntry(10, 800, 10, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(200, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(1000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(10000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100000, 200000, 10000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(500000, 1000000, 50000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(2500000, 5000000, 250000, 0.03, 0, "", "", nil, true, ""),
	})
	testGame.betData = betData

	return testGame
}

func (game *TestGameWithMultiplier) ConfigEditObject() *htmlutils.EditObject {
	return nil
}

func (game *TestGameWithMultiplier) CurrencyType() string {
	return game.currencyType
}

func (game *TestGameWithMultiplier) Load() {

}

func (game *TestGameWithMultiplier) GameCode() string {
	return game.gameCode
}

func (game *TestGameWithMultiplier) CheatCode() string {
	return ""
}

func (game *TestGameWithMultiplier) GameData() *GameData {
	return game.gameData
}

func (game *TestGameWithMultiplier) DefaultNumberOfPlayers() int {
	return game.defaultNumberOfPlayers
}

func (game *TestGameWithMultiplier) MinNumberOfPlayers() int {
	return game.minNumberOfPlayers
}

func (game *TestGameWithMultiplier) MaxNumberOfPlayers() int {
	return game.maxNumberOfPlayers
}

func (game *TestGameWithMultiplier) MaxInactiveTimeInSeconds() time.Duration {
	return game.maxInactiveTimeInSeconds
}

func (game *TestGameWithMultiplier) SerializedData() map[string]interface{} {
	return nil
}

func (game *TestGameWithMultiplier) UpdateData(data map[string]interface{}) {

}

func (game *TestGameWithMultiplier) VipThreshold() int64 {
	return 0
}

func (game *TestGameWithMultiplier) LeavePenalty() float64 {
	return game.leavePenalty
}

func (game *TestGameWithMultiplier) Tax() float64 {
	return game.tax
}

func (game *TestGameWithMultiplier) RequirementMultiplier() float64 {
	return game.requirementMultiplier
}

func (game *TestGameWithMultiplier) NeedOwner() bool {
	return game.needOwner
}

func (game *TestGameWithMultiplier) OwnerCanRaiseRequirement() bool {
	return game.ownerCanRaiseRequirement
}

func (game *TestGameWithMultiplier) IsPlayedAgainstOwner() bool {
	return game.isPlayedAgainstOwner
}

func (game *TestGameWithMultiplier) Version() string {
	return "1.0"
}

func (game *TestGameWithMultiplier) BetData() BetDataInterface {
	return game.betData
}

func (game *TestGameWithMultiplier) MoneyOnTableMultiplier() int64 {
	return game.moneyOnTableMultiplier
}

func (game *TestGameWithMultiplier) ShouldSaveRoom() bool {
	return true
}

func (game *TestGameWithMultiplier) RoomType() string {
	return game.roomType
}

func (game *TestGameWithMultiplier) Properties() []string {
	return game.properties
}

func (session *TestGameWithMultiplier) SerializedDataForAdmin() map[string]interface{} {
	data := make(map[string]interface{})
	return data
}

func (game *TestGameWithMultiplier) NewSession(models ModelsInterface, finishCallback ActivityGameSessionCallback, data map[string]interface{}) (session GameSessionInterface, err error) {
	return &TestGameSession{}, nil
}
func (game *TestGameWithMultiplier) StartGame(finishCallback ActivityGameSessionCallback,
	owner GamePlayer,
	players []GamePlayer,
	bet int64,
	bets map[int64]int64,
	lastMatchResults map[int64]*GameResult) (session GameSessionInterface, err error) {

	return &TestGameSession{
		currencyType: game.currencyType,
		isPlaying:    true,
		players:      players,
		callback:     finishCallback,
	}, nil
}

func (game *TestGameWithMultiplier) HandleRoomCreated(room *Room) {

}
func (gameInstance *TestGameWithMultiplier) IsRoomRequirementValid(requirement int64) bool {
	return IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *TestGameWithMultiplier) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}
func (gameInstance *TestGameWithMultiplier) IsPlayerMoneyValidToJoinRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToJoinRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGameWithMultiplier) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGameWithMultiplier) IsPlayerMoneyValidToCreateRoom(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int) (err error) {
	return IsPlayerMoneyValidToCreateRoom(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers)
}
func (gameInstance *TestGameWithMultiplier) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	return IsPlayerMoneyValidToBecomeOwner(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}
func (gameInstance *TestGameWithMultiplier) IsPlayerMoneyValidToStayOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	return IsPlayerMoneyValidToStayOwner(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers, numberOfPlayers)
}

func (gameInstance *TestGameWithMultiplier) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, float64(gameInstance.moneyOnTableMultiplier))
}

func (gameInstance *TestGameWithMultiplier) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, float64(gameInstance.moneyOnTableMultiplier))
}

type TestGameCasinoOwner struct {
	gameCode                 string
	currencyType             string
	minNumberOfPlayers       int
	maxNumberOfPlayers       int
	defaultNumberOfPlayers   int
	maxInactiveTimeInSeconds time.Duration
	gameData                 *GameData

	roomType   string
	properties []string

	tax          float64
	leavePenalty float64

	requirementMultiplier    float64
	moneyOnTableMultiplier   int64
	needOwner                bool
	ownerCanRaiseRequirement bool
	isPlayedAgainstOwner     bool

	betData BetDataInterface

	becomeOwnerRequirement int64
}

func NewTestGameCasinoOwner(currencyType string) *TestGameCasinoOwner {

	testGame := &TestGameCasinoOwner{
		gameCode:                 "test_game",
		currencyType:             currencyType,
		gameData:                 NewGameData(),
		minNumberOfPlayers:       4,
		maxNumberOfPlayers:       9,
		defaultNumberOfPlayers:   4,
		requirementMultiplier:    1,
		moneyOnTableMultiplier:   1,
		needOwner:                true,
		ownerCanRaiseRequirement: false,
		isPlayedAgainstOwner:     true,
		roomType:                 RoomTypeList,
		maxInactiveTimeInSeconds: time.Duration(1000) * time.Second,
		tax:                    0.05,
		leavePenalty:           2,
		properties:             []string{GamePropertyCasino, GamePropertyRegisterOwner, GamePropertyCanKick, GamePropertyPersistentSession},
		becomeOwnerRequirement: 1000,
	}

	betData := NewBetData(testGame, []BetEntryInterface{
		NewBetEntry(10, 800, 10, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(200, 10000, 100, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(1000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(10000, 40000, 2000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(100000, 200000, 10000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(500000, 1000000, 50000, 0.03, 0, "", "", nil, true, ""),
		NewBetEntry(2500000, 5000000, 250000, 0.03, 0, "", "", nil, true, ""),
	})
	testGame.betData = betData

	return testGame
}

func (game *TestGameCasinoOwner) ConfigEditObject() *htmlutils.EditObject {
	return nil
}

func (game *TestGameCasinoOwner) CurrencyType() string {
	return game.currencyType
}

func (game *TestGameCasinoOwner) Load() {

}

func (game *TestGameCasinoOwner) GameCode() string {
	return game.gameCode
}

func (game *TestGameCasinoOwner) CheatCode() string {
	return ""
}

func (game *TestGameCasinoOwner) GameData() *GameData {
	return game.gameData
}

func (game *TestGameCasinoOwner) DefaultNumberOfPlayers() int {
	return game.defaultNumberOfPlayers
}

func (game *TestGameCasinoOwner) MinNumberOfPlayers() int {
	return game.minNumberOfPlayers
}

func (game *TestGameCasinoOwner) MaxNumberOfPlayers() int {
	return game.maxNumberOfPlayers
}

func (game *TestGameCasinoOwner) MaxInactiveTimeInSeconds() time.Duration {
	return game.maxInactiveTimeInSeconds
}

func (game *TestGameCasinoOwner) SerializedData() map[string]interface{} {
	return nil
}

func (game *TestGameCasinoOwner) UpdateData(data map[string]interface{}) {

}

func (game *TestGameCasinoOwner) VipThreshold() int64 {
	return 0
}

func (game *TestGameCasinoOwner) LeavePenalty() float64 {
	return game.leavePenalty
}

func (game *TestGameCasinoOwner) Tax() float64 {
	return game.tax
}

func (game *TestGameCasinoOwner) RequirementMultiplier() float64 {
	return game.requirementMultiplier
}

func (game *TestGameCasinoOwner) NeedOwner() bool {
	return game.needOwner
}

func (game *TestGameCasinoOwner) OwnerCanRaiseRequirement() bool {
	return game.ownerCanRaiseRequirement
}

func (game *TestGameCasinoOwner) IsPlayedAgainstOwner() bool {
	return game.isPlayedAgainstOwner
}

func (game *TestGameCasinoOwner) Version() string {
	return "1.0"
}

func (game *TestGameCasinoOwner) BetData() BetDataInterface {
	return game.betData
}

func (game *TestGameCasinoOwner) MoneyOnTableMultiplier() int64 {
	return game.moneyOnTableMultiplier
}

func (game *TestGameCasinoOwner) ShouldSaveRoom() bool {
	return true
}

func (game *TestGameCasinoOwner) RoomType() string {
	return game.roomType
}

func (game *TestGameCasinoOwner) Properties() []string {
	return game.properties
}

func (session *TestGameCasinoOwner) SerializedDataForAdmin() map[string]interface{} {
	data := make(map[string]interface{})
	return data
}

func (game *TestGameCasinoOwner) NewSession(models ModelsInterface, finishCallback ActivityGameSessionCallback, data map[string]interface{}) (session GameSessionInterface, err error) {
	return &TestGameSession{}, nil
}
func (game *TestGameCasinoOwner) StartGame(finishCallback ActivityGameSessionCallback,
	owner GamePlayer,
	players []GamePlayer,
	bet int64,
	bets map[int64]int64,
	lastMatchResults map[int64]*GameResult) (session GameSessionInterface, err error) {

	return &TestGameSession{
		currencyType: game.currencyType,
		isPlaying:    true,
		players:      players,
		callback:     finishCallback,
	}, nil
}

func (game *TestGameCasinoOwner) HandleRoomCreated(room *Room) {

	players := make([]GamePlayer, 0)
	for _, player := range room.players.copy() {
		players = append(players, player)
	}

	session := &TestGameSession{
		currencyType: room.currencyType,
		isPlaying:    true,
		players:      players,
		callback:     room,
	}

	room.SetSession(session)
}

func (gameInstance *TestGameCasinoOwner) IsRoomRequirementValid(requirement int64) bool {
	return IsRoomRequirementValid(gameInstance, requirement)
}
func (gameInstance *TestGameCasinoOwner) IsRoomMaxPlayersValid(maxPlayer int, roomRequirement int64) bool {
	return IsRoomMaxPlayersValid(gameInstance, maxPlayer, roomRequirement)
}
func (gameInstance *TestGameCasinoOwner) IsPlayerMoneyValidToJoinRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToJoinRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGameCasinoOwner) IsPlayerMoneyValidToStayInRoom(playerMoney int64, roomRequirement int64) (err error) {
	return IsPlayerMoneyValidToStayInRoom(gameInstance, playerMoney, roomRequirement)
}
func (gameInstance *TestGameCasinoOwner) IsPlayerMoneyValidToCreateRoom(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int) (err error) {
	return IsPlayerMoneyValidToCreateRoom(gameInstance, playerMoney, roomRequirement, maxNumberOfPlayers)
}
func (gameInstance *TestGameCasinoOwner) IsPlayerMoneyValidToBecomeOwner(playerMoney int64, roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) (err error) {
	if playerMoney < gameInstance.becomeOwnerRequirement {
		return details_error.NewError("err:requirement_not_meet", map[string]interface{}{
			"need": gameInstance.becomeOwnerRequirement,
		})
	}
	return nil
}

func (gameInstance *TestGameCasinoOwner) MoneyOnTable(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return utils.Int64AfterApplyFloat64Multiplier(roomRequirement, float64(gameInstance.moneyOnTableMultiplier))
}

func (gameInstance *TestGameCasinoOwner) MoneyOnTableForOwner(roomRequirement int64, maxNumberOfPlayers int, numberOfPlayers int) int64 {
	return gameInstance.becomeOwnerRequirement
}
