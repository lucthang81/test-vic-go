package player

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game"
	"github.com/vic/vic_go/models/game/ceme"
	"github.com/vic/vic_go/models/game/tienlen"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/test"
	"github.com/vic/vic_go/utils"
	// "github.com/vic/vic_go/utils"

	// "github.com/vic/vic_go/log"
	. "gopkg.in/check.v1"
	"math/rand"
	"testing"
	"time"
	// "log"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter *datacenter.DataCenter
	server     *TestServer
	dbName     string
}

var _ = Suite(&TestSuite{
	dbName: "casino_player_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	feature.UnlockAllFeature()
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../sql/init_schema.sql", "../../sql/test_schema/player_test.sql", "../../sql/test_schema/leaderboard_test.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	RegisterDataCenter(s.dataCenter)
	currency.RegisterDataCenter(s.dataCenter)
	game.RegisterDataCenter(s.dataCenter)
	record.RegisterDataCenter(s.dataCenter)

	s.server = NewTestServer()
	RegisterServer(s.server)

	// xidachGame := xidach.NewXiDachGame(currency.Money)
	// RegisterGame(xidachGame)

	cemeGame := ceme.NewCemeGame(currency.Money)
	RegisterGame(cemeGame)

	tienlenGame := tienlen.NewTienLenGame(currency.Money)
	RegisterGame(tienlenGame)

	testVersion := &TestVersion{
		version: "1.3",
	}
	RegisterVersion(testVersion)
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
	c.Assert(false, Equals, false)
}

func (s *TestSuite) TestGenerateNewPlayerWithRandomName(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player, NotNil)

	// check achievement
	c.Assert(len(player.achievementManager.achievements), Equals, 4)

	player1, err := GenerateNewPlayer("HênXui289", "dsfa2333", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player1, NotNil)

	player2, err := GenerateNewPlayer("VôHếtTiền2", "dsfa2333", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player2, NotNil)
}

