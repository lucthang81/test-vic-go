package tienlen

import (
	"errors"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/tienlen/logic"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/test"
	"github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	"math/rand"
	"sync"
	"testing"
	"time"
	// "log"
)

const waitTimeForTurn time.Duration = 100 * time.Millisecond

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter      *datacenter.DataCenter
	server          *TestServer
	dbName          string
	playerIdCounter int64
}

var _ = Suite(&TestSuite{
	dbName: "casino_tienlen_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../../sql/init_schema.sql", "../../../sql/test_schema/player_test.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	s.server = NewTestServer()
	currency.RegisterDataCenter(s.dataCenter)
	game.RegisterDataCenter(s.dataCenter)
	game.RegisterServer(s.server)
	record.RegisterDataCenter(s.dataCenter)
}

func (s *TestSuite) TearDownSuite(c *C) {
	s.dataCenter.Db().Close()
	test.DropTestDatabase(s.dbName)

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
	c.Assert(true, Equals, true)
}

func (s *TestSuite) TestRemoveCards(c *C) {
	cards := removeCardsFromCards([]string{"c j", "c q", "s k", "h a", "s 3"}, []string{"c j", "c q"})
	c.Assert(len(cards), Equals, 3)
	c.Assert(logic.ContainCards(cards, []string{"c j", "c q"}), Equals, false)
}

func (s *TestSuite) TestWhoGoFirst(c *C) {
	currencyType := currency.Money
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	session.betEntry = testGame.BetData().GetEntry(100)
	testFinishCallback := NewTestFinishCallback(s.server, currencyType,
		[]game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback

	session.cards = make(map[int64][]string)
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "h 3", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 3", "d 3", "d 4", "h 10", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 9", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	session.start()
	// since no last result, the one go first will be player3 with 3s card
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	session.finished = true

	session = NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	session.betEntry = testGame.BetData().GetEntry(100)
	session.sessionCallback = testFinishCallback

	session.cards = make(map[int64][]string)
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 3", "d 3", "d 4", "h 10", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 9", "d 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	result1 := game.NewGameResult("win", 100, 1, 0)
	result2 := game.NewGameResult("win", 100, 3, 0)
	result3 := game.NewGameResult("win", 100, 2, 0)
	result4 := game.NewGameResult("win", 100, 0, 0)
	session.lastMatchResult = map[int64]*game.GameResult{
		player1.id: result1,
		player2.id: result2,
		player3.id: result3,
		player4.id: result4,
	}
	session.start()
	utils.DelayInDuration(waitTimeForTurn)
	// ignore last match result, player 3 will now go first
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	session.finished = true
	// all the player in the old room leave.... so the last result is irrelevant

	session = NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	session.betEntry = testGame.BetData().GetEntry(100)
	session.sessionCallback = testFinishCallback

	session.cards = make(map[int64][]string)
	result1 = game.NewGameResult("win", 100, 1, 0)
	result2 = game.NewGameResult("win", 100, 3, 0)
	result3 = game.NewGameResult("win", 100, 2, 0)
	result4 = game.NewGameResult("win", 100, 0, 0)
	session.lastMatchResult = map[int64]*game.GameResult{
		100: result1,
		101: result2,
		102: result3,
		103: result4,
	}

	session.cards = make(map[int64][]string)
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", " 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", " 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 3", " 3", "d 4", "h 10", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 9", " 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.start()
	utils.DelayInDuration(waitTimeForTurn)
	// ignore last result, the one go first will be player3 with 3s card
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

}

func (s *TestSuite) TestFreezeMoney(c *C) {
	currencyType := currency.Money
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	players := []game.GamePlayer{player1, player2, player3, player4}

	for _, player := range players {
		fmt.Printf("money %d freeze %d \n", player.GetMoney(currencyType), player.GetFreezeValue(currencyType))
	}

	testGame := NewTienLenGame(currencyType)
	room, err := game.JoinRoomByRequirement(testGame, player1, 100)
	c.Assert(err, IsNil)
	c.Assert(room, NotNil)

	_, err = game.JoinRoomById(testGame, player2, room.Id(), "")
	c.Assert(err, IsNil)
	_, err = game.JoinRoomById(testGame, player3, room.Id(), "")
	c.Assert(err, IsNil)
	_, err = game.JoinRoomById(testGame, player4, room.Id(), "")
	c.Assert(err, IsNil)

	for _, player := range players {
		fmt.Printf("money %d freeze %d \n", player.GetMoney(currencyType), player.GetFreezeValue(currencyType))
	}

	freezeAmount := utils.Int64AfterApplyFloat64Multiplier(100, testGame.requirementMultiplier)

	player1.DecreaseMoney(100000, currencyType, true)
	c.Assert(player1.GetMoney(currencyType), DeepEquals, freezeAmount)
	c.Assert(player1.GetFreezeValue(currencyType), DeepEquals, freezeAmount)

	for _, player := range players {
		fmt.Printf("money %d freeze %d \n", player.GetMoney(currencyType), player.GetFreezeValue(currencyType))
	}

	room.StartGame(player1)
	session := room.Session().(*TienLenSession)
	c.Assert(session, NotNil)
	c.Assert(len(session.players), Equals, 4)

	room.DecreaseMoney(player2, 90000, true)
	c.Assert(player2.GetMoney(currencyType), DeepEquals, 100000-freezeAmount)
	c.Assert(player2.GetFreezeValue(currencyType), DeepEquals, int64(0))

}

func (s *TestSuite) TestInstantWin(c *C) {
	currencyType := currency.Money
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	players := []game.GamePlayer{player1, player2, player3, player4}

	for _, player := range players {
		fmt.Printf("money %d freeze %d \n", player.GetMoney(currencyType), player.GetFreezeValue(currencyType))
	}

	testGame := NewTienLenGame(currencyType)
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType,
		[]game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c k", "c 2", "h 2", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 2", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q", "c k"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c k", "c 2", "c 2", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"c k", "c 2", "c 2", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	session.start()
	utils.DelayInDuration(waitTimeForTurn)
	// player2 will instant win
	c.Assert(session.finished, Equals, true)
	var gain int64
	mul1, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player1.id], false)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], false)
	mul4, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player4.id], false)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul1)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul4)
	resultsData := session.results
	c.Assert(resultsData, NotNil)
	fmt.Println(resultsData)
	for _, resultData := range resultsData {
		playerId := utils.GetInt64AtPath(resultData, "id")
		result := utils.GetStringAtPath(resultData, "result")
		change := utils.GetInt64AtPath(resultData, "change")
		cards := utils.GetStringSliceAtPath(resultData, "cards")
		c.Assert(len(cards) != 0, Equals, true)
		if playerId == player1.id {
			c.Assert(result, Equals, "lose")
			mul, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player1.id], false)
			c.Assert(change, Equals,
				-moneyAfterApplyMultiplier(session.betEntry.Min(), mul))
		} else if playerId == player2.id {
			c.Assert(result, Equals, "win")
			c.Assert(utils.GetStringAtPath(resultData, "win_type"), Equals, "instant")
			c.Assert(change, Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player3.id {
			c.Assert(result, Equals, "lose")
			mul, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], false)
			c.Assert(change, Equals,
				-moneyAfterApplyMultiplier(session.betEntry.Min(), mul))
		} else if playerId == player4.id {
			c.Assert(result, Equals, "lose")
			mul, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player4.id], false)
			c.Assert(change, Equals,
				-moneyAfterApplyMultiplier(session.betEntry.Min(), mul))
		}
	}
}

