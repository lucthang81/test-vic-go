package money

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/test"
	"github.com/vic/vic_go/utils"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter *datacenter.DataCenter
	dbName     string
	server     *TestServer
	adminId    int64
}

var _ = Suite(&TestSuite{
	dbName: "casino_money_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	feature.UnlockAllFeature()
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabaseWithError(s.dbName, []string{"../../sql/init_schema.sql", "../../sql/test_schema/player_test.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	s.dataCenter.FlushCache()
	RegisterDataCenter(s.dataCenter)
	currency.RegisterDataCenter(s.dataCenter)
	player.RegisterDataCenter(s.dataCenter)
	record.RegisterDataCenter(s.dataCenter)

	version := &TestVersion{
		version: "1.3",
	}
	RegisterVersion(version)

	s.generateAdmin(c)

	s.server = NewTestServer()
	player.RegisterServer(s.server)
	player.RegisterVersion(version)

}

func (s *TestSuite) TearDownSuite(c *C) {
	test.DropTestDatabase(s.dbName)
}

func (s *TestSuite) SetUpTest(c *C) {
	// Use s.dir to prepare some data.

	fmt.Printf("start test %s \n", c.TestName())
}

func (s *TestSuite) TearDownTest(c *C) {
	s.dataCenter.FlushCache()
}

/*



THE ACTUAL TESTS



*/

func (s *TestSuite) TestRunnable(c *C) {
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestHide(c *C) {
	c.Assert(hideString("1234567889"), Equals, "12345xxxxx")
	c.Assert(hideString("8943078473"), Equals, "89430xxxxx")
	c.Assert(hideString("753489758936242"), Equals, "7534897589xxxxx")
}
func (s *TestSuite) TestCreateCard(c *C) {
	currencyType := currency.Money
	version := &TestVersion{
		version: "1.2",
	}
	RegisterVersion(version)
	err := CreateCardType("mobi_100", 1000)
	c.Assert(err, IsNil)

	err = CreateCard("mobi", "mobi_101", "dfaskjl", "fdskljfalsdk")
	c.Assert(err.Error(), Equals, "err:card_code_not_exist_mobi_101")

	err = CreateCard("mobi", "mobi_100", "dfaskjl", "fdskljfalsdk")
	c.Assert(err, IsNil)

	c.Assert(getCard("mobi_100"), NotNil)

	playerInstance, _ := generateNewPlayerWithRandomName()
	playerInstance.IncreaseMoney(10000-playerInstance.GetMoney(currencyType), currencyType, true)

	paymentRequirement.minMoneyLeftAfterPayment = 5000

	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase_type,purchase) VALUES ($1,$2,$3)", playerInstance.Id(), "iap", 1000)
	c.Assert(err, IsNil)
	playerInstance.IncreaseBet(1000000)

	_, err = requestPayment(playerInstance, "mobi_20")
	c.Assert(err.Error(), Equals, "err:wrong_card_code")

	oldMoney := playerInstance.GetMoney(currencyType)
	paymentId, err := requestPayment(playerInstance, "mobi_100")
	c.Assert(err, IsNil)
	c.Assert(oldMoney-playerInstance.GetMoney(currencyType), Equals, int64(1000))

	err = acceptPayment(s.adminId, paymentId)
	c.Assert(err, IsNil)
	c.Assert(getCard("mobi_100"), IsNil)

	oldMoney = playerInstance.GetMoney(currencyType)

	paymentId, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err, IsNil)
	c.Assert(oldMoney-playerInstance.GetMoney(currencyType), Equals, int64(1000))

	err = acceptPayment(s.adminId, paymentId)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:not_enough_card")

	playerInstance.IncreaseMoney(5000-playerInstance.GetMoney(currencyType), currencyType, true)
	paymentId, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err.Error(), Equals, "err:not_enough_money_for_payment")
	c.Assert(playerInstance.GetMoney(currencyType), Equals, int64(5000))

	paymentRequirement.minMoneyLeftAfterPayment = 1000
	paymentId, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err, IsNil)
	c.Assert(playerInstance.GetMoney(currencyType), Equals, int64(4000))

	err = declinePayment(s.adminId, paymentId)
	c.Assert(err, IsNil)
}

