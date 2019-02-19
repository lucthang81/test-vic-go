package encryption

import (
	"fmt"
	//	"time"
	//	"encoding/base64"
	"testing"
)

var _, _ = fmt.Println("")

func TestDummy(t *testing.T) {
	fmt.Printf("key %s\n", MINAH_AES_KEY_32)
	s := "hohohaha1"
	e, err := EncryptAesCbc(s)
	if err != nil {
		fmt.Println("err DecryptAesCbc", err)
	}
	e = "ZGFvbWluYWhkYW9taW5haIfbY1gLLVQANgCI3VVhtPM="
	d, err := DecryptAesCbc(e)
	if err != nil {
		fmt.Println("err DecryptAesCbc", err)
	}
	fmt.Printf("string:     %v\n", s)
	fmt.Printf("cipherText: %v\n", e)
	fmt.Printf("plainText:  %v\n", d)
}
