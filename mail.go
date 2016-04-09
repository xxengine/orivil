package orivil

import (
	"net/smtp"
	"strings"
)

var cfgEmail = &struct {
	User     string
	Password string
	From     string
	Host     string
	Port     string
	Type     string
}{}

func init() {

	Cfg.ReadStruct("mail.yml", cfgEmail)
}

func SendEmail(to string, title, body string) error {

	auth := smtp.PlainAuth("", cfgEmail.User, cfgEmail.Password, cfgEmail.Host)

	addr := cfgEmail.Host + cfgEmail.Port
	_to := strings.Split(to, ";")
	msg := []byte("To: " + to + "\r\n" +
		"From: " + cfgEmail.From + "\r\n" +
		"Subject: " + title + "\r\n" +
		"Content-Type: " + cfgEmail.Type +
		"\r\n\r\n" + body + "\r\n")
	return smtp.SendMail(addr, auth, cfgEmail.User, _to, []byte(msg))
}
