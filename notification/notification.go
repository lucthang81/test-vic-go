package notification

import (
	"gopkg.in/gomail.v1"
)

var supportMailer = gomail.NewMailer("smtp.gmail.com", "cskh.bai69pro@gmail.com", "daihung123456", 587)

func SendSupportEmailWithHTML(to string, title string, htmlContent string) (err error) {
	msg := gomail.NewMessage()
	msg.SetHeader("From", "kengvipcskh@gmail.com")
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", title)
	msg.SetBody("text/html", htmlContent)
	if err = supportMailer.Send(msg); err != nil {
		return err
	}
	return nil
}

func SendSupportEmail(to string, title string, content string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", "kengvipcskh@gmail.com")
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", title)
	msg.SetBody("text/plain", content)
	if err := supportMailer.Send(msg); err != nil {
		return err
	}
	return nil
}