func (s *TestSuite) TestTurn(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()
	player1.playerType = "normal"
	player2.playerType = "bot"
	player3.playerType = "normal"
	player4.playerType = "normal"

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	session.playersData = []*PlayerData{
		&PlayerData{id: player1.Id()},
		&PlayerData{id: player2.Id()},
		&PlayerData{id: player3.Id()},
		&PlayerData{id: player4.Id()},
	}

	testFinishCallback := NewTestFinishCallback(s.server, currencyType,
		[]game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win
	cards1 := logic.SortCards(testGame.logicInstance, []string{"c 3", "d 5", "s 6", "d 8", "s 9", "h 9", "h 10", "s q", "c q", "s k", "s a", "c a", "h 2"})
	cards2 := logic.SortCards(testGame.logicInstance, []string{"s 3", "s 4", "c 5", "d 6", "c 8", "h 8", "c 9", "d 9", "d 10", "h q", "d k", "h k", "d 2"})
	cards3 := logic.SortCards(testGame.logicInstance, []string{"d 3", "c 4", "d 4", "h 4", "h 6", "s 7", "h 7", "s 8", "s 10", "c 10", "d q", "d a", "c 2"})
	cards4 := logic.SortCards(testGame.logicInstance, []string{"h 3", "s 5", "h 5", "c 6", "c 7", "d 7", "s j", "c j", "d j", "h j", "c k", "h a", "s 2"})

	session.cards[player1.Id()] = cards1
	session.cards[player2.Id()] = cards2
	session.cards[player3.Id()] = cards3
	session.cards[player4.Id()] = cards4

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id, player4.id)

	s.server.cleanupAllResponse()
	session.start()
	// no instant win
	c.Assert(session.finished, Equals, false)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_start_game_session")
		if player.Id() == player1.id {
			playersData := utils.GetMapSliceAtPath(response, "data/players_data")
			// check no cards in player data
			for _, playerData := range playersData {
				c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
			}

			// check card in order
			cards := utils.GetStringSliceAtPath(response, "data/cards")
			c.Assert(len(cards), Equals, 13)
			previousCard := ""
			counter := 0
			cardsToCompare := cards1
			for _, card := range cards {
				c.Assert(card, Equals, cardsToCompare[counter])
				if previousCard == "" {
					previousCard = card
				} else {
					c.Assert(logic.IsCard1BiggerThanCard2(testGame.logicInstance, previousCard, card), Equals, false)
				}
				counter++
			}

		} else if player.Id() == player2.id {
			fmt.Println(response)
			playersData := utils.GetMapSliceAtPath(response, "data/players_data")
			// check no cards in player data
			for _, playerData := range playersData {
				c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 13) //player2 is bot
			}

			// check card in order
			cards := utils.GetStringSliceAtPath(response, "data/cards")
			c.Assert(len(cards), Equals, 13)
			previousCard := ""
			counter := 0
			cardsToCompare := cards2
			for _, card := range cards {
				c.Assert(card, Equals, cardsToCompare[counter])
				if previousCard == "" {
					previousCard = card
				} else {
					c.Assert(logic.IsCard1BiggerThanCard2(testGame.logicInstance, previousCard, card), Equals, false)
				}
				counter++
			}
		} else if player.Id() == player3.id {
			playersData := utils.GetMapSliceAtPath(response, "data/players_data")
			// check no cards in player data
			for _, playerData := range playersData {
				c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
			}

			// check card in order
			cards := utils.GetStringSliceAtPath(response, "data/cards")
			c.Assert(len(cards), Equals, 13)
			previousCard := ""
			counter := 0
			cardsToCompare := cards3
			for _, card := range cards {
				c.Assert(card, Equals, cardsToCompare[counter])
				if previousCard == "" {
					previousCard = card
				} else {
					c.Assert(logic.IsCard1BiggerThanCard2(testGame.logicInstance, previousCard, card), Equals, false)
				}
				counter++
			}
		} else if player.Id() == player4.id {
			playersData := utils.GetMapSliceAtPath(response, "data/players_data")
			// check no cards in player data
			for _, playerData := range playersData {
				c.Assert(len(utils.GetStringSliceAtPath(playerData, "cards")), Equals, 0)
			}

			// check card in order
			cards := utils.GetStringSliceAtPath(response, "data/cards")
			c.Assert(len(cards), Equals, 13)
			previousCard := ""
			counter := 0
			cardsToCompare := cards4
			for _, card := range cards {
				c.Assert(card, Equals, cardsToCompare[counter])
				if previousCard == "" {
					previousCard = card
				} else {
					c.Assert(logic.IsCard1BiggerThanCard2(testGame.logicInstance, previousCard, card), Equals, false)
				}
				counter++
			}
		}
	}
	utils.DelayInDuration(waitTimeForTurn)
	// change with the first turn
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_change_game_session")
		if player.Id() == player1.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player2.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player3.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player4.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		}
	}

	// now player2 will play not s 3 card
	err = session.playCards(player2, []string{"c 8", "h 8"})
	c.Assert(err.Error(), Equals, l.Get(l.M0014))

	// player2 will try to skip too (cannot)
	err = session.skipTurn(player2)
	c.Assert(err.Error(), Equals, l.Get(l.M0015))

	utils.Delay(5)
	// change to seconds turn
	// wrap up last turn, since no move match, the first card of player2 will play
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 1)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "s 3")
	c.Assert(len(session.cards[player2.id]), Equals, 12)
	c.Assert(logic.ContainCards(session.cards[player2.id], []string{"s 3"}), Equals, false)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_change_game_session")
		if player.Id() == player1.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 1)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player3.id))
		} else if player.Id() == player2.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 1)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player3.id))
		} else if player.Id() == player3.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 1)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player3.id))
		} else if player.Id() == player4.id {
			// should return what turn it is (player2 since it has 3s)
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 1)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player3.id))
		}
	}

	utils.DelayInDuration(waitTimeForTurn)

	// some other player will try to skip
	err = session.skipTurn(player4)
	c.Assert(err.Error(), Equals, l.Get(l.M0012))

	// player 3 will play some cards
	err = session.playCards(player3, []string{"s 4"})
	c.Assert(err.Error(), Equals, l.Get(l.M0011))
	err = session.playCards(player3, []string{"h 4", "d 4"})
	c.Assert(err.Error(), Equals, l.Get(l.M0014))
	err = session.playCards(player3, []string{"h 4", "d 4", "h 6"})
	c.Assert(err.Error(), Equals, l.Get(l.M0014))
	err = session.playCards(player4, []string{"s 5"})
	c.Assert(err.Error(), Equals, l.Get(l.M0012))
	err = session.playCards(player3, []string{"d 4"})
	c.Assert(err, IsNil)
	// turn start in here, wait 1s after turn start
	utils.Delay(1)
	// change to third turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 2)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player3.id)
	c.Assert(session.cardsOnTable[0], Equals, "d 4")
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 12)
	c.Assert(session.turnCounter, Equals, 2)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_change_game_session")
		if player.Id() == player1.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 2)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player4.id))
		} else if player.Id() == player2.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 2)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player4.id))
		} else if player.Id() == player3.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 2)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player4.id))
		} else if player.Id() == player4.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 2)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player4.id))
		}
	}

	sessionData := session.serializedDataForAll()
	// we wait for 1 seconds, so turn time will now be 4 (max turn time is 5s)
	c.Assert(utils.GetFloat64AtPath(sessionData, "turn_time") > 3.9, Equals, true)
	c.Assert(utils.GetFloat64AtPath(sessionData, "turn_time") < 4.1, Equals, true)

	// player 4 will play too
	// test play 2 times instantly
	err = session.playCards(player4, []string{"h 5"})
	c.Assert(err, IsNil)
	err = session.playCards(player4, []string{"h 5"})
	c.Assert(err, NotNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 4th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 3)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "h 5")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player4.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 12)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 13)
	c.Assert(session.turnCounter, Equals, 3)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_change_game_session")
		if player.Id() == player1.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 3)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player1.id))
		} else if player.Id() == player2.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 3)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player1.id))
		} else if player.Id() == player3.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 3)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player1.id))
		} else if player.Id() == player4.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 3)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player1.id))
		}
	}

	// player 1 will play too
	// test play 2 times instantly different cards
	err = session.playCards(player1, []string{"s 6"})
	c.Assert(err, IsNil)
	err = session.playCards(player1, []string{"d 8"})
	c.Assert(err, NotNil)

	utils.DelayInDuration(waitTimeForTurn)
	// change to 5th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "s 6")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 12)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 12)
	c.Assert(session.turnCounter, Equals, 4)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "tienlen_change_game_session")
		if player.Id() == player1.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 4)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player2.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 4)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player3.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 4)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		} else if player.Id() == player4.id {
			c.Assert(utils.GetIntAtPath(response, "data/turn_counter"), Equals, 4)
			c.Assert(utils.GetInt64AtPath(response, "data/current_player_id_turn"), Equals, int64(player2.id))
		}
	}

	// player 2 will play
	err = session.playCards(player2, []string{"d k"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 6th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 5)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "d k")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player2.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 11)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 12)
	c.Assert(session.turnCounter, Equals, 5)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 3 will play not play -> skip automatically
	utils.Delay(5)
	// will now change to player 4
	// change to 7th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 5)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(len(session.playersInCurrentRound), Equals, 3)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(session.cardsOnTable[0], Equals, "d k")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player2.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 11)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 12)
	c.Assert(session.turnCounter, Equals, 6)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 4 will call skip
	err = session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 8th turn, player 1
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 5)
	c.Assert(len(session.playersInCurrentRound), Equals, 2)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "d k")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player2.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 11)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 12)
	c.Assert(session.turnCounter, Equals, 7)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 will not call skip
	err = session.playCards(player1, []string{"s a"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 9th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 6)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "s a")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 11)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 11)
	c.Assert(session.turnCounter, Equals, 8)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 will play
	err = session.playCards(player2, []string{"d 2"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 10th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 7)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "d 2")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player2.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 10)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 11)
	c.Assert(session.turnCounter, Equals, 9)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 will play
	err = session.playCards(player1, []string{"h 2"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 11th turn
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 8)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "h 2")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 10)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 10)
	c.Assert(session.turnCounter, Equals, 10)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 will skip now, so player 1 will win this round, start next round right away
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// change to 1st turn, second round, player 1
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 0)
	c.Assert(len(session.playersInCurrentRound), Equals, 4)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 0)
	// fmt.Println(session.cards)
	c.Assert(len(session.cards[player3.id]), Equals, 12)
	c.Assert(len(session.cards[player2.id]), Equals, 10)
	c.Assert(len(session.cards[player4.id]), Equals, 12)
	c.Assert(len(session.cards[player1.id]), Equals, 10)
	c.Assert(session.turnCounter, Equals, 0)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	/*
			cards1: [c 3, d 5, d 8, s 9, h 9, h 10, s q, c q, s k, c a]
		 	cards2: [s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
		 	cards3: [d 3, c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
		 	cards4: [h 3, s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	*/
	// player1 will wait -> auto play c 3
	utils.Delay(6)
	// change to 2nd turn, second round
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 1)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "c 3")
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(len(session.cards[player1.id]), Equals, 9)  //[d 5, d 8, s 9, h 9, h 10, s q, c q, s k, c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 12) //[d 3, c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 12) //[h 3, s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 1)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 2 will skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 3 will play
	err = session.playCards(player3, []string{"d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player3.id)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player4 will play
	err = session.playCards(player4, []string{"h 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player4.id)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 will play
	err = session.playCards(player1, []string{"h 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 3,4 will skip
	err = session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}
	err = session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// end round 2, player 1 win, next round with player 1
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 0)
	c.Assert(len(session.playersInCurrentRound), Equals, 4)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 0)
	// fmt.Println(session.cards)
	c.Assert(len(session.cards[player1.id]), Equals, 8)  //[d 5, d 8, s 9, h 10, s q, c q, s k, c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 0)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 will play a streak
	err = session.playCards(player1, []string{"d 8", "s 9", "h 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 1)
	c.Assert(len(session.cardsOnTable), Equals, 3)
	c.Assert(session.cardsOnTable[0], Equals, "d 8")
	c.Assert(session.cardsOnTable[1], Equals, "s 9")
	c.Assert(session.cardsOnTable[2], Equals, "h 10")
	c.Assert(len(session.cards[player1.id]), Equals, 5)  //[d 5, s q, c q, s k, c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 1)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 will play and fail
	err = session.playCards(player2, []string{"h 8", "c 9", "d 10"})
	c.Assert(err, NotNil) // cannot play, cause streak is smaller
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 1)
	c.Assert(len(session.cardsOnTable), Equals, 3)
	c.Assert(session.cardsOnTable[0], Equals, "d 8")
	c.Assert(session.cardsOnTable[1], Equals, "s 9")
	c.Assert(session.cardsOnTable[2], Equals, "h 10")
	c.Assert(len(session.cards[player1.id]), Equals, 5)  //[d 5, s q, c q, s k, c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 2)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	//player3,player4 will skip too
	err = session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}
	err = session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	// end round 2, player 1 win, next round with player 1
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 0)
	c.Assert(len(session.playersInCurrentRound), Equals, 4)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 0)
	c.Assert(len(session.cards[player1.id]), Equals, 5)  //[d 5, s q, c q, s k, c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 0)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// cheat a bit and remove player1 cards s q, c q, s k
	session.cards[player1.id] = []string{"d 5", "c a"}
	err = session.playCards(player1, []string{"d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 1)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "d 5")
	c.Assert(len(session.cards[player1.id]), Equals, 1)  //[c a]
	c.Assert(len(session.cards[player2.id]), Equals, 10) //[s 4, c 5, d 6, c 8, h 8, c 9, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 1)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	//player 2 play
	err = session.playCards(player2, []string{"c 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 2)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "c 9")
	c.Assert(len(session.cards[player1.id]), Equals, 1)  //[c a]
	c.Assert(len(session.cards[player2.id]), Equals, 9)  //[s 4, c 5, d 6, c 8, h 8, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 2)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	//player 3 skip
	err = session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 2)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "c 9")
	c.Assert(len(session.cards[player1.id]), Equals, 1)  //[c a]
	c.Assert(len(session.cards[player2.id]), Equals, 9)  //[s 4, c 5, d 6, c 8, h 8, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 11) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, c k, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 3)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	//player 4 play
	err = session.playCards(player4, []string{"c k"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 3)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.cardsOnTable[0], Equals, "c k")
	c.Assert(len(session.cards[player1.id]), Equals, 1)  //[c a]
	c.Assert(len(session.cards[player2.id]), Equals, 9)  //[s 4, c 5, d 6, c 8, h 8, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 10) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, h a, s 2]]
	c.Assert(session.turnCounter, Equals, 4)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// //player1 play and win
	err = session.playCards(player1, []string{"c a"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.players), Equals, 4)
	c.Assert(len(session.playersInCurrentRound), Equals, 0)
	c.Assert(len(session.allMovesOfCurrentTurn), Equals, 4)
	c.Assert(len(session.cardsOnTable), Equals, 1)
	c.Assert(session.ownerOfCardsOnTable.Id(), Equals, player1.id)
	c.Assert(len(session.cards[player1.id]), Equals, 0)  // first rank
	c.Assert(len(session.cards[player2.id]), Equals, 9)  //[s 4, c 5, d 6, c 8, h 8, d 9, d 10,  h q, h k]
	c.Assert(len(session.cards[player3.id]), Equals, 11) //[c 4, h 4, h 6, s 7, h 7, s 8, s 10, c 10, d q, d a, c 2]
	c.Assert(len(session.cards[player4.id]), Equals, 10) //[s 5, c 6, c 7, d 7, s j, c j, d j, h j, h a, s 2]]

	// end game
	c.Assert(session.finished, Equals, true)
	c.Assert(len(session.results), Equals, 4)

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], true)
	mul4, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player4.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul4)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			c.Assert(len(utils.GetStringSliceAtPath(resultData, "cards")), Equals, 0)
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals,
				game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(len(utils.GetStringSliceAtPath(resultData, "cards")), Equals, 9)
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))

		} else if playerId == player3.id {
			c.Assert(len(utils.GetStringSliceAtPath(resultData, "cards")), Equals, 11)
			player3Gain := -moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, player3Gain)
		} else if playerId == player4.id {
			c.Assert(len(utils.GetStringSliceAtPath(resultData, "cards")), Equals, 10)
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			player4Gain := -moneyAfterApplyMultiplier(session.betEntry.Min(), mul4)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, player4Gain)
		}
	}

}

