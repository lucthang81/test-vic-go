package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/vic/vic_go/datacenter"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/utils"
	"github.com/vic/vic_go/zconfig"
	"github.com/vic/vic_go/zglobal"
)

func init() {
	_ = fmt.Printf
}

func Test1(t *testing.T) {
	//	fmt.Println("hihi")
	_ = datacenter.NewDataCenter
	_ = player.GetPlayer
	_ = utils.PFormat
	_ = errors.New
	_ = json.Marshal
}

func Test2(t *testing.T) {
	//	m := map[int64]float64{
	//		2018:    0.80,
	//		12018:   0.08,
	//		22018:   0.05,
	//		52018:   0.04,
	//		102018:  0.02,
	//		202018:  0.004,
	//		302018:  0.003,
	//		502018:  0.002,
	//		1002018: 0.001,
	//	}
	//	c := map[int64]int{}
	//	for i := 0; i < 10000; i++ {
	//		r := RandomFromMapChange(m)
	//		c[r] += 1
	//	}
	//	fmt.Println("c", c)
}

func Test3(t *testing.T) {
	//	d := datacenter.NewDataCenter(
	//		"vic_user", "123qwe", ":5432", "casino_vic_db", ":6379")
	//	RegisterDataCenter(d)
	//	p, e := player.GetPlayer(2987)
	//	v := p.GetMoney("test_money1")
	//	fmt.Println(p, e, v)
}

func Test4(t *testing.T) {
	//	s := `{"pubkey":"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvBjH1MkNhimRkxHcq5rmLMop84rurjW32STQbNHtPycF4qQtDx3ql0bmyIb2XMfCs1k2Q8pOhnizdEG7TPLWNjnefELS9Kx2rULR+ulRgzkhJB1Co5Z8mJNy+lxmS+a50kCILsXoAALPqSCZ0X8FZg9I8y5LB6atKYgDxDtG3ZAgAgqE4L61k7Ju50D849Y0YAhMyivr6xxEMW8tAYrwBrTUwC+X9WqIsGhxnIvTt+z1hHmuIjxk6NKXEZKIDJIGPhvRNR8mvzmm1WL6HNtF6A3mr33AA0E8IDscVfj9SHjfIS37bumFuqvKqr810CBUef4rNVhlyU6JsrQC7Pa7PQIDAQAB","purchase_type":"iapAndroid","receipt":"{\"signature\":\"GrbMKt72gfClRed0Ga27R9KV/+H7MikXrfmzRMN6yZEiT9uEj0TnHMrecUFHBmvm1vlC3cAwYJki69GiQSqjS94DV7Jqb6brLMQORZVgRk4ovvUJh8lm47Xp4mtRypICOfmnWhMpp7YZF7gdUHHXgLszWa9ka4hkgafEwwXMrUEsNyScQ36GyZsq4OeTF1oYShKZoXWiZ3mP509+TAgJ4Idnzsy3O1s3C7p1dsHll0V6HcH5CPhH3lxR16VbqXNx6sobQVq9cxJjupakOMFOQ7O1E5AB6i0AT151CV4uFIO1ag9E7PuuhH7HfoPRodx7PFLDMsO7WhnjDZRI6TiuTg==\",\"jsonData\":\"{\\\"orderId\\\":\\\"GPA.3391-3596-2272-25807\\\",\\\"packageName\\\":\\\"air.choilon.thanglon\\\",\\\"productId\\\":\\\"goi1do\\\",\\\"purchaseTime\\\":1521441178044,\\\"purchaseState\\\":0,\\\"purchaseToken\\\":\\\"jelfpccpdlbgfdlklidkmhig.AO-J1OwXU3D0uD5VrKMIFj6l3hS83Fr7AP19iNs85lOipga4mQkChkntzNZcE60NMXkjAsUDCqxILjGfPdPGSGXQKpET3OmnE4Ex-gWvBIIWRdKukc2GWqY\\\"}\",\"developerPayload\":\"\",\"itemType\":\"inapp\",\"itemId\":\"goi1do\",\"purchaseToken\":\"jelfpccpdlbgfdlklidkmhig.AO-J1OwXU3D0uD5VrKMIFj6l3hS83Fr7AP19iNs85lOipga4mQkChkntzNZcE60NMXkjAsUDCqxILjGfPdPGSGXQKpET3OmnE4Ex-gWvBIIWRdKukc2GWqY\",\"orderId\":\"GPA.3391-3596-2272-25807\",\"purchaseTime\":1521441178044}","sig":"GrbMKt72gfClRed0Ga27R9KV/+H7MikXrfmzRMN6yZEiT9uEj0TnHMrecUFHBmvm1vlC3cAwYJki69GiQSqjS94DV7Jqb6brLMQORZVgRk4ovvUJh8lm47Xp4mtRypICOfmnWhMpp7YZF7gdUHHXgLszWa9ka4hkgafEwwXMrUEsNyScQ36GyZsq4OeTF1oYShKZoXWiZ3mP509+TAgJ4Idnzsy3O1s3C7p1dsHll0V6HcH5CPhH3lxR16VbqXNx6sobQVq9cxJjupakOMFOQ7O1E5AB6i0AT151CV4uFIO1ag9E7PuuhH7HfoPRodx7PFLDMsO7WhnjDZRI6TiuTg=="}`
	//	var d map[string]interface{}
	//	err := json.Unmarshal([]byte(s), &d)
	//	fmt.Println("d", utils.PFormat(d))
	//	fmt.Println("err", err)
	//	pubkey, e1 := d["pubkey"].(string)
	//	receipt, e2 := d["receipt"].(string)
	//	sig, e3 := d["sig"].(string)
	//	if !e1 || !e2 || !e3 {
	//		fmt.Println("e1, e2, e3", e1, e2, e3)
	//	}
	//	isValid, e := VerifySignature(pubkey, []byte(receipt), sig)
	//	fmt.Println("isValid, e", isValid, e)
	//	//
	//	var temp interface{}
	//	err = json.Unmarshal([]byte(receipt), &temp)
	//	if err != nil {
	//		fmt.Println("ERROR 1", err)
	//	}
	//	receiptObj, isOk := temp.(map[string]interface{})
	//	if !isOk {
	//		fmt.Println("receiptObj, !isOk")
	//	}
	//	receiptFieldJsonData, isOk := receiptObj["jsonData"].(string)
	//	var receiptFieldJsonDataObj map[string]interface{}
	//	json.Unmarshal([]byte(receiptFieldJsonData), &receiptFieldJsonDataObj)
	//	productId, isOk := receiptFieldJsonDataObj["productId"].(string)
	//	fmt.Println("productId", productId)
	//
	//	isValid, e = VerifySignature(pubkey, []byte(receiptFieldJsonData), sig)
	//	fmt.Println("isVal"github.com/vic/vic_go/zglobal"id, e2", isValid, e)
}

func Test5(t *testing.T) {
	dataCenterInstance := datacenter.NewDataCenter(
		zconfig.PostgresUsername, zconfig.PostgresPassword,
		zconfig.PostgresAddress, zconfig.PostgresDatabaseName,
		zconfig.RedisAddress)
	record.RegisterDataCenter(dataCenterInstance)
	//
	zglobal.SmsSender = "onewaysms"
	e := SendSms("+849018169694", "chao ban toi la Tung")
	if e != nil {
		t.Error(e)
	}
}
