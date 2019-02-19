package otp

import (
	"fmt"
	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/feature"
	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/game_config"
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
}

var _ = Suite(&TestSuite{
	dbName: "casino_otp_test",
})

func (s *TestSuite) SetUpSuite(c *C) {
	feature.UnlockAllFeature()
	rand.Seed(time.Now().UTC().UnixNano())
	test.CloneSchemaToTestDatabase(s.dbName, []string{"../../sql/init_schema.sql"})
	s.dataCenter = datacenter.NewDataCenter("vic_user", "9ate328di4rese7dra", s.dbName, ":63791")
	s.dataCenter.FlushCache()
	RegisterDataCenter(s.dataCenter)
	s.server = &TestServer{}
	player.RegisterServer(s.server)
	player.RegisterDataCenter(s.dataCenter)
	record.RegisterDataCenter(s.dataCenter)
	currency.RegisterDataCenter(s.dataCenter)
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

func (s *TestSuite) TestVersion(c *C) {

}

func (s *TestSuite) TestHashing(c *C) {
	accessKey := "jsqvd53030vband1z5go"
	command := "nv3"
	content := "Nv3 kingclub otp"
	phoneNumberFull := "841208958030"
	requestId := "8x98|834699|841208958030"
	requestTime := "2016-06-03T10:32:57Z"
	shortCode := "8098"
	signature := "9a517a30d0784820a4c8866754d90d51d4ea2ae000dae1f3c0cc409b228296bf"

	willBeHased := fmt.Sprintf("access_key=%s&command=%s&mo_message=%s&msisdn=%s&request_id=%s&request_time=%s&short_code=%s",
		accessKey, command, content, phoneNumberFull, requestId, requestTime, shortCode)

	hashedSignature := ComputeHmac256(willBeHased, kSecretKey)
	c.Assert(hashedSignature, Equals, signature)
}

func (s *TestSuite) TestOtpCode(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player2, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.SetPhoneNumber("0989543629")
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, false)

	otpCode := getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, NotNil) // did not sent sms yet

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)
	otpCode = getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode.status, Equals, kStatusValid)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)
	c.Assert(AlreadyReceiveVerifyReward("098975648", player.Id()), Equals, true)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player2.Id()), Equals, true)

	c.Assert(player.PhoneNumberChangeAvailable(), Equals, false)

	// register change phone number
	err = RegisterChangePhoneNumber(player.Id())
	c.Assert(err, IsNil)

	otpCode = getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")
	c.Assert(otpCode.reason, Equals, "change_phone_number")

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)
	otpCode = getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode.status, Equals, kStatusValid)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)
	c.Assert(player.PhoneNumberChangeAvailable(), Equals, true)

	player.UpdatePhoneNumber("0984753724")
	c.Assert(player.IsVerify(), Equals, false)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, true)
}

func (s *TestSuite) TestOtpCodeRequestAgain(c *C) {
	configData := game_config.GetDefaultData()
	configData["OtpCodeExpiredAfter"] = "10s"
	configData["OtpExpiredAfter"] = "10s"
	configData["OtpRequestAgainDuration"] = "10s"
	game_config.UpdateTestData(configData)

	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.UpdatePhoneNumber("098933434329")
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, false)
	otpCode1 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode1, NotNil)

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)

	otpCode2 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode2, NotNil)
	c.Assert(otpCode1.otpCode, Equals, otpCode2.otpCode)

	utils.DelayInDuration(10 * time.Second)
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)

	otpCode3 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode3, NotNil)
	c.Assert(otpCode3.otpCode != otpCode2.otpCode, Equals, true)

	otpCode3.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode3.status, Equals, kStatusAlreadySentSms)

	otpCode4 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode4, NotNil)
	c.Assert(otpCode3.otpCode, Equals, otpCode4.otpCode)

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)

	utils.DelayInDuration(10 * time.Second)
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)

	otpCode5 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode5, NotNil)
	c.Assert(otpCode5.otpCode != otpCode3.otpCode, Equals, true)

	_, err = VerifyOtpCode(player.Id(), otpCode1.otpCode)
	c.Assert(err, NotNil)
	_, err = VerifyOtpCode(player.Id(), otpCode3.otpCode)
	c.Assert(err, NotNil)
	_, err = VerifyOtpCode(player.Id(), otpCode5.otpCode)
	c.Assert(err, NotNil) // havent send sms

	otpCode5.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode5.status, Equals, kStatusAlreadySentSms)
	_, err = VerifyOtpCode(player.Id(), otpCode5.otpCode)
	c.Assert(err, IsNil)

	otpCode6 := getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode6.status, Equals, kStatusValid)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)
	c.Assert(player.PhoneNumberChangeAvailable(), Equals, false)
}