func (s *TestSuite) TestWinInStreak(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 7"}
	session.cards[player2.Id()] = []string{"s 4", "d 8"}
	session.cards[player3.Id()] = []string{"s 5", "d 9"}
	session.cards[player4.Id()] = []string{"c 6", "d 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id, player4.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 play
	err = session.playCards(player2, []string{"s 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player3 play
	err = session.playCards(player3, []string{"s 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player4.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player4 play
	err = session.playCards(player4, []string{"c 6"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 play
	err = session.playCards(player1, []string{"d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 4)

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], true)
	mul4, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player4.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul4)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))
		} else if playerId == player3.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul3))
		} else if playerId == player4.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul4))
		}
	}
}

func (s *TestSuite) Test3PeopleWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 7"}
	session.cards[player2.Id()] = []string{"s 4", "d 8"}
	session.cards[player3.Id()] = []string{"s 5", "d 9"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 play
	err = session.playCards(player2, []string{"s 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player3 play
	err = session.playCards(player3, []string{"s 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 play
	err = session.playCards(player1, []string{"d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals,
				game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals,
				-moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))
		} else if playerId == player3.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals,
				-moneyAfterApplyMultiplier(session.betEntry.Min(), mul3))
		}
	}
}

func (s *TestSuite) Test2PeopleWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 7"}
	session.cards[player2.Id()] = []string{"s 4", "d 8"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)
	for _, player := range []game.GamePlayer{player1, player2} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 play
	err = session.playCards(player2, []string{"s 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)
	for _, player := range []game.GamePlayer{player1, player2} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player1 play
	err = session.playCards(player1, []string{"d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2} {
		s.server.getAndRemoveResponse(player.Id())
	}

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))
		}
	}
}

