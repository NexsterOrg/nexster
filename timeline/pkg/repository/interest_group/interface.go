package interestgroup

import (
	"context"
	"fmt"
)

const interestGroupBaseKey = "interestGroup"

var interestGroupPropBaseKey string = fmt.Sprintf("%s#prop", interestGroupBaseKey)

// properties
const (
	statusCreating string = "creating"
	// statusOk       string = "ok"
	// statusErr      string = "error"
)

type Interface interface {
	ListVideoIdsForGroup(ctx context.Context, groupId string) ([]string, error)
	IsCreating(ctx context.Context, groupId string) (bool, error)
}
