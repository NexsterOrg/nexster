package mediaowner

import "context"

const MediaOwnerColl string = "mediaOwnerEdges" // Name of collection
const MediaOwnerKind string = "media_owner"

type Interface interface {
	Create(ctx context.Context, fromId, toId string) (string, error)
	ListStringValueJson(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]string, error)
	GetOwnerForMedia(ctx context.Context, mediaKey string) (string, error)
}

type MediaOwner struct {
	Key  string `json:"_key"`
	From string `json:"_from"`
	To   string `json:"_to"`
	Kind string `json:"kind"`
}
