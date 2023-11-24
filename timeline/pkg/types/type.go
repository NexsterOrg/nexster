package types

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	vdtor "github.com/go-playground/validator/v10"
)

type Post struct {
	Link        string `json:"link" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Visibility  string `json:"visibility" validate:"required"`
}

type TimelineTypes interface {
	Post
}

// Generic function to read http req json body
func ReadJsonBody[T TimelineTypes](r *http.Request) (*T, error) {
	var data *T = new(T)
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}
	if err = vdtor.New().Struct(data); err != nil {
		return data, fmt.Errorf("required fields are missing: %v", err)
	}
	return data, nil
}