func (s *TestSuite) TestGenNameFacebook(c *C) {
	isTesting = true
	_, player, err := AuthenticatePlayerByFacebook("abc", "1", "GiaDangggggggggg hihihihi", "avatar", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.username, Equals, "GiaDangggggggggg hihihihi")
	deviceIdentifier := player.deviceIdentifier
	_, player, err = AuthenticatePlayerByFacebook("abc", "2", "GiaDangggggggggg hihihihi", "avatar", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.username, Equals, "GiaDangggggggggg hihihihi (fb)")
	_, player, err = AuthenticatePlayerByFacebook("abc", "3", "GiaDangggggggggg hihihihi", "avatar", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.username, Equals, "GiaDangggggggggg hihihihi (fb1)")
	_, player, err = AuthenticatePlayerByFacebook("abc", "1", "GiaDangggggggggg hihihihi", "avatar", deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	_, player, err = AuthenticatePlayerByFacebook("abc", "4", "GiaDangggggggggg hihihihi", "avatar", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.username, Equals, "GiaDangggggggggg hihihihi (fb2)")
	_, player, err = AuthenticatePlayerByFacebook("abc", "5", "Gia {}[}fs D", "avatar", utils.RandSeq(10), "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.username, Equals, "Gia {}[}fs D")
}

func (s *TestSuite) TestVerifyName(c *C) {
	startTime := time.Now()
	err := verifyUsername("ahihidocho")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	err = verifyUsername("ahihidocho1")
	endTime := time.Now()
	c.Assert(err, IsNil)
	fmt.Println("time", endTime.Sub(startTime).String())
	// c.Assert(true, Equals, false)
}

func (s *TestSuite) TestAuthPlayer(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player, NotNil)
	deviceIdentifier := player.deviceIdentifier

	player, err = AuthenticateOldPlayer(player.Identifier(), player.Token(), player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)

	// _, err = AuthenticateOldPlayer(player.Identifier(), player.Token(), "xyz", "bighero")
	// c.Assert(err.Error(), Equals, "err:login_multiple_devices")
	// c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)

	player, password, err := generateNewPlayerWithRandomNameAndPassword()
	c.Assert(err, IsNil)
	deviceIdentifier = player.deviceIdentifier
	player, err = AuthenticateOldPlayerByPassword(player.username, password, deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)

	_, err = AuthenticateOldPlayerByPassword(player.username, "123333333", utils.RandSeq(10), "bighero")
	c.Assert(err.Error(), Equals, "err:invalid_password")

	player, err = GenerateNewPlayer(player.username, "123432989", utils.RandSeq(10), "bighero")
	c.Assert(err.Error(), Equals, l.Get(l.M0073))

	player, err = GenerateNewPlayer(GenerateRandomPlayerIdentifier(), "1", utils.RandSeq(10), "bighero")
	c.Assert(err.Error(), Equals, l.Get(l.M0076))

	player, password, err = generateNewPlayerWithRandomNameAndPassword()
	deviceIdentifier = player.deviceIdentifier
	player, err = AuthenticateOldPlayerByPassword(player.username, password, player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	token := player.Token()

	// _, err = AuthenticateOldPlayer(player.Identifier(), token, "efg", "bighero")
	// c.Assert(err.Error(), Equals, "err:login_multiple_devices")
	// c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)
	// _, err = AuthenticateOldPlayerByPassword(player.username, password, "fsadkfjal", "bighero")
	// c.Assert(err.Error(), Equals, "err:login_multiple_devices")
	// c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)

	player, err = AuthenticateOldPlayer(player.Identifier(), token, player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)

	player, err = AuthenticateOldPlayerByPassword(player.username, password, player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.Token() != token, Equals, true)

	_, err = AuthenticateOldPlayer(player.Identifier(), token, player.deviceIdentifier, "bighero")
	c.Assert(err.Error(), Equals, l.Get(l.M0068))

	_, err = AuthenticateOldPlayer(player.Identifier(), player.Token(), player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)

	token = player.Token()
	player.CleanUpAndLogout()

	_, err = AuthenticateOldPlayer(player.Identifier(), token, player.deviceIdentifier, "bighero")
	c.Assert(err.Error(), Equals, l.Get(l.M0068))

	player, err = AuthenticateOldPlayerByPassword(player.username, password, player.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	c.Assert(player.Token() != token, Equals, true)

	player1, _ := generateNewPlayerWithRandomName()
	_, err = AuthenticateOldPlayer(player1.Identifier(), player1.Token(), player1.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)

	player2, _ := generateNewPlayerWithRandomName()
	_, err = AuthenticateOldPlayer(player2.Identifier(), player2.Token(), player2.deviceIdentifier, "bighero")
	c.Assert(err, IsNil)
	err = player1.CleanUpAndLogout()
	c.Assert(err, IsNil)
	err = player2.CleanUpAndLogout()
	c.Assert(err, IsNil)
}

// func (s *TestSuite) TestCannotRegisterSameDevice(c *C) {
// 	player, err := generateNewPlayerWithRandomNameAndDeviceIdentifier("abcxyz")
// 	c.Assert(err, IsNil)
// 	c.Assert(player, NotNil)
// 	deviceIdentifier := player.deviceIdentifier

// 	player, err = AuthenticateOldPlayer(player.Identifier(), player.Token(), player.deviceIdentifier, "bighero")
// 	c.Assert(err, IsNil)
// 	c.Assert(player.deviceIdentifier, Equals, deviceIdentifier)

// 	_, err = generateNewPlayerWithRandomNameAndDeviceIdentifier("abcxyz")
// 	c.Assert(err.Error(), Equals, "Một thiết bị không được dùng để đăng ký nhiều tài khoản")
// }

func (s *TestSuite) TestGetPlayerData(c *C) {
	currencyType := currency.Money

	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player, NotNil)

	player.IncreaseMoney(1000-player.GetMoney(currencyType), currencyType, true)
	data := player.SerializedData()
	c.Assert(utils.GetInt64AtPath(data, fmt.Sprintf("currency/%s", currencyType)), Equals, int64(1000))
}

func (s *TestSuite) TestUpdateUsername(c *C) {
	player1, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player1, NotNil)

	err = player1.UpdateUsername("hihihi")
	c.Assert(err, IsNil)

	player2, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player2, NotNil)

	err = player2.UpdateUsername("hihihi")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, l.Get(l.M0073))

	err = player2.UpdateUsername("i")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, l.Get(l.M0074))

	err = player2.UpdateUsername("i     ")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, l.Get(l.M0074))

	err = player2.UpdateUsername("giá gia")
	c.Assert(err, IsNil)

	err = player2.UpdateUsername("giá gia192")
	c.Assert(err, IsNil)

	err = player2.UpdateUsername("giá -_.gia")
	c.Assert(err, IsNil)

	err = player2.UpdateUsername("giá+gia")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, l.Get(l.M0075))
}

func (s *TestSuite) TestUpdateEmail(c *C) {
	player1, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player1, NotNil)

	err = player1.UpdateEmail("hihihi@haha.com")
	c.Assert(err, IsNil)

	player2, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player2, NotNil)

	err = player2.UpdateEmail("hihihi@haha.com")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:duplicate_email")

	err = player2.UpdateEmail("giá@kjd@fdaskl")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:invalid_email_format")

	err = player2.UpdateEmail("gi@kjd@fdaskl")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:invalid_email_format")

	err = player2.UpdateEmail("gfdsakj29890")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:invalid_email_format")

	err = player2.UpdateEmail("gfds$#ak@j29.890")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:invalid_email_format")

	err = player2.UpdateEmail("ds@x93.f9e")
	c.Assert(err, IsNil)

	err = player2.UpdateEmail("d_s@x93.f9e")
	c.Assert(err, IsNil)

	err = player2.UpdateEmail("d_fdsafa.ss@x93.f9e")
	c.Assert(err, IsNil)
}

