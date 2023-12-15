package student

import "context"

const StudnetColl string = "student"

type Interface interface {
	Create(ctx context.Context, doc *Student) (string, error)
}

type Student struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
	Kind string `json:"kind"`
}