func (s *TestSuite) TestPaymentRequirement(c *C) {
	currencyType := currency.Money
	version := &TestVersion{
		version: "1.3",
	}
	RegisterVersion(version)
	err := CreateCardType("mobi_101", 1000)
	c.Assert(err, IsNil)

	playerInstance, _ := generateNewPlayerWithRandomName()
	playerInstance.IncreaseMoney(1500-playerInstance.GetMoney(currencyType), currencyType, true)

	_, err = requestPayment(playerInstance, "mobi_20")
	c.Assert(err.Error(), Equals, "err:wrong_card_code")

	paymentRequirement.minMoneyLeftAfterPayment = 800

	_, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err.Error(), Equals, "err:not_enough_money_for_payment")

	paymentRequirement.minMoneyLeftAfterPayment = 100
	_, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err.Error(), Equals, "Lần cuối nạp thẻ đã quá lâu")

	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase_type) VALUES ($1,$2)", playerInstance.Id(), "iap")
	c.Assert(err, IsNil)

	paymentRequirement.minTotalBet = 50
	_, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err.Error(), Equals, "Tổng cược đã đặt trong game không đủ")

	playerInstance.IncreaseBet(60)
	_, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err.Error(), Equals, fmt.Sprintf("Không thể đổi thưởng quá x%d số tiền bạn đã nạp", paymentRequirement.purchaseMultiplier))

	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase, purchase_type) VALUES ($1,$2,$3)", playerInstance.Id(), 10000, "iap")
	c.Assert(err, IsNil)
	_, err = requestPayment(playerInstance, "mobi_100")
	c.Assert(err, IsNil)

}

func (s *TestSuite) TestConcurrentProcess(c *C) {
	currencyType := currency.Money
	version := &TestVersion{
		version: "1.2",
	}
	RegisterVersion(version)
	err := CreateCardType("mobi_10", 1000)
	c.Assert(err, IsNil)
	playerInstance1, _ := generateNewPlayerWithRandomName()
	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase_type,purchase) VALUES ($1,$2,$3)", playerInstance1.Id(), "iap", 1000)
	c.Assert(err, IsNil)
	playerInstance1.IncreaseBet(1000000)
	playerInstance1.IncreaseMoney(10000-playerInstance1.GetMoney(currencyType), currencyType, true)

	playerInstance2, _ := generateNewPlayerWithRandomName()
	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase_type,purchase) VALUES ($1,$2,$3)", playerInstance2.Id(), "iap", 1000)
	c.Assert(err, IsNil)
	playerInstance2.IncreaseBet(1000000)
	playerInstance2.IncreaseMoney(10000-playerInstance2.GetMoney(currencyType), currencyType, true)

	playerInstance3, _ := generateNewPlayerWithRandomName()
	_, err = dataCenter.Db().Exec("INSERT INTO purchase_record (player_id, purchase_type,purchase) VALUES ($1,$2,$3)", playerInstance3.Id(), "iap", 1000)
	c.Assert(err, IsNil)
	playerInstance3.IncreaseBet(1000000)
	playerInstance3.IncreaseMoney(10000-playerInstance3.GetMoney(currencyType), currencyType, true)

	paymentRequirement.minMoneyLeftAfterPayment = 5000

	go RequestPayment(playerInstance1, "mobi_10")
	go RequestPayment(playerInstance2, "mobi_10")
	go RequestPayment(playerInstance3, "mobi_10")

	utils.Delay(1)
	c.Assert(playerInstance1.GetMoney(currencyType), Equals, int64(9000))
	c.Assert(playerInstance2.GetMoney(currencyType), Equals, int64(9000))
	c.Assert(playerInstance3.GetMoney(currencyType), Equals, int64(9000))

}

/*
helper
*/

func generateNewPlayerWithRandomName() (playerInstance *player.Player, err error) {
	return player.GenerateNewPlayer(player.GenerateRandomValidPlayerUsername(), player.GenerateRandomPlayerIdentifier(), utils.RandSeq(10), "abc")
}

func (s *TestSuite) generateAdmin(c *C) {
	queryString := "INSERT INTO admin_account (username, password) VALUES ($1,$2) RETURNING id"
	row := s.dataCenter.Db().QueryRow(queryString, "admin", "fdsakljfals")
	err := row.Scan(&s.adminId)
	c.Assert(err, IsNil)
}

type TestVersion struct {
	version string
}

func (version *TestVersion) GetVersion() string {
	return version.version
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

func (server *TestServer) SendHotFixRequest(requestType string, data map[string]interface{}, currencyType string, toPlayerId int64) {

}
func (server *TestServer) LogoutPlayer(playerId int64) {
	return
}