func (s *TestSuite) TestWinWithFreeze(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(testFinishCallback.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 3", "c 3", "d 4", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "c 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "h 3", "c 3", "d 4", "h 10", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 9", "c 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id, player4.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// cheat
	session.cards[player1.Id()] = []string{"c 3"}
	session.cards[player2.Id()] = []string{"s 4", "d 8", "s 10", "c 10", "d 10", "h 10"}
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c k", "s k", "c k", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player4.Id()] = []string{"c 6"}

	// player1 play and win
	err = session.playCards(player1, []string{"c 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3, player4} {
		s.server.getAndRemoveResponse(player.Id())
	}
	// player3 freeze lose
	c.Assert(len(session.results), Equals, 4)

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], true)
	mul4, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player4.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul4)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			// win with freeze lose of player3, player3 has s2 c 2, and then player2 has s 10 c 10 d 10 h 10
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))
		} else if playerId == player3.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul3))
		} else if playerId == player4.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul4))
		}
	}
}

func (s *TestSuite) TestWinWithFreeze3People(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, len(session.players)))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "s 3", "c 3", "d 4", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "c 3", "d 4", "h 4", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c a", "c 3", "c 3", "d 4", "h 10", "s 5", "c 5", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// cheat
	session.cards[player1.Id()] = []string{"c 3"}
	session.cards[player2.Id()] = []string{"s 4", "d 8", "s 10", "c 10", "d 10", "h 10"}
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c k", "s k", "c k", "d 3", "h 4", "s 5", "c 6", "d 7", "c 8", "c 9", "h 10", "d j", "s q"})

	// player1 play and win,player3 freeze lose, and then game end
	err = session.playCards(player1, []string{"c 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	var gain int64
	mul2, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player2.id], true)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], true)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul2)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)

	for _, resultData := range session.results {
		playerId := utils.GetInt64AtPath(resultData, "id")
		if playerId == player1.id {
			// win with freeze lose of player3, player3 has s2 c 2, and then player2 has s 10 c 10 d 10 h 10
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, 0)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player2.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul2))
		} else if playerId == player3.id {
			c.Assert(utils.GetIntAtPath(resultData, "rank"), Equals, -1)
			c.Assert(utils.GetInt64AtPath(resultData, "change"), Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul3))
		}
	}
}

