package eventreaction

import "context"

const EventReactionColl string = "eventReactedBy"

type Interface interface {
	Get(ctx context.Context, key string) (*EventReaction, error)
	Create(ctx context.Context, data *EventReaction) (string, error)
	UpdateState(ctx context.Context, edgeKey string, data map[string]bool) error
}

type EventReaction struct {
	From  string `json:"_from"`
	To    string `json:"_to"`
	Key   string `json:"_key"`
	Type  string `json:"type,omitempty"`
	Love  bool   `json:"love"`
	Going bool   `json:"going"`
}
