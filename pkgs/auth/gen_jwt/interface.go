package genjwt

type Interface interface {
	GenJwtToken(subject string, roles, audience []string) (string, error)
}
