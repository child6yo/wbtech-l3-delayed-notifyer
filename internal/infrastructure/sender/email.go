package sender

import (
	"fmt"
	"net/smtp"
)

// EmailSender определяет отправщик электронных писем через SMTP сервер.
type EmailSender struct {
	from string
	host string
	port string
}

// NewEmailSenderсоздает новый EmailSender.
func NewEmailSender(from, host, port string) *EmailSender {
	return &EmailSender{
		from: from,
		host: host,
		port: port,
	}
}

// Send отправляет сообщение на указанный адрес.
func (es *EmailSender) Send(emailAddr string, data string) error {
	to := []string{emailAddr}
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Уведомление\r\n"+
		"\r\n"+
		"%s\r\n",
		es.from, to, data)

	addr := es.host + ":" + es.port
	return smtp.SendMail(addr, nil, es.from, to, []byte(msg))
}
