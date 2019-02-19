package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	//	"fmt"
	"io/ioutil"
)

// output is base64 encoded,
// padding scheme is the PKCS#1 v1.5
func Encrypt(plaintext string, publicKey *rsa.PublicKey) (string, error) {
	input := []byte(plaintext)
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, input)
	if err != nil {
		return "", err
	} else {
		return base64.StdEncoding.EncodeToString(cipherBytes), nil
	}
}

// input is base64 encoded
// padding scheme is the PKCS#1 v1.5
func Decrypt(cipherBase64 string, privateKey *rsa.PrivateKey) (string, error) {
	input, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}
	bs, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, input)
	if err != nil {
		return "", err
	} else {
		return string(bs), nil
	}
}

//
func readPublicKeyBytes(bs []byte) (*rsa.PublicKey, error) {
	//
	block, _ := pem.Decode(bs)
	keyI, e := x509.ParsePKIXPublicKey(block.Bytes)
	if e != nil {
		return nil, e
	}
	key, isOk := keyI.(*rsa.PublicKey)
	if isOk != true {
		return nil, errors.New("ReadPublicKeyFile type assertion error")
	}
	return key, nil
}

//
func ReadPublicKeyFile(path string) (*rsa.PublicKey, error) {
	bs, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return readPublicKeyBytes(bs)
}

//
func ReadPublicKeyString(pem string) (*rsa.PublicKey, error) {
	bs := []byte(pem)
	return readPublicKeyBytes(bs)
}

// input: ASN.1 PKCS#1 DER encoded
func readPrivateKeyBytes(bs []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(bs)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// input: ASN.1 PKCS#1 DER encoded
func ReadPrivateKeyFile(path string) (*rsa.PrivateKey, error) {
	bs, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return readPrivateKeyBytes(bs)
}

// input: ASN.1 PKCS#1 DER encoded
func ReadPrivateKeyString(pem string) (*rsa.PrivateKey, error) {
	bs := []byte(pem)
	return readPrivateKeyBytes(bs)
}
