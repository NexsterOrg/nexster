package types

type Event struct {
	Link        string `json:"link" validate:"required"`
	ImgType     string `json:"imgType" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Date        string `json:"date" validate:"required"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	Mode        string `json:"mode" validate:"required"`
	EventLink   string `json:"eventLink"`
}