func (s *TestSuite) TestAchievement(c *C) {
	currencyType := currency.Money
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player, NotNil)

	// check achievement
	c.Assert(len(player.achievementManager.achievements), Equals, 4) // 2 type of currency

	player.RecordGameResult("ceme", "win", 100, currencyType)
	achievement := player.achievementManager.getAchievement("ceme", currencyType)
	c.Assert(achievement.gameCode, Equals, "ceme")
	c.Assert(achievement.biggestWin, Equals, int64(100))
	c.Assert(achievement.biggestWinThisWeek, Equals, int64(100))
	c.Assert(achievement.totalGainThisWeek, Equals, int64(100))
	player.RecordGameResult("ceme", "win", 50, currencyType)
	achievement = player.achievementManager.getAchievement("ceme", currencyType)
	c.Assert(achievement.biggestWin, Equals, int64(100))
	c.Assert(achievement.biggestWinThisWeek, Equals, int64(100))
	c.Assert(achievement.totalGainThisWeek, Equals, int64(150))
	c.Assert(achievement.winCount, Equals, 2)

	player.RecordGameResult("ceme", "lose", 50, currencyType)
	player.RecordGameResult("ceme", "lose", 50, currencyType)
	player.RecordGameResult("ceme", "draw", 50, currencyType)
	player.RecordGameResult("ceme", "lose", 50, currencyType)

	c.Assert(achievement.totalGainThisWeek, Equals, int64(350))
	c.Assert(achievement.biggestWinThisWeek, Equals, int64(100))

	// win 2 lose 3 draw 1
	playerData := player.SerializedDataWithFields([]string{"achievements"})
	c.Assert(len(utils.GetMapSliceAtPath(playerData, "achievements")), Equals, 4) // 2type of currency

	for _, achievementData := range utils.GetMapSliceAtPath(playerData, "achievements") {
		if utils.GetStringAtPath(achievementData, "game_code") == "ceme" && utils.GetStringAtPath(achievementData, "currency_type") == currencyType {
			c.Assert(utils.GetStringAtPath(achievementData, "game_code"), Equals, "ceme")
			c.Assert(utils.GetIntAtPath(achievementData, "win_count"), Equals, 2)
			c.Assert(utils.GetIntAtPath(achievementData, "lose_count"), Equals, 3)
			c.Assert(utils.GetIntAtPath(achievementData, "draw_count"), Equals, 1)
		} else if utils.GetStringAtPath(achievementData, "game_code") == "tienlen" && utils.GetStringAtPath(achievementData, "currency_type") == currencyType {
			c.Assert(utils.GetStringAtPath(achievementData, "game_code"), Equals, "tienlen")
			c.Assert(utils.GetIntAtPath(achievementData, "win_count"), Equals, 0)
			c.Assert(utils.GetIntAtPath(achievementData, "lose_count"), Equals, 0)
			c.Assert(utils.GetIntAtPath(achievementData, "draw_count"), Equals, 0)
		}
	}
}

func (s *TestSuite) TestSerializedPlayer(c *C) {
	player1, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player1, NotNil)

	c.Assert(player1.username, Equals, utils.GetStringAtPath(player1.SerializedData(), "username"))
}

func (s *TestSuite) TestMoneyCannotBeNegative(c *C) {
	currencyType := currency.TestMoney
	player1, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	c.Assert(player1, NotNil)
	_, err = player1.IncreaseMoney(1000-player1.GetMoney(currencyType), currencyType, true)
	c.Assert(err, IsNil)
	money, err := player1.DecreaseMoney(2000, currencyType, true)
	c.Assert(err, IsNil)
	c.Assert(money, Equals, int64(0))

	money, err = player1.IncreaseMoney(2000, currencyType, true)
	c.Assert(err, IsNil)
	c.Assert(money, Equals, int64(2000))
	money, err = player1.IncreaseMoney(-5000, currencyType, true)
	c.Assert(err, IsNil)
	c.Assert(money, Equals, int64(0))
}

