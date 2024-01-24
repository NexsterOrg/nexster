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
