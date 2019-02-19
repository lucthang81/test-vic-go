package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dchest/captcha"
)

func init() {
	_ = ioutil.WriteFile
	_ = fmt.Sprintf
	_ = base64.StdEncoding.EncodeToString
}

type Captcha struct {
	CaptchaId string
	PngImage  []byte
}

// create a captcha, return id and its png representation,
// expire after call func VerifyCaptcha or 10 minutes
func CreateCaptcha() Captcha {
	captchaLength := 4
	width, height := 160, 80

	captchaId := captcha.NewLen(captchaLength)
	// fmt.Println("captchaId", captchaId)
	var buf bytes.Buffer
	writer := io.MultiWriter(&buf)
	captcha.WriteImage(writer, captchaId, width, height)
	pngBytes := buf.Bytes()
	// fmt.Print(captchaId, base64.StdEncoding.EncodeToString(pngBytes))
	// ioutil.WriteFile(fmt.Sprintf("./%v.png", "hihi"), pngBytes, 0777)
	return Captcha{
		CaptchaId: captchaId,
		PngImage:  pngBytes,
	}
}

// call this remove captchaId from the store, even if return false
func VerifyCaptcha(captchaId string, digits string) bool {
	return captcha.VerifyString(captchaId, digits)
}
