package genjwt

type Interface interface {
	GenJwtToken(subject string, audience []string) (string, error)
}