func (s *TestSuite) TestSendAddFriendRequest(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()
	player3, _ := generateNewPlayerWithRandomName()

	_, err := player1.SendFriendRequest(player2.Id())
	c.Assert(err, IsNil)
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 1)
	c.Assert(len(player2.relationshipManager.relationships), Equals, 0)
	_, err = player1.SendFriendRequest(player2.Id())
	c.Assert(err.Error(), Equals, "err:already_send_friend_request")
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 1)

	err = player2.DeclineFriendRequest(player3.Id())
	c.Assert(err.Error(), Equals, "err:friend_request_not_found")
	err = player2.DeclineFriendRequest(player1.Id())
	c.Assert(err, IsNil)
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player2.relationshipManager.relationships), Equals, 0)

	player1.SendFriendRequest(player2.Id())
	err = player2.AcceptFriendRequest(player3.Id())
	c.Assert(err.Error(), Equals, "err:friend_request_not_found")
	err = player2.AcceptFriendRequest(player1.Id())
	c.Assert(err, IsNil)
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player2.relationshipManager.relationships), Equals, 1)
	c.Assert(len(player1.relationshipManager.relationships), Equals, 1)
	c.Assert(len(player3.relationshipManager.relationships), Equals, 0)

	player3.SendFriendRequest(player1.Id())
	player1.AcceptFriendRequest(player3.Id())
	c.Assert(len(player1.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player3.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player2.relationshipManager.relationships), Equals, 1)
	c.Assert(len(player1.relationshipManager.relationships), Equals, 2)
	c.Assert(len(player3.relationshipManager.relationships), Equals, 1)

	c.Assert(player3.GetNumberOfFriends(), Equals, 1)
	c.Assert(player1.GetNumberOfFriends(), Equals, 2)
}

func (s *TestSuite) TestSendFriendRequestToPlayerWhoSendYouFriendRequest(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()

	_, err := player1.SendFriendRequest(player2.Id())
	c.Assert(err, IsNil)
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 1)
	becomeFriend, err := player2.SendFriendRequest(player1.Id())
	c.Assert(becomeFriend, Equals, true)
	c.Assert(len(player2.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player1.relationshipManager.friendRequests), Equals, 0)
	c.Assert(len(player2.relationshipManager.relationships), Equals, 1)
	c.Assert(len(player1.relationshipManager.relationships), Equals, 1)

}

// func (s *TestSuite) TestGetFriendList(c *C) {
// 	currencyType := currency.TestMoney

// 	player1, _ := generateNewPlayerWithRandomName()
// 	player2, _ := generateNewPlayerWithRandomName()
// 	player3, _ := generateNewPlayerWithRandomName()
// 	player4, _ := generateNewPlayerWithRandomName()
// 	player5, _ := generateNewPlayerWithRandomName()
// 	player6, _ := generateNewPlayerWithRandomName()
// 	player7, _ := generateNewPlayerWithRandomName()
// 	player8, _ := generateNewPlayerWithRandomName()
// 	player9, _ := generateNewPlayerWithRandomName()

// 	for _, player := range []*Player{player2, player3, player4, player5, player6, player7, player8, player9} {
// 		player1.SendFriendRequest(player.Id())
// 		player.AcceptFriendRequest(player1.Id())
// 	}

// 	c.Assert(len(player1.relationshipManager.relationships), Equals, 8)
// 	c.Assert(len(player1.GetFriendListData()), Equals, 8)

// 	for _, data := range player1.GetFriendListData() {
// 		c.Assert(utils.GetBoolAtPath(data, "to_player/is_online"), Equals, false)
// 		c.Assert(utils.GetStringAtPath(data, "to_player/username") != player1.username, Equals, true)
// 		c.Assert(utils.GetInt64AtPath(data, "to_player/current_activity/room/requirement"), Equals, int64(0))
// 	}

// 	player5.room, _ = game.NewRoom(games["ceme"], player5, 4, 100, currencyType, "aaa")
// 	player6.SetIsOnline(true)
// 	for _, data := range player1.GetFriendListData() {
// 		if utils.GetInt64AtPath(data, "to_player/id") == player6.Id() {
// 			c.Assert(utils.GetBoolAtPath(data, "to_player/is_online"), Equals, true)
// 		} else {
// 			c.Assert(utils.GetBoolAtPath(data, "to_player/is_online"), Equals, false)
// 		}
// 		if utils.GetInt64AtPath(data, "to_player/id") == player5.Id() {
// 			currentActivityData := utils.GetMapAtPath(data, "to_player/current_activity")
// 			c.Assert(utils.GetInt64AtPath(currentActivityData, "room/requirement"), Equals, int64(100))
// 		} else {
// 			c.Assert(utils.GetInt64AtPath(data, "to_player/current_activity/room/requirement"), Equals, int64(0))
// 		}
// 	}

// 	err := player1.Unfriend(player2.Id())
// 	c.Assert(err, IsNil)
// 	c.Assert(len(player1.relationshipManager.relationships), Equals, 7)
// 	c.Assert(len(player2.relationshipManager.relationships), Equals, 0)

// 	err = player1.Unfriend(player2.Id())
// 	c.Assert(err.Error(), Equals, "err:not_friend_yet")
// }

