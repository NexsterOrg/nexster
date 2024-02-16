package mail

import (
	"regexp"
	"strings"
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

func IsValidEmailDomain(email, validDomain string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	return parts[1] == validDomain
}

// Email format: <6 digits><alphabeticalCharacter>@uom.lk
func IsValidEmailV1(email string) bool {
	return regexp.MustCompile(`^\d{6}[a-zA-Z]@uom\.lk$`).MatchString(email)
}

// Email format: <name>.<batch>@uom.lk
func IsValidEmailV2(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+\.([0-9]{2})@uom\.lk$`).MatchString(email)
}
