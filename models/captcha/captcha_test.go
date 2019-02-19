package captcha

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-martini/martini"
)

func init() {
	fmt.Print("")
	_ = errors.New("")
	_ = time.Now()
	_ = rand.Intn(10)
	_ = strings.Join([]string{}, "")
}

func registerHandles(r *martini.ClassicMartini) {
	r.Get("/verify", vHandle)
}

func vHandle(request *http.Request) string {
	captchaId := request.URL.Query().Get("captchaId")
	digits := request.URL.Query().Get("digits")
	vr := VerifyCaptcha(captchaId, digits)
	res := fmt.Sprintf("%v", vr)
	fmt.Println("response: ", res)
	return res
}

func TestHihi(t *testing.T) {
	CreateCaptcha()

	r := martini.Classic()
	registerHandles(r)
	//	r.RunOnAddr(":8080")
}
