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
	GetAdWithOwner(ctx context.Context, adId string) (result *BdAdsWithOwner, err error)
	Update(ctx context.Context, key string, updateFields map[string]interface{}) error
	ListAdsWithFilters(ctx context.Context, minRent, maxRent, maxDist, minBeds, maxBeds, minBaths, maxBaths,
		offset, count int, sortBy string, genders, billTypes []string) ([]*AdInfoForList, error)
}

type BoardingAds struct {
	Key                 string   `json:"_key"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	Bills               string   `json:"bills"`
	ImageUrls           []string `json:"imageUrls"`
	Rent                int      `json:"rent"`
	Address             string   `json:"address"`
	Beds                int      `json:"beds"`
	Baths               int      `json:"baths"`
	Gender              string   `json:"gender"`
	Distance            float32  `json:"distance"` // TODO: Need to get this from google map.
	DistanceUnit        string   `json:"distanceUnit"`
	CreatedAt           string   `json:"createdAt"`
	Status              string   `json:"status"`
	LocationSameAsOwner bool     `json:"locationSameAsOwner"`
}

type BdAdsWithOwner struct {
	From BoardingAds `json:"from"`
	To   struct {
		Key           string   `json:"_key"`
		CreatedAt     string   `json:"createdAt"`
		Name          string   `json:"name"`
		MainContact   string   `json:"mainContact"`
		OtherContacts []string `json:"otherContacts"`
		Address       string   `json:"address"`
		Location      string   `json:"location"`
		Status        string   `json:"status"`
	} `json:"to"`
}

// List ads
type AdInfoForList struct {
	Key       string   `json:"key"`
	Title     string   `json:"title"`
	ImageUrls []string `json:"imageUrls"`
	Rent      int      `json:"rent"`
	Beds      int      `json:"beds"`
	Baths     int      `json:"baths"`
	Gender    string   `json:"gender"`
	Distance  float32  `json:"distance"`
	CreatedAt string   `json:"createdAt"`
}