func (s *TestSuite) TestGetNotificationList(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()
	player3, _ := generateNewPlayerWithRandomName()
	player4, _ := generateNewPlayerWithRandomName()
	player5, _ := generateNewPlayerWithRandomName()
	player6, _ := generateNewPlayerWithRandomName()
	player7, _ := generateNewPlayerWithRandomName()
	player8, _ := generateNewPlayerWithRandomName()
	player9, _ := generateNewPlayerWithRandomName()

	player1.ClaimGift(player1.giftManager.getGifts()[0].Id())

	// for _, player := range []*Player{player2, player3, player4, player5, player6, player7, player8, player9} {
	// 	player1.SendFriendRequest(player.Id())
	// 	c.Assert(len(player.notificationManager.getNotificationListData()), Equals, 1)
	// }

	s.server.cleanupAllResponse()
	for _, player := range []*Player{player2, player3, player4, player5, player6, player7, player8, player9} {
		player.SendFriendRequest(player1.Id())
	}
	c.Assert(len(player1.notificationManager.getNotificationListData()), Equals, 8)
	c.Assert(player1.notificationManager.getTotalNumberOfNotifications(), Equals, 8)
	utils.DelayInDuration(1 * time.Second)
	for i := 0; i < 8; i++ {
		response := s.server.getAndRemoveResponse(player1.Id())
		c.Assert(response, NotNil)
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_receive_notification")
		c.Assert(utils.GetStringAtPath(response, "data/notification/notification_type"), Equals, "friend_request")
		c.Assert(utils.GetInt64AtPath(response, "data/notification/data/from_player/id") != player1.Id(), Equals, true)
		c.Assert(utils.GetIntAtPath(response, "data/number_of_notifications/total"), Equals, i+1)
		c.Assert(utils.GetIntAtPath(response, "data/number_of_notifications/friend_request"), Equals, i+1)
	}

	player1.AcceptFriendRequest(player2.Id())
	c.Assert(len(player1.notificationManager.getNotificationListData()), Equals, 7)
	utils.DelayInDuration(1 * time.Second)
	for _, player := range []*Player{player1, player2} {
		response := s.server.getAndRemoveResponse(player.Id())
		c.Assert(response, NotNil)
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_number_of_friends_increase")
		c.Assert(utils.GetIntAtPath(response, "data/number_of_friends"), Equals, 1)
	}
}

func (s *TestSuite) TestGetTwoTypeOfNotificationList(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()
	player3, _ := generateNewPlayerWithRandomName()
	player4, _ := generateNewPlayerWithRandomName()
	player5, _ := generateNewPlayerWithRandomName()
	player6, _ := generateNewPlayerWithRandomName()
	player7, _ := generateNewPlayerWithRandomName()
	player8, _ := generateNewPlayerWithRandomName()
	player9, _ := generateNewPlayerWithRandomName()

	// for _, player := range []*Player{player2, player3, player4, player5, player6, player7, player8, player9} {
	// 	player1.SendFriendRequest(player.Id())
	// 	c.Assert(len(player.notificationManager.getNotificationListData()), Equals, 1)
	// }

	s.server.cleanupAllResponse()
	for _, player := range []*Player{player2, player3, player4, player5, player6, player7, player8, player9} {
		player.SendFriendRequest(player1.Id())
	}
	c.Assert(len(player1.notificationManager.getFriendRequestNotificationListData()), Equals, 8)
	utils.DelayInDuration(1 * time.Second)
	for i := 0; i < 8; i++ {
		response := s.server.getAndRemoveResponse(player1.Id())
		c.Assert(response, NotNil)
		c.Assert(utils.GetStringAtPath(response, "method"), Equals, "player_receive_notification")
		c.Assert(utils.GetStringAtPath(response, "data/notification/notification_type"), Equals, "friend_request")
		c.Assert(utils.GetInt64AtPath(response, "data/notification/data/from_player/id") != player1.Id(), Equals, true)
	}

}

func (s *TestSuite) TestGetRelationshipData(c *C) {

	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()

	_, err := player1.SendFriendRequest(player2.Id())
	c.Assert(err, IsNil)
	relationshipData := player1.GetRelationshipDataWithPlayer(player2.Id())
	c.Assert(utils.GetBoolAtPath(relationshipData, "request_from_current_player"), Equals, true)
	c.Assert(utils.GetInt64AtPath(relationshipData, "friend_request/from_player/id"), Equals, player1.Id())
	relationshipData = player2.GetRelationshipDataWithPlayer(player1.Id())
	c.Assert(utils.GetBoolAtPath(relationshipData, "request_from_current_player"), Equals, false)
	c.Assert(utils.GetInt64AtPath(relationshipData, "friend_request/from_player/id"), Equals, player1.Id())

	player2.AcceptFriendRequest(player1.Id())
	relationshipData = player1.GetRelationshipDataWithPlayer(player2.Id())
	c.Assert(utils.GetStringAtPath(relationshipData, "relationship/relationship_type"), Equals, "friend")
	relationshipData = player2.GetRelationshipDataWithPlayer(player1.Id())
	c.Assert(utils.GetStringAtPath(relationshipData, "relationship/relationship_type"), Equals, "friend")

}

