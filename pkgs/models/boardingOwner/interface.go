package boardingowner

import "context"

const BdOwnerColl string = "boardingOwners"

// boarding owner account status
const (
	Pending  string = "pending"
	Accepted string = "accepted"
	Rejected string = "rejected"
)

type Interface interface {
	Create(ctx context.Context, doc *BoardingOwner) (key string, err error)
	Exist(ctx context.Context, key string) (bool, error)
	ListAnyJsonValue(ctx context.Context, query string, bindVars map[string]interface{}) ([]*map[string]interface{}, error)
}

type BoardingOwner struct {
	Key           string   `json:"_key"`
	CreatedAt     string   `json:"createdAt"`
	Name          string   `json:"name"`
	MainContact   string   `json:"mainContact"`
	OtherContacts []string `json:"otherContacts"`
	Email         string   `json:"email"`
	Password      string   `json:"password"`
	ImageUrl      string   `json:"imageUrl"`
	Address       string   `json:"address"`
	Location      string   `json:"location"`
	Status        string   `json:"status"`
}
