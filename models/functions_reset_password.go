package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/vic/vic_go/models/player"
	"github.com/vic/vic_go/utils"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (models *Models) getResetPasswordPage(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	code := request.URL.Query().Get("code")
	email := request.URL.Query().Get("email")
	id, _ := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)

	if email == "" {
		_, err := models.GetPlayer(id)
		if err != nil {
			renderErrorForUser(renderer, err, fmt.Sprintf("/user/reset_password?code=%s&id=%s", code, id))
			return
		}
		data := make(map[string]interface{})
		data["code"] = code
		data["id"] = id

		data["for_user"] = true
		navLinks := make([]map[string]interface{}, 0)
		navLinks = appendCurrentNavLink(navLinks, "Reset password", "")
		data["nav_links"] = navLinks
		data["page_title"] = "Reset password"

		renderer.HTML(200, "user/reset_password", data)

	} else if player.IsEmailAndResetPasswordCodeValid(email, code) {

		data := make(map[string]interface{})
		data["code"] = code
		data["email"] = email

		data["for_user"] = true
		navLinks := make([]map[string]interface{}, 0)
		navLinks = appendCurrentNavLink(navLinks, "Reset password", "")
		data["nav_links"] = navLinks
		data["page_title"] = "Reset password"

		renderer.HTML(200, "user/reset_password", data)
	} else {
		renderErrorForUser(renderer, errors.New("Không tìm thấy email hoặc bạn đã thay đổi mật khẩu rồi"), fmt.Sprintf("/user/reset_password?code=%s&email=%s", code, email))
		return
	}

}

func (models *Models) sendResetPasswordEmail(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	email := utils.GetStringAtPath(data, "email")
	err = player.SendResetPasswordEmail(email)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	renderer.JSON(200, map[string]interface{}{})
}

func (models *Models) resetPassword(c martini.Context, request *http.Request, renderer render.Render, session sessions.Session) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)

	password := utils.GetStringAtPath(data, "password")
	phoneNumber := utils.NormalizePhoneNumber(utils.GetStringAtPath(data, "phone_number"))
	code := utils.GetStringAtPath(data, "code")

	playerInstance := player.FindPlayerWithPhoneNumber(phoneNumber)
	if playerInstance == nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Số điện thoại này chưa được xác nhận",
			"error_code": 0,
		})
		return
	}

	if len(code) == 0 {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Mã thiết lập mật khẩu mới không đúng",
			"error_code": 0,
		})
		return
	}

	if !player.IsIdAndResetPasswordCodeValid(playerInstance.Id(), code) {
		renderer.JSON(200, map[string]interface{}{
			"message":    "Mã thiết lập mật khẩu mới không đúng",
			"error_code": 0,
		})
		return
	}

	err = playerInstance.UpdatePassword(password)
	if err != nil {
		renderer.JSON(200, map[string]interface{}{
			"message":    err.Error(),
			"error_code": 0,
		})
		return
	}
	renderer.JSON(200, map[string]interface{}{})
}