func (s *TestSuite) TestSearchPlayer(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	player1.UpdateUsername("vodich")
	player2, _ := generateNewPlayerWithRandomName()
	player2.UpdateUsername("vodich123")
	player3, _ := generateNewPlayerWithRandomName()
	player3.UpdateUsername("22vodich")
	player4, _ := generateNewPlayerWithRandomName()
	player4.UpdateUsername("vvvvvodich334d")
	player5, _ := generateNewPlayerWithRandomName()
	player5.UpdateUsername("voDich")
	player6, _ := generateNewPlayerWithRandomName()
	player6.UpdateUsername("vo dich222")
	player7, _ := generateNewPlayerWithRandomName()
	player7.UpdateUsername("vodfdahskhich")

	data, err := SearchPlayer("vodich") // can't really generate a name that contain this right......
	c.Assert(err, IsNil)
	c.Assert(len(utils.GetMapSliceAtPath(data, "results")), Equals, 5)
}

func (s *TestSuite) TestGetPrizeList(c *C) {
	prizes := getBiggestWinPrizeList(games["ceme"])
	c.Assert(len(prizes), Equals, 7)
	var lastFromRank int64
	for _, prize := range prizes {
		if lastFromRank == 0 {
			lastFromRank = prize.fromRank
		} else {
			c.Assert(lastFromRank < prize.fromRank, Equals, true)
		}
	}

	rangesToFetch := getFetchRangeFromPrizeList(prizes)
	c.Assert(len(rangesToFetch), Equals, 2) // 1 to 20 and 777 to 777
	c.Assert(rangesToFetch[0][0], Equals, int64(1))
	c.Assert(rangesToFetch[0][1], Equals, int64(20))
	c.Assert(rangesToFetch[1][0], Equals, int64(777))
	c.Assert(rangesToFetch[1][1], Equals, int64(777))
}

func (s *TestSuite) TestLeaderboard(c *C) {
	currencyType := currency.TestMoney
	// create player
	players := make([]*Player, 0)
	for i := 0; i < 100; i++ {
		player, _ := generateNewPlayerWithRandomName()
		player.RecordGameResult("tienlen", "lose", 10, currencyType)
		players = append(players, player)
	}

	player1 := players[0] // rank 1
	player1.RecordGameResult("tienlen", "lose", 500000, currencyType)

	player2 := players[1] // rank 2
	player2.RecordGameResult("tienlen", "lose", 60000, currencyType)
	player3 := players[2] // rank 2
	player3.RecordGameResult("tienlen", "lose", 60000, currencyType)

	for i := 3; i <= 20; i++ {
		player := players[i]
		player.RecordGameResult("tienlen", "win", 1000*int64(i), currencyType) // rank 3 and up to rank 23. win from 3000 to 20000
	}

	for i := 21; i < 40; i++ {
		player := players[i]
		for j := 0; j < 300+(40-i); j++ {
			player.RecordGameResult("tienlen", "win", 100, currencyType) // total gain will all be 30000 + (40-i)*100
		}
	}

	for _, player := range players {
		// claim first time bonus gift
		player.ClaimGift(player.giftManager.getGifts()[0].Id())
	}

	s.server.cleanupAllResponse()
	// get leaderboard
	playersData, _, _ := fetchPlayersInLeaderboard(50, 0, "biggest_win_this_week", currencyType, "tienlen")
	// fmt.Println("playerData", playersData)
	data, _ := json.Marshal(playersData)
	fmt.Println("string", string(data))
	c.Assert(len(playersData), Equals, 50)
	rank1Counter := 0
	rank2Counter := 0
	rank3Counter := 0
	for _, playerData := range playersData {
		rank := utils.GetInt64AtPath(playerData, "rank")
		if rank == 1 {
			rank1Counter++
			c.Assert(utils.GetInt64AtPath(playerData, "value"), Equals, int64(500000))
		} else if rank == 2 {
			rank2Counter++
			c.Assert(utils.GetInt64AtPath(playerData, "value"), Equals, int64(60000))
		} else if rank == 3 {
			rank3Counter++
			c.Assert(utils.GetInt64AtPath(playerData, "value"), Equals, int64(20000))
		}
	}
	c.Assert(rank1Counter, Equals, 1)
	c.Assert(rank2Counter, Equals, 2)
	c.Assert(rank3Counter, Equals, 1)

}