func (s *TestSuite) TestOtpCodeRequestAgainChangeExpiredAt(c *C) {
	configData := game_config.GetDefaultData()
	configData["OtpCodeExpiredAfter"] = "10s"
	configData["OtpExpiredAfter"] = "10s"
	configData["OtpRequestAgainDuration"] = "10s"
	game_config.UpdateTestData(configData)

	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.UpdatePhoneNumber("098933435329")
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)
	c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, false)
	otpCode1 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode1, NotNil)

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)

	otpCode2 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode2, NotNil)
	c.Assert(otpCode1.otpCode, Equals, otpCode2.otpCode)

	utils.DelayInDuration(10 * time.Second)

	// change otp expired at here
	query := "UPDATE otp_code set expired_at = $1" +
		" WHERE id = $2"
	_, err = dataCenter.Db().Exec(query, otpCode2.expiredAt.Add(10*time.Second).UTC(), otpCode2.id)

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil) // still cannot, code hasn't expired yet

	otpCode3 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode3, NotNil)
	c.Assert(otpCode2.otpCode, Equals, otpCode3.otpCode)

	utils.DelayInDuration(20 * time.Second)

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil) // expired, create new one

	otpCode4 := getOtpCodeForPhoneNumber(player.PhoneNumber())
	c.Assert(otpCode4, NotNil)
	c.Assert(otpCode4.otpCode != otpCode3.otpCode, Equals, true)
}

func (s *TestSuite) TestOtpCodeChangePhoneNumber(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.UpdatePhoneNumber("0989543629")
	player.SetIsVerify(true)
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)
	// c.Assert(AlreadyReceiveVerifyReward(player.PhoneNumber(), player.Id()), Equals, false)

	err = RegisterChangePhoneNumber(player.Id())
	c.Assert(err, IsNil)
	otpCode := getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)

	c.Assert(player.PhoneNumberChangeAvailable(), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)

	err = player.UpdatePhoneNumber("0986352735")
	c.Assert(err, IsNil)

	c.Assert(player.PhoneNumberChangeAvailable(), Equals, false)
	c.Assert(player.IsVerify(), Equals, false)
	c.Assert(player.PhoneNumber(), Equals, "84986352735")

	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, IsNil)
	otpCode = getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)
	c.Assert(player.IsVerify(), Equals, true)
	c.Assert(player.PhoneNumber(), Equals, "84986352735")
}

func (s *TestSuite) TestOtpCodeChangePassword(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.UpdatePhoneNumber("0989543649")
	player.SetIsVerify(true)
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, false)
	err = player.UpdatePassword("abcxyz")
	c.Assert(err, NotNil)

	err = RegisterChangePassword(player.Id())
	c.Assert(err, IsNil)
	otpCode := getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)

	err = player.UpdatePassword("abcxyz")
	c.Assert(err, IsNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, false)
	c.Assert(player.IsVerify(), Equals, true)
}

func (s *TestSuite) TestOtpCodeResetPassword(c *C) {
	player, err := generateNewPlayerWithRandomName()
	c.Assert(err, IsNil)
	player.UpdatePhoneNumber("84989343649")
	player.SetIsVerify(true)
	err = RegisterVerifyPhoneNumber(player.Id())
	c.Assert(err, NotNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, false)
	err = player.UpdatePassword("abcxyz")
	c.Assert(err, NotNil)

	c.Assert(player.PhoneNumber(), Equals, "84989343649")
	err = RegisterResetPassword(player.PhoneNumber())
	c.Assert(err, IsNil)
	otpCode := getLatestOtpCodeForPlayer(player.Id())
	c.Assert(otpCode, NotNil)
	c.Assert(otpCode.status, Equals, "wait")

	otpCode.updateCurrentStatus(kStatusAlreadySentSms)
	c.Assert(otpCode.status, Equals, kStatusAlreadySentSms)

	_, err = VerifyOtpCode(player.Id(), otpCode.otpCode)
	c.Assert(err, IsNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, true)
	c.Assert(player.IsVerify(), Equals, true)

	err = player.UpdatePassword("abcxyz")
	c.Assert(err, IsNil)

	c.Assert(player.PasswordChangeAvailable(), Equals, false)
	c.Assert(player.IsVerify(), Equals, true)
}

// func (s *TestSuite)

/*
helper
*/

func generateNewPlayerWithRandomName() (playerInstance *player.Player, err error) {
	return player.GenerateNewPlayer(player.GenerateRandomValidPlayerUsername(), player.GenerateRandomPlayerIdentifier(), utils.RandSeq(10), "bighero")
}

type TestServer struct {
}

func (server *TestServer) SendRequest(requestType string, data map[string]interface{}, toPlayerId int64) {

}
func (server *TestServer) SendRequests(requestType string, data map[string]interface{}, toPlayerIds []int64) {

}
func (server *TestServer) SendRequestsToAll(requestType string, data map[string]interface{}) {

}
func (server *TestServer) DisconnectPlayer(playerId int64, data map[string]interface{}) {

}
func (server *TestServer) NumberOfRequest() int64 {
	return 0
}
func (server *TestServer) AverageRequestHandleTime() float64 {
	return 0
}

func (server *TestServer) SendHotFixRequest(requestType string, data map[string]interface{}, currencyType string, toPlayerId int64) {

}

func (server *TestServer) LogoutPlayer(playerId int64) {
	return
}