func (s *TestSuite) TestWinWillFreezeAndStuck(c *C) {
	currencyType := currency.Money
	var err error
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)

	result1 := game.NewGameResult("win", 100, 0, 0)
	result2 := game.NewGameResult("win", 100, 1, 0)
	result3 := game.NewGameResult("win", 100, 2, 0)
	session.lastMatchResult = map[int64]*game.GameResult{
		player1.id: result1,
		player2.id: result2,
		player3.id: result3,
	}

	// sample cards to play, no instant win
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"s 3", "s 8", "s 9", "c 9", "d 10", "h 10", "c j", "h j", "s q", "c q", "d q", "h q", "d k"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"d 3", "c 3", "s 4", "c 4", "d 4", "h 4", "s 5", "d 6", "d 9", "d j", "s k", "s 2", "h 2"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"h 5", "s 6", "c 6", "h 6", "c 7", "d 7", "h 7", "c 8", "d 8", "h 8", "s 10", "c k", "s a"})

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"s 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 2 3 skip
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"s 8", "s 9", "d 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player 2 3 skip
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1
	err = session.playCards(player1, []string{"c 9", "h 10", "c j"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player2 3
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1
	err = session.playCards(player1, []string{"h j", "s q", "d k"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player2 3
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1 play final cards
	err = session.playCards(player1, []string{"c q", "d q", "h q"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	err = session.playCards(player1, []string{"h a"})
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestWinWillFreezeRightDuringTurn(c *C) {
	currencyType := currency.Money
	var err error
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()
	player4 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3, player4})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType,
		[]game.GamePlayer{player1, player2, player3, player4}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 4))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)

	result1 := game.NewGameResult("win", 100, 0, 0)
	result2 := game.NewGameResult("win", 100, 1, 0)
	result3 := game.NewGameResult("win", 100, 2, 0)
	result4 := game.NewGameResult("win", 100, 3, 0)
	session.lastMatchResult = map[int64]*game.GameResult{
		player1.id: result1,
		player2.id: result2,
		player3.id: result3,
		player4.id: result4,
	}

	// sample cards to play, no instant win
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"s 3", "s 8", "s 9", "c 9", "d 10", "h 10", "c j", "h j", "s q", "c q", "d q", "h q", "d k"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"d 3", "c 3", "s 4", "c 4", "d 4", "h 4", "s 5", "d 6", "d 9", "d j", "s k", "c 2", "h 2"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"h 5", "s 6", "c 6", "h 6", "c 7", "d 7", "h 7", "c 8", "d 8", "h 8", "s 10", "c k"}) // remove 1 cards so player3 will not freeze lose
	session.cards[player4.Id()] = logic.SortCards(testGame.logicInstance, []string{"h 5", "s 6", "c 6", "h 6", "c 7", "d 7", "h 7", "c 8", "d 8", "h 8", "s 10", "c k"}) // remove 1 cards so player3 will not freeze lose

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"s 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 3 4
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"s 8", "s 9", "d 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	for _, player := range []game.GamePlayer{player1, player2, player3} {
		s.server.getAndRemoveResponse(player.Id())
	}

	// player2 3 4
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1
	err = session.playCards(player1, []string{"c 9", "h 10", "c j"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player2 3 4
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1
	err = session.playCards(player1, []string{"h j", "s q", "d k"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player2 3 4
	session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	session.skipTurn(player4)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	// player 1 play final cards, then player2 will freeze lose
	err = session.playCards(player1, []string{"c q", "d q", "h q"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)

	c.Assert(len(session.results), Equals, 4)
	c.Assert(session.finished, Equals, true)

	err = session.playCards(player2, []string{"h a"})
	c.Assert(err, NotNil)
}

func (s *TestSuite) TestRecord(c *C) {
	currencyType := currency.Money
	player1 := s.newPlayer()
	player2 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2})
	session.playersData = []*PlayerData{
		&PlayerData{
			id: player1.Id(),
		},
		&PlayerData{
			id: player2.Id(),
		},
	}
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 2))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"c 3", "s 5", "c 5", "d 5", "h 5", "s 7", "c 8", "h 9", "d q", "d k", "h a", "c 2", "d 2"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"c 6", "d 6", "c 7", "h 7", "s 8", "h 8", "c j", "d j", "s k", "c k", "s a", "c a", "h 2"})

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player2 will instant win
	c.Assert(session.finished, Equals, true)

	var gain int64
	mul1, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player1.id], false)
	gain = moneyAfterApplyMultiplier(session.betEntry.Min(), mul1)

	resultsData := session.results
	c.Assert(resultsData, NotNil)
	for _, resultData := range resultsData {
		playerId := utils.GetInt64AtPath(resultData, "id")
		result := utils.GetStringAtPath(resultData, "result")
		change := utils.GetInt64AtPath(resultData, "change")
		cards := utils.GetStringSliceAtPath(resultData, "cards")
		c.Assert(len(cards) != 0, Equals, true)
		if playerId == player1.id {
			c.Assert(result, Equals, "lose")
			c.Assert(change, Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul1)) // lose the bet on table
		} else if playerId == player2.id {
			c.Assert(result, Equals, "win")
			c.Assert(utils.GetStringAtPath(resultData, "win_type"), Equals, "instant")
			c.Assert(change, Equals, game.MoneyAfterTax(gain, session.betEntry))
		}
	}

	c.Assert(session.tax+session.botWin, Equals, session.botLose)
}