func (s *TestSuite) TestGetWeeklyReward(c *C) {
	xidachGame := games["ceme"]
	totalGainData := getTotalGainPrizeListData(xidachGame)
	biggestWinData := getBiggestWinPrizeListData(xidachGame)
	utils.Delay(1)
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (1,1,   10000000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (2,2,    5000000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (3,3,    2000000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (4,7,     500000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (8,12,    100000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (13,20,    50000,'xidach');
	// INSERT INTO biggest_win_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (777,777,1000000,'xidach');

	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (1,1,  10000000,'xidach');
	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (2,2,   5000000,'xidach');
	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (3,3,   2000000,'xidach');
	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (4,7,    500000,'xidach');
	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (8,12,   100000,'xidach');
	// INSERT INTO total_gain_weekly_prize (from_rank,to_rank,prize,game_code) VALUES (13,20,   50000,'xidach');

	/*
		player1 will have 2 prizes, both rank 1
		player2 and player3 will have 2 prizes, both rank 2

		player4 -> player21 (total 18 players) will have 1 prize, rank from 3 to 20, (biggest win)
		player22 -> player39 (total 18 players) will have 1 prize, rank from 3 to 20, (total gain)

	*/

	c.Assert(len(totalGainData), Equals, 6)
	c.Assert(len(biggestWinData), Equals, 7)
	c.Assert(utils.GetIntAtPath(totalGainData[0], "from_rank"), Equals, 1)
	c.Assert(utils.GetIntAtPath(totalGainData[1], "from_rank"), Equals, 2)
	c.Assert(utils.GetIntAtPath(totalGainData[2], "from_rank"), Equals, 3)
	c.Assert(utils.GetIntAtPath(totalGainData[3], "from_rank"), Equals, 4)
	c.Assert(utils.GetIntAtPath(totalGainData[4], "from_rank"), Equals, 8)

	c.Assert(utils.GetInt64AtPath(biggestWinData[0], "prize"), Equals, int64(10000000))
	c.Assert(utils.GetInt64AtPath(biggestWinData[1], "prize"), Equals, int64(5000000))

}

func (s *TestSuite) TestTimeBonus(c *C) {
	currencyType := currency.TestMoney
	player1, _ := generateNewPlayerWithRandomName()
	player2, _ := generateNewPlayerWithRandomName()

	player1OldMoney := player1.GetMoney(currencyType)

	_, err := player1.claimTimeBonus()
	c.Assert(err.Error(), Equals, "err:not_time_yet")
	_, err = player2.claimTimeBonus()
	c.Assert(err.Error(), Equals, "err:not_time_yet")

	timeBonusData = []map[string]interface{}{
		map[string]interface{}{
			"duration": "3h", // samples: 100y143d29h3m40s
			"bonus":    1000,
		},
		map[string]interface{}{
			"duration": "4h",
			"bonus":    3000,
		},
		map[string]interface{}{
			"duration": "5h",
			"bonus":    5000,
		},
		map[string]interface{}{
			"duration": "6h",
			"bonus":    7000,
		},
		map[string]interface{}{
			"duration": "7h",
			"bonus":    20000,
		},
	}

	// edit the last receive time
	queryString := fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-3*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err := player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 0)
	c.Assert(player1OldMoney+1000, Equals, player1.GetMoney(currencyType)) // lvl 1 bonus
	_, err = player2.claimTimeBonus()
	c.Assert(err.Error(), Equals, "err:not_time_yet")

	data = player1.getTimeBonusData()
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 0)

	queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-4*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err = player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(player1OldMoney+1000+3000, Equals, player1.GetMoney(currencyType)) // lvl 2 bonus
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 1)

	queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-5*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err = player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(player1OldMoney+1000+3000+5000, Equals, player1.GetMoney(currencyType)) // lvl 3 bonus
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 2)

	queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-6*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err = player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(player1OldMoney+1000+3000+5000+7000, Equals, player1.GetMoney(currencyType)) // lvl 4 bonus
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 3)

	queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-7*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err = player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(player1OldMoney+1000+3000+5000+7000+20000, Equals, player1.GetMoney(currencyType)) // lvl 5 bonus
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 4)

	queryString = fmt.Sprintf("UPDATE %s SET last_received_bonus = $1 WHERE player_id = $2", TimeBonusRecordDatabaseTableName)
	_, err = s.dataCenter.Db().Exec(queryString, time.Now().UTC().Add(-3*time.Hour), player1.Id())
	c.Assert(err, IsNil)

	// player1 should receive the first bonus now
	data, err = player1.claimTimeBonus()
	c.Assert(err, IsNil)
	c.Assert(player1OldMoney+1000+3000+5000+7000+20000+1000, Equals, player1.GetMoney(currencyType)) // lvl 1 bonus
	c.Assert(utils.GetIntAtPath(data, "last_bonus_index"), Equals, 0)

}

