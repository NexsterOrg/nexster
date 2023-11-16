package types

import (
	"encoding/json"
	"io"
	"net/http"
)

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

type EventReaction struct {
	Love  bool `json:"love"`
	Going bool `json:"going"`
}

type EventTypes interface {
	Event | EventReaction
}

// Generic function to read http req json body
func ReadJsonBody[T EventTypes](r *http.Request) (*T, error) {
	var data *T = new(T)
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}
	return data, nil
}
