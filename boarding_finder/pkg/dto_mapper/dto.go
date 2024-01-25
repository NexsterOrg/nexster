package dtomapper

type CreateAdDto struct {
	Title        string   `json:"title" validate:"required,max=255"`
	Description  string   `json:"description" validate:"max=1000"`
	Bills        string   `json:"bills" validate:"required,oneof=include exclude"`
	ImageUrls    []string `json:"imageUrls" validate:"required,min=1,max=5"`
	Rent         int      `json:"rent" validate:"required,gt=0"`
	Address      string   `json:"address" validate:"required,max=255"`
	Beds         int      `json:"beds" validate:"required,gt=0"`
	Baths        int      `json:"baths" validate:"required,gt=0"`
	Gender       string   `json:"gender" validate:"required,oneof=boy girl any"`
	Distance     float32  `json:"distance" validate:"required"` // TODO: Need to get this from google map.
	DistanceUnit string   `json:"distanceUnit" validate:"required,oneof=m km"`
}

type CreateBoardingOwner struct {
	Name          string   `json:"name" validate:"required,max=50"`
	MainContact   string   `json:"mainContact" validate:"required,max=20"`
	OtherContacts []string `json:"otherContacts" validate:"max=4"`
	Email         string   `json:"email" validate:"max=60"`
	Password      string   `json:"password" validate:"required,min=8,max=30"`
	ImageUrl      string   `json:"imageUrl" validate:"max=512"`
	Address       string   `json:"address" validate:"required,max=200"`
	Location      string   `json:"location"`
}

// Ad with owner
type BasicBdAd struct {
	Key                 string   `json:"key"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	Bills               string   `json:"bills"`
	ImageUrls           []string `json:"imageUrls"`
	Rent                int      `json:"rent"`
	Address             string   `json:"address,omitempty"`
	Beds                int      `json:"beds"`
	Baths               int      `json:"baths"`
	Gender              string   `json:"gender"`
	Distance            float32  `json:"distance,omitempty"` // TODO: Need to get this from google map.
	DistanceUnit        string   `json:"distanceUnit,omitempty"`
	CreatedAt           string   `json:"createdAt"`
	LocationSameAsOwner bool     `json:"locationSameAsOwner"`
}

type BasicBdOwner struct {
	Key           string   `json:"key"`
	CreatedAt     string   `json:"createdAt"`
	Name          string   `json:"name"`
	MainContact   string   `json:"mainContact"`
	OtherContacts []string `json:"otherContacts"`
	Address       string   `json:"address,omitempty"`
	Location      string   `json:"location,omitempty"`
}

type AdsWithOwner struct {
	Ad    *BasicBdAd    `json:"ad"`
	Owner *BasicBdOwner `json:"owner"`
}
