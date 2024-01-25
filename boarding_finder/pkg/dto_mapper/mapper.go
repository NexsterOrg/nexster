package dtomapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	vdtor "github.com/go-playground/validator/v10"

	bda "github.com/NamalSanjaya/nexster/pkgs/models/boarding_ads"
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

func ConvertAdWithOwnerData(data *bda.BdAdsWithOwner) *AdsWithOwner {
	return &AdsWithOwner{
		Ad: &BasicBdAd{
			Key:                 data.From.Key,
			Title:               data.From.Title,
			Description:         data.From.Description,
			Bills:               data.From.Bills,
			ImageUrls:           data.From.ImageUrls,
			Rent:                data.From.Rent,
			Address:             data.From.Address,
			Beds:                data.From.Beds,
			Baths:               data.From.Baths,
			Gender:              data.From.Gender,
			Distance:            data.From.Distance,
			DistanceUnit:        data.From.DistanceUnit,
			CreatedAt:           data.From.CreatedAt,
			LocationSameAsOwner: data.From.LocationSameAsOwner,
		},
		Owner: &BasicBdOwner{
			Key:           data.To.Key,
			CreatedAt:     data.To.CreatedAt,
			Name:          data.To.Name,
			MainContact:   data.To.MainContact,
			OtherContacts: data.To.OtherContacts,
			Address:       data.To.Address,
			Location:      data.To.Location,
		},
	}
}
