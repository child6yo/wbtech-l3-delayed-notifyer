package sender

import (
	"fmt"
	"net/smtp"
)

// Email определяет отправщик электронных писем через SMTP сервер.
type Email struct {
	from string
	host string
	port string
}

// NewEmail создает новый Email.
func NewEmail(from, host, port string) *Email {
	return &Email{
		from: from,
		host: host,
		port: port,
	}
}

// Send отправляет сообщение на указанный адрес.
func (e *Email) Send(emailAddr string, data string) error {
	to := []string{emailAddr}
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Уведомление\r\n"+
		"\r\n"+
		"%s\r\n",
		e.from, to, data)

	addr := e.host + ":" + e.port
	return smtp.SendMail(addr, nil, e.from, to, []byte(msg))
}
