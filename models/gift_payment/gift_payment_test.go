package gift_payment

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
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

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	dataCenter      *datacenter.DataCenter
	dbName          string
	server          *TestServer
	adminId         int64
	playerIdCounter int64
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
	record.RegisterDataCenter(s.dataCenter)

	s.generateAdmin(c)

	s.server = NewTestServer()

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

/*
helper
*/

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

type TestPlayer struct {
	money    int64
	vipPoint int64
	id       int64
	name     string
	isOnline bool
	exp      int64

	recentResult   string
	recentChange   int64
	recentGameCode string
	playerType     string

	gameCount int
	bet       int64

	moneyMutex    sync.Mutex
	vipPointMutex sync.Mutex
}

func (s *TestSuite) newPlayer() *TestPlayer {
	return &TestPlayer{
		money:      100000,
		id:         s.genPlayerId(),
		name:       "ddd",
		playerType: "",
		isOnline:   true,
	}
}

func (player *TestPlayer) Id() int64 {
	return player.id
}
func (player *TestPlayer) LockVipPoint() {
	player.vipPointMutex.Lock()
}

func (player *TestPlayer) UnlockVipPoint() {
	player.vipPointMutex.Unlock()
}

func (player *TestPlayer) VipPoint() int64 {
	return player.vipPoint
}

func (player *TestPlayer) IncreaseVipPoint(vipPoint int64, action string, data map[string]interface{}, shouldLock bool) (newVipPoint int64, err error) {
	if shouldLock {
		player.LockVipPoint()
		defer player.UnlockVipPoint()
	}
	playerVipPoint := utils.MaxInt64(player.vipPoint+vipPoint, int64(0))
	newVipPoint, err = player.updateVipPoint(playerVipPoint)
	if err != nil {
		return
	}
	return
}

func (player *TestPlayer) DecreaseVipPoint(vipPoint int64, action string, data map[string]interface{}, shouldLock bool) (newVipPoint int64, err error) {
	if shouldLock {
		player.LockVipPoint()
		defer player.UnlockVipPoint()
	}
	playerVipPoint := utils.MaxInt64(player.vipPoint-vipPoint, int64(0))
	newVipPoint, err = player.updateVipPoint(playerVipPoint)
	if err != nil {
		return
	}
	return
}

func (player *TestPlayer) updateVipPoint(vipPoint int64) (newVipPoint int64, err error) {
	player.vipPoint = vipPoint
	return player.vipPoint, nil
}

// commu
func (player *TestPlayer) GetUnreadCountOfInboxMessages() (total int64, err error) {
	return 0, nil
}
func (player *TestPlayer) CreateRawMessage(title string, content string) (err error) {
	return nil
}
func (player *TestPlayer) AppType() string {
	return ""
}
func (player *TestPlayer) DeviceType() string {
	return ""
}
func (player *TestPlayer) APNSDeviceToken() string {
	return ""
}
func (player *TestPlayer) GCMDeviceToken() string {
	return ""
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

func (s *TestSuite) genPlayerId() int64 {
	s.playerIdCounter = s.playerIdCounter + 1
	return s.playerIdCounter
}