func (s *TestSuite) TestRecordNormally(c *C) {
	currencyType := currency.Money
	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})

	session.playersData = []*PlayerData{
		&PlayerData{
			id: player1.Id(),
		},
		&PlayerData{
			id: player2.Id(),
		},
		&PlayerData{
			id: player3.Id(),
		},
	}
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	session.cards[player1.Id()] = logic.SortCards(testGame.logicInstance, []string{"s 3", "c 6", "d 6", "h 6", "s 7", "c 7", "h 7", "s 8", "c 8", "h 8", "d 10", "s q", "h k"})
	session.cards[player2.Id()] = logic.SortCards(testGame.logicInstance, []string{"d 3", "s 4", "c 4", "d 4", "h 4", "d 7", "d j", "s k", "c k", "s 2", "c 2", "d 2", "h 2"})
	session.cards[player3.Id()] = logic.SortCards(testGame.logicInstance, []string{"c 9", "d 9", "c 10", "h 10", "s j", "h j", "c q", "h q", "c a", "s 3", "s 4", "d 5", "h 7"})

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player2 will instant win
	c.Assert(session.finished, Equals, true)

	var gain int64
	mul1, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player1.id], false)
	mul3, _ := testGame.logicInstance.LoseMultiplierByCardLeft(session.cards[player3.id], false)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul1)
	gain = gain + moneyAfterApplyMultiplier(session.betEntry.Min(), mul3)

	resultsData := session.results
	c.Assert(resultsData, NotNil)
	fmt.Println(resultsData)
	for _, resultData := range resultsData {
		playerId := utils.GetInt64AtPath(resultData, "id")
		result := utils.GetStringAtPath(resultData, "result")
		change := utils.GetInt64AtPath(resultData, "change")
		cards := utils.GetStringSliceAtPath(resultData, "cards")
		c.Assert(len(cards) != 0, Equals, true)
		if playerId == player1.id {
			c.Assert(result, Equals, "lose")
			c.Assert(change, Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul1))
		} else if playerId == player2.id {
			c.Assert(result, Equals, "win")
			c.Assert(utils.GetStringAtPath(resultData, "win_type"), Equals, "instant")
			c.Assert(change, Equals, game.MoneyAfterTax(gain, session.betEntry))
		} else if playerId == player3.id {
			c.Assert(result, Equals, "lose")
			c.Assert(change, Equals, -moneyAfterApplyMultiplier(session.betEntry.Min(), mul3))
		}
	}
}

