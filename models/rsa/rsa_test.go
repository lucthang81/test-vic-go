package rsa

import (
	//	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	plain := "hohohaha"
	//
	pubKey, e1 := ReadPublicKeyFile(
		"/home/tungdt/go/src/github.com/vic/vic_go/models/rsa/key.pub")
	priKey, e2 := ReadPrivateKeyFile(
		"/home/tungdt/go/src/github.com/vic/vic_go/models/rsa/key.priv")
	if e1 != nil || e2 != nil {
		t.Error(e1, e2)
		return
	}
	//
	cipher, e3 := Encrypt(plain, pubKey)
	plain2, e4 := Decrypt(cipher, priKey)
	if e3 != nil || e4 != nil {
		t.Error(e3, e4)
	} else if plain != plain2 {
		t.Error()
	}

	//
	pubKey, e1 = ReadPublicKeyString(`-----BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDg8kjD9gbGoTjwoRVspQTNNYou
PykzVNBasJDe1z0V5jJri34bOG87AvF0qURXZzLqOmXkYHBVgK9k4utmIADxMtyW
Gwh3dmcOSy0dm7FnVR1BcHpZ6u2QaiNmMDNR19RLWjujNYGRsn/PKyuoIfSvW4or
HkXZ+h4xqT9/ejmbzQIDAQAB
-----END RSA PUBLIC KEY-----`)
	priKey, e2 = ReadPrivateKeyString(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDg8kjD9gbGoTjwoRVspQTNNYouPykzVNBasJDe1z0V5jJri34b
OG87AvF0qURXZzLqOmXkYHBVgK9k4utmIADxMtyWGwh3dmcOSy0dm7FnVR1BcHpZ
6u2QaiNmMDNR19RLWjujNYGRsn/PKyuoIfSvW4orHkXZ+h4xqT9/ejmbzQIDAQAB
AoGAMN3RWuiubiYF9Zg4zEJI+b9gxk0oSSNqo9jpj89YUNKSL3S9L3KiD0LDa2F+
HDKqB+Ip0mP0404yTAtTsfrP2S2gEHoyTBW7U7zDO+cgQ68WdDX2J7BEvYmA9aNn
VdbAYVJP3j39da/xiRbSLqNzRenD02CanHjBIm+SLhsSp20CQQD7c2i1/FmmfBzD
MYUSf0YvW9zLCc041SSnF3GFBvwWA25uX6E/GeF7Daqq/JtsX/7gn40Ltn2iVsM4
+A0CtuiPAkEA5QQe19gXaQoK7+pSYmOeE0lBLfux/9InSN9p2KqYbiUWDbJjkVHG
YVs3P7KqJG/t1OrrrBETmhVOGHKADmDL4wJBAPBshhdT5Uh5bWr5g1qPVUVdGX0N
rysDKZuWn9VpO0m1GDbyuxPBpEXraF87T0TNeL+/7rXfVLsPKHTlQFNzHmMCQQDF
ZfTT5V3gWxisTRQv3F+/je/Rm9aEg/b6mB/a8siqf+rvaWjrNEpDVmVb0TtYZuXg
FZGH4bw8nsqOxfrc6dAzAkARH4j3SmVSD6l27cUVTAzwOtfXEXCo4TqipJHbFTIg
HnJQqCqN5LLgRITyXD3GwONQGgmXheRoIMYMGALmWKZN
-----END RSA PRIVATE KEY-----
	`)
	if e1 != nil || e2 != nil {
		t.Error(e1, e2)
		return
	}
	//
	cipher, e3 = Encrypt(plain, pubKey)
	plain2, e4 = Decrypt(cipher, priKey)
	if e3 != nil || e4 != nil {
		t.Error(e3, e4)
	} else if plain != plain2 {
		t.Error()
	}
}
