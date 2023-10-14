package uuid

import (
	"github.com/google/uuid"
)

func GenUUID4() string {
	return uuid.New().String()
}