func (s *TestSuite) TestNewRoundAfterWin3PlayerMidOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h 4"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 6", "h 6"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 5", "d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 skip
	err = session.skipTurn(player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.Id())

	// player2 play and win
	err = session.playCards(player2, []string{"d 6", "h 6"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestNewRoundAfterWin3PlayerLastOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h 7", "d 7", "s a"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 8", "h 8"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10", "h 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 5", "d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play
	err = session.playCards(player1, []string{"h 7", "d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.Id())

	// player2 skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.Id())

	// player3 play and win
	err = session.playCards(player3, []string{"c 10", "h 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestNewRoundAfterWin3PlayerFirstOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h 7", "d 7"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 8", "h 8"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10", "h 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 skip
	err = session.skipTurn(player3)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play and win
	err = session.playCards(player1, []string{"h 7", "d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestNewRoundAfterWinNoSkip3PlayerMidOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h 8", "d 8", "s a"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 9", "h 9"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 5", "d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play
	err = session.playCards(player1, []string{"h 8", "d 8"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.Id())

	// player2 play and win
	err = session.playCards(player2, []string{"d 9", "h 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestNewRoundAfterWinNoSkip3PlayerLastOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h 7", "d 7", "s a"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 8", "h 8", "s k"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10", "h 10"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 5", "d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play
	err = session.playCards(player1, []string{"h 7", "d 7"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.Id())

	// player2 play
	err = session.playCards(player2, []string{"d 8", "h 8"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.Id())

	// player3 play and win
	err = session.playCards(player3, []string{"c 10", "h 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestNewRoundAfterWinNoSkip3PlayerFirstOrderWin(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"c 3", "d 3", "h j", "d j"}
	session.cards[player2.Id()] = []string{"s 4", "d 4", "d 8", "h 8"}
	session.cards[player3.Id()] = []string{"s 5", "d 5", "c 10", "h 10", "d a"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player1 play
	err = session.playCards(player1, []string{"c 3", "d 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 play
	err = session.playCards(player2, []string{"s 4", "d 4"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"c 10", "h 10"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play and win
	err = session.playCards(player1, []string{"h j", "d j"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)
	c.Assert(session.finished, Equals, true)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

func (s *TestSuite) TestInactiveWholeGame(c *C) {
	currencyType := currency.Money
	var err error

	player1 := s.newPlayer()
	player2 := s.newPlayer()
	player3 := s.newPlayer()

	testGame := NewTienLenGame(currencyType)
	testGame.turnTimeInSeconds = 5 * time.Second
	session := NewTienLenSession(testGame, currencyType, player1, []game.GamePlayer{player1, player2, player3})
	testFinishCallback := NewTestFinishCallback(s.server, currencyType, []game.GamePlayer{player1, player2, player3}, player1)
	testFinishCallback.setMoneysOnTable(testGame.MoneyOnTable(100, 4, 3))
	session.sessionCallback = testFinishCallback
	session.betEntry = testGame.BetData().GetEntry(100)

	result1 := game.NewGameResult("win", 100, 3, 0)
	result2 := game.NewGameResult("win", 100, 2, 0)
	result3 := game.NewGameResult("win", 100, 1, 0)
	session.lastMatchResult = map[int64]*game.GameResult{
		player1.id: result1,
		player2.id: result2,
		player3.id: result3,
	}

	session.cards = make(map[int64][]string)
	// sample cards to play, no instant win

	session.cards[player1.Id()] = []string{"d 3", "c 3", "c 4", "d 4", "d 5", "s 6", "d 6", "d 8", "d 9", "c 10", "h 10", "d q", "h q"}
	session.cards[player2.Id()] = []string{"s 5", "h 5", "c 6", "h 6", "s 7", "c 7", "h 8", "h 9", "d 10", "h j", "h k", "d a", "h a"}
	session.cards[player3.Id()] = []string{"s 3", "h 7", "c 8", "s 9", "c 9", "d j", "s q", "s k", "s a", "c a", "s 2", "c 2", "d 2"}

	fmt.Printf("players %d %d %d %d \n", player1.id, player2.id, player3.id)

	s.server.cleanupAllResponse()
	session.start()
	utils.DelayInDuration(waitTimeForTurn)

	// player3 play
	err = session.playCards(player3, []string{"s 3"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play
	err = session.playCards(player1, []string{"d 5"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 play
	err = session.playCards(player1, []string{"d 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"d 2"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 skip
	err = session.skipTurn(player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"h 7", "c 8", "c 9"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 skip
	err = session.skipTurn(player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"d j", "s q", "s k"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 skip
	err = session.skipTurn(player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s a", "c a"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player1.id)

	// player1 skip
	err = session.skipTurn(player1)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player2.id)

	// player2 skip
	err = session.skipTurn(player2)
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(session.currentPlayerTurn.Id(), Equals, player3.id)

	// player3 play
	err = session.playCards(player3, []string{"s 2", "c 2"})
	c.Assert(err, IsNil)
	utils.DelayInDuration(waitTimeForTurn)
	c.Assert(len(session.results), Equals, 3)

	session.timeOutForTurn.SetShouldHandle(false) // stop the game
}

/*
helper
*/

type TestFinishSessionCallback struct {
	currencyType string
	server       *TestServer
	didFinish    bool
	players      []game.GamePlayer
	owner        game.GamePlayer

	moneysOnTable map[int64]int64
}

func NewTestFinishCallback(server *TestServer, currencyType string, players []game.GamePlayer, owner game.GamePlayer) *TestFinishSessionCallback {
	return &TestFinishSessionCallback{
		currencyType: currencyType,
		didFinish:    false,
		server:       server,
		players:      players,
		owner:        owner,
	}
}

func (callback *TestFinishSessionCallback) setMoneysOnTable(money int64) {
	callback.moneysOnTable = make(map[int64]int64)
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.moneysOnTable[playerId] = money
	}

	for _, player := range callback.players {
		playerHere := player.(*TestPlayer)
		playerHere.DecreaseMoney(money, callback.currencyType, true)
	}
}

func (callback *TestFinishSessionCallback) AssignOwner(player game.GamePlayer) (err error) {
	return nil
}

func (callback *TestFinishSessionCallback) RemoveOwner() (err error) {
	return nil
}

func (callback *TestFinishSessionCallback) GetPlayerAtIndex(index int) game.GamePlayer {
	for indexInList, player := range callback.players {
		if index == indexInList {
			return player
		}
	}
	return nil
}

func (callback *TestFinishSessionCallback) DidStartGame(session game.GameSessionInterface) {
	method := fmt.Sprintf("%s_start_game_session", "tienlen")
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.server.SendRequest(method, session.SerializedDataForPlayer(session.GetPlayer(playerId)), playerId)
	}

}

func (callback *TestFinishSessionCallback) DidChangeGameState(session game.GameSessionInterface) {
	method := fmt.Sprintf("%s_change_game_session", "tienlen")
	for _, playerId := range getIdFromPlayersMap(callback.players) {
		callback.server.SendRequest(method, session.SerializedDataForPlayer(session.GetPlayer(playerId)), playerId)
	}
}

func (callback *TestFinishSessionCallback) DidEndGame(result map[string]interface{}, delaySeconds int) {
	method := fmt.Sprintf("%s_finish_game_session", "tienlen")
	callback.server.SendRequests(method, result, getIdFromPlayersMap(callback.players))

	callback.didFinish = true
	for _, player := range callback.players {
		player.IncreaseMoney(callback.moneysOnTable[player.Id()], callback.currencyType, true)
	}
	callback.moneysOnTable = make(map[int64]int64)
}

func (callback *TestFinishSessionCallback) MoneyDidChange(session game.GameSessionInterface, playerId int64, change int64, reason string, additionalData map[string]interface{}) {

}
func (callback *TestFinishSessionCallback) SendNotifyMoneyChange(playerId int64, change int64, reason string, additionalData map[string]interface{}) {
}

func (callback *TestFinishSessionCallback) SendMessageToPlayer(session game.GameSessionInterface, playerId int64, method string, data map[string]interface{}) {

}

func (callback *TestFinishSessionCallback) Owner() game.GamePlayer {
	return callback.owner
}

func (callback *TestFinishSessionCallback) SetMoneyOnTable(playerId int64, value int64, shouldLock bool) (err error) {
	callback.moneysOnTable[playerId] = value
	return nil
}

func (callback *TestFinishSessionCallback) IncreaseMoney(playerInstance game.GamePlayer, amount int64, shouldBlock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}

	// normal card game, money add decrease will first use money on table
	// increase
	callback.moneysOnTable[playerInstance.Id()] += amount
	return nil
}

func (callback *TestFinishSessionCallback) IncreaseAndFreezeThoseMoney(playerInstance game.GamePlayer, amount int64, shouldLock bool) (err error) {
	if amount < 0 {
		// decrease
		return errors.New("err:increase_negative")
	}
	callback.moneysOnTable[playerInstance.Id()] += amount
	return err
}

func (callback *TestFinishSessionCallback) DecreaseMoney(playerInstance game.GamePlayer, amount int64, shouldBlock bool) (err error) {
	if amount < 0 {
		return errors.New("err:decrease_negative")
	}
	// decrease
	callback.moneysOnTable[playerInstance.Id()] -= amount
	if callback.moneysOnTable[playerInstance.Id()] < 0 {
		decreaseFromPlayer := -callback.moneysOnTable[playerInstance.Id()]
		callback.moneysOnTable[playerInstance.Id()] = 0
		_, err = playerInstance.DecreaseMoney(decreaseFromPlayer, callback.currencyType, shouldBlock)
		return err
	}
	return nil
}

func (callback *TestFinishSessionCallback) MoveMoneyFromPlayerToTable(playerInstance game.GamePlayer, amount int64, shouldNotify bool, shouldBlock bool) {
	if amount > 0 {
		callback.moneysOnTable[playerInstance.Id()] += amount
	}
}

func (callback *TestFinishSessionCallback) MoveMoneyFromTableToPlayer(playerInstance game.GamePlayer, amount int64, shouldNotify bool, shouldBlock bool) {
	if amount > 0 {
		callback.moneysOnTable[playerInstance.Id()] -= amount
		if callback.moneysOnTable[playerInstance.Id()] < 0 {
			playerInstance.DecreaseMoney(-callback.moneysOnTable[playerInstance.Id()], callback.currencyType, shouldBlock)
			callback.moneysOnTable[playerInstance.Id()] = 0
		}
	}
}

func (callback *TestFinishSessionCallback) GetTotalPlayerMoney(playerId int64) int64 {
	playerInstance := callback.getPlayer(playerId)
	if playerInstance == nil {
		return 0
	}
	return playerInstance.GetMoney(callback.currencyType) + callback.moneysOnTable[playerId]
}

func (callback *TestFinishSessionCallback) GetMoneyOnTable(playerId int64) int64 {
	return callback.moneysOnTable[playerId]
}

func (callback *TestFinishSessionCallback) getPlayer(playerId int64) (player game.GamePlayer) {
	for _, player := range callback.players {
		if player.Id() == playerId {
			return player
		}
	}
	return nil
}

func (callback *TestFinishSessionCallback) GetPlayersDataForDisplay(currentPlayer game.GamePlayer) (playersData map[string]map[string]interface{}) {
	return nil
}

type TestPlayer struct {
	currencyGroup *currency.CurrencyGroup
	id            int64
	name          string
	room          *game.Room
	isOnline      bool
	exp           int64
	bet           int64

	recentResult   string
	recentChange   int64
	recentGameCode string

	playerType string

	gameCount int

	moneyMutex sync.Mutex
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
		playerType:    "bot",
		isOnline:      true,
	}
	testPlayer.IncreaseMoney(100000, currency.Money, true)
	testPlayer.IncreaseMoney(100000, currency.TestMoney, true)
	return testPlayer
}

func (player *TestPlayer) LockMoney(currencyType string) {
	player.currencyGroup.Lock(currencyType)
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
	player.currencyGroup.IncreaseMoney(money-player.GetMoney(currencyType), currencyType, true)
}

func (player *TestPlayer) Id() int64 {
	return player.id
}
func (player *TestPlayer) Name() string {
	return player.name
}

func (player *TestPlayer) Room() *game.Room {
	return player.room
}

func (player *TestPlayer) SetRoom(room *game.Room) {
	player.room = room
}

func (player *TestPlayer) IpAddress() string {
	return ""
}

func (player *TestPlayer) IsOnline() bool {
	return player.isOnline
}

func (player *TestPlayer) SetIsOnline(isOnline bool) {
	player.isOnline = isOnline
}
func (player *TestPlayer) IncreaseBet(bet int64) {
	player.bet += bet
}

func (player *TestPlayer) SerializedData() map[string]interface{} {
	return map[string]interface{}{
		"id": player.id,
	}
}

func (player *TestPlayer) SerializedDataMinimal() map[string]interface{} {
	return map[string]interface{}{
		"id": player.id,
	}
}
func (player *TestPlayer) IncreaseVipPointForMatch(bet int64, matchId int64, gameCode string) {
}

func (player *TestPlayer) IncreaseExp(exp int64) (newExp int64, err error) {
	player.exp = player.exp + exp
	return player.exp, nil
}

func (player *TestPlayer) RecordGameResult(gameCode string, result string, change int64, currencyType string) (err error) {
	player.recentResult = result
	player.recentChange = change
	player.recentGameCode = gameCode
	player.gameCount += 1
	return nil
}

func (player *TestPlayer) PlayerType() string {
	return player.playerType
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

func (server *TestServer) getNumberOfResponse(playerId int64) int {
	return len(server.receiveDataMap[playerId])
}

func (server *TestServer) getAndRemoveResponse(toPlayerId int64) map[string]interface{} {
	fullData := server.receiveDataMap[toPlayerId][0]
	server.receiveDataMap[toPlayerId] = server.receiveDataMap[toPlayerId][1:]
	return fullData
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

// func (s *TestSuite) justCreateRoom(game GameInterface) {
// 	player := &TestPlayer{
// 		id:    s.genPlayerId(),
// 		name:  "test2",
// 		money: 10000,
// 	}
// 	CreateRoom(game, player, 10, 4)
// }
