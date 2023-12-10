package mail

type MailConfig struct {
	SmtpHostname string `yaml:"smtpHostname"` // smtp server
	Port         int    `yaml:"port"`
	HostMail     string `yaml:"hostMail"` // from
	Password     string `yaml:"password"`
}

type Interface interface {
	SendEmail(to, subject, htmlMailBody string) error
}
