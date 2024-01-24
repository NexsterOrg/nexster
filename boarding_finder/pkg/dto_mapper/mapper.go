package dtomapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	vdtor "github.com/go-playground/validator/v10"
)

type DtoTypes interface {
	CreateAdDto | CreateBoardingOwner
}

// Generic function to read http req json body
func ReadJsonBody[T DtoTypes](r *http.Request, validator *vdtor.Validate) (*T, error) {
	var data *T = new(T)
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(b, &data); err != nil {
		return data, err
	}

	if err = validator.Struct(data); err != nil {
		return nil, fmt.Errorf("failed to validate: %v", err)
	}

	return data, nil
}
