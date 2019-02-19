package encryption

// base on /crypto/cipher/example_test.go

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	//"os"
)

// The key argument should be the AES key, either 16 or 32 bytes
// to select AES-128 or AES-256.
var MINAH_AES_KEY_32 []byte
var INITIALIZATION_VECTOR []byte

// CBC mode works on blocks so plaintexts may need to be padded to the
// next whole block.
var PAD_CHAR byte

func init() {
	fmt.Print("")
	_ = hex.ErrLength
	_ = rand.Int
	_ = io.EOF
	_ = errors.New("")
	//
	MINAH_AES_KEY_32 = []byte("daominahdaominahdaominahdaominah")
	INITIALIZATION_VECTOR = []byte("daominahdaominah")
	PAD_CHAR = '_'
}

// input plainText string, output base64
func EncryptAesCbc(plaintextS string) (string, error) {
	key := MINAH_AES_KEY_32
	plaintext := []byte(plaintextS)
	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.

	if len(plaintext)%aes.BlockSize != 0 {
		//return "", errors.New("plaintext is not a multiple of the block size")
		paddingLength := aes.BlockSize - len(plaintext)%aes.BlockSize
		padding := make([]byte, paddingLength)
		for i := 0; i < paddingLength; i++ {
			padding[i] = PAD_CHAR
		}
		plaintext = append(plaintext, padding...)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := INITIALIZATION_VECTOR
	copy(ciphertext[:aes.BlockSize], iv)
	//
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// input base64, output plainText string
func DecryptAesCbc(ciphertextB64 string) (string, error) {
	key := MINAH_AES_KEY_32
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	//
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	//
	mode := cipher.NewCBCDecrypter(block, iv)
	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)
	// Remove padding
	i := len(ciphertext) - 1
	for {
		if i >= 0 {
			if ciphertext[i] == PAD_CHAR {
				i = i - 1
			} else {
				break
			}
		} else {
			break
		}
	}
	r := string(ciphertext[:i+1])
	// fmt.Println("i r", i, r)
	return r, nil
}
