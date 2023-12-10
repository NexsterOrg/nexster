package mail

import (
	"time"

	"github.com/go-mail/mail"
)

type mailClient struct {
	dialer *mail.Dialer
	from   string
}

var _ Interface = (*mailClient)(nil)

func New(cfg *MailConfig) *mailClient {
	d := mail.NewDialer(cfg.SmtpHostname, cfg.Port, cfg.HostMail, cfg.Password)
	d.Timeout = 20 * time.Second // set 20 sec timeout
	return &mailClient{
		dialer: d,
		from:   cfg.HostMail,
	}
}

func (mc *mailClient) SendEmail(to, subject, htmlMailBody string) error {
	m := mail.NewMessage()
	m.SetHeader("From", mc.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlMailBody)

	return mc.dialer.DialAndSend(m)
}