func (s *TestSuite) TestVip(c *C) {
	player1, _ := generateNewPlayerWithRandomName()

	var data map[string]interface{}
	data = player1.SerializedData()
	c.Assert(data["vip_code"], Equals, "vip_1")
	c.Assert(data["vip_score"], Equals, int64(0))

	_, _, err := player1.increaseVipScore(100)
	c.Assert(err, IsNil)
	data = player1.SerializedData()
	c.Assert(data["vip_code"], Equals, "vip_1")
	c.Assert(data["vip_score"], Equals, int64(100))

	_, _, err = player1.increaseVipScore(150)
	c.Assert(err, IsNil)
	data = player1.SerializedData()
	c.Assert(data["vip_code"], Equals, "vip_2")
	c.Assert(data["vip_score"], Equals, int64(250))

	_, _, err = player1.increaseVipScore(4000)
	c.Assert(err, IsNil)
	data = player1.SerializedData()
	c.Assert(data["vip_code"], Equals, "vip_3")
	c.Assert(data["vip_score"], Equals, int64(4250))
}

func (s *TestSuite) TestSendFeedback(c *C) {
	player1, _ := generateNewPlayerWithRandomName()
	err := player1.SendFeedback("1.0", 10, "bla bla bla")
	c.Assert(err, IsNil)
	err = player1.SendFeedback("1.0", 10, "hyhyhy")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:already_sent_feedback")
	err = player1.SendFeedback("1.1", 10, "tatata")
	c.Assert(err, IsNil)

	player1.fetchData()
	err = player1.SendFeedback("1.1", 10, "tatatatata")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "err:already_sent_feedback")

}

func (s *TestSuite) TestResetPassword(c *C) {
	player, password, _ := generateNewPlayerWithRandomNameAndPassword()

	_, err := generateResetPasswordCode("blablabla")
	c.Assert(err.Error(), Equals, "err:email_not_found")

	player.UpdateEmail("q@w.e")
	code, err := generateResetPasswordCode("q@w.e")
	c.Assert(err, IsNil)

	query := fmt.Sprintf("SELECT password FROM %s WHERE id = $1", PlayerDatabaseTableName)
	row := s.dataCenter.Db().QueryRow(query, player.Id())
	var passwordFromDB string
	err = row.Scan(&passwordFromDB)
	c.Assert(err, IsNil)
	c.Assert(utils.CompareHashedPassword(password, passwordFromDB), Equals, true)

	c.Assert(IsEmailAndResetPasswordCodeValid("qq@w.e", code), Equals, false)
	c.Assert(IsEmailAndResetPasswordCodeValid("q@w.e", "dfsakj"), Equals, false)
	c.Assert(IsEmailAndResetPasswordCodeValid("q@w.e", code), Equals, true)

	err = ResetPassword("qq@w.e", code, "abcdef")
	fmt.Println("wh", err)
	c.Assert(err, NotNil)
	err = ResetPassword("q@w.e", "dsds", "dsafjaslkdf")
	c.Assert(err, NotNil)
	err = ResetPassword("q@w.e", code, "d")
	c.Assert(err, NotNil)
	err = ResetPassword("q@w.e", code, "qwerty")
	c.Assert(err, IsNil)

	query = fmt.Sprintf("SELECT password FROM %s WHERE id = $1", PlayerDatabaseTableName)
	row = s.dataCenter.Db().QueryRow(query, player.Id())
	err = row.Scan(&passwordFromDB)
	c.Assert(err, IsNil)
	c.Assert(utils.CompareHashedPassword(password, passwordFromDB), Equals, false)
	c.Assert(utils.CompareHashedPassword("qwerty", passwordFromDB), Equals, true)

}
func (s *TestSuite) TestRegisterPNDevice(c *C) {
	player, _, _ := generateNewPlayerWithRandomNameAndPassword()
	err := player.RegisterPNDevice("cde", "xyz")
	c.Assert(err, IsNil)
	c.Assert(player.apnsDeviceToken, Equals, "cde")
	c.Assert(player.gcmDeviceToken, Equals, "xyz")
}

/*
helper
*/

func generateNewPlayerWithRandomName() (player *Player, err error) {
	return GenerateNewPlayer(GenerateRandomValidPlayerUsername(), GenerateRandomPlayerIdentifier(), utils.RandSeq(10), "bighero")
}

func generateNewPlayerWithRandomNameAndDeviceIdentifier(deviceIdentifier string) (player *Player, err error) {
	return GenerateNewPlayer(GenerateRandomValidPlayerUsername(), GenerateRandomPlayerIdentifier(), deviceIdentifier, "bighero")
}

func generateNewPlayerWithRandomNameAndPassword() (player *Player, password string, err error) {
	password = GenerateRandomPlayerIdentifier()
	player, err = GenerateNewPlayer(GenerateRandomValidPlayerUsername(), password, utils.RandSeq(10), "bighero")
	return player, password, err
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

func (server *TestServer) LogoutPlayer(playerId int64) {

}
