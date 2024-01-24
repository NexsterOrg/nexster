package boardingads

import "context"

const BdAdsColl string = "boardingAds"

// boarding ad status
const (
	Pending  string = "pending"
	Accepted string = "accepted"
	Rejected string = "rejected"
)

type Interface interface {
	Create(ctx context.Context, doc *BoardingAds) (string, error)
}

type BoardingAds struct {
	Key          string   `json:"_key"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Bills        string   `json:"bills"`
	ImageUrls    []string `json:"imageUrls"`
	Rent         int      `json:"rent"`
	Address      string   `json:"address"`
	Beds         int      `json:"beds"`
	Baths        int      `json:"baths"`
	Gender       string   `json:"gender"`
	Distance     float32  `json:"distance"` // TODO: Need to get this from google map.
	DistanceUnit string   `json:"distanceUnit"`
	CreatedAt    string   `json:"createdAt"`
	Status       string   `json:"status"`
}
