package dtomapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	vdtor "github.com/go-playground/validator/v10"

	bda "github.com/NamalSanjaya/nexster/pkgs/models/boarding_ads"
)

type dtoTypes interface {
	CreateAdDto | CreateBoardingOwner | AdStatus | Otp
}

// Generic function to read http req json body
func ReadJsonBody[T dtoTypes](r *http.Request, validator *vdtor.Validate) (*T, error) {
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

func ConvertQueryParams(r *http.Request) *ListFilterQueryParams {
	result := &ListFilterQueryParams{}
	queryParams := r.URL.Query()

	// Extract and convert individual parameters
	result.Pg = convertToInt(queryParams.Get("pg"), 1)
	result.PgSize = convertToInt(queryParams.Get("pgSize"), 10)

	result.MinRent = convertToInt(queryParams.Get("mnr"), 0)
	result.MaxRent = convertToInt(queryParams.Get("mxr"), 1e5)
	result.MaxDistance = convertToInt(queryParams.Get("mxd"), 1e4)
	result.MinBeds = convertToInt(queryParams.Get("mnb"), 0)
	result.MaxBeds = convertToInt(queryParams.Get("mxb"), 100)
	result.MinBaths = convertToInt(queryParams.Get("mnba"), 0)
	result.MaxBaths = convertToInt(queryParams.Get("mxba"), 100)
	result.Genders = convertToStrArr(queryParams["for"], []string{"boys", "girls", "any"}, 3)
	result.BillTypes = convertToStrArr(queryParams["b"], []string{"include", "exclude"}, 2)
	result.SortBy = queryParams.Get("sort")
	if result.SortBy != "date" && result.SortBy != "rental" {
		result.SortBy = "date"
	}
	return result
}

func convertToInt(valStr string, defaultVal int) int {
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return defaultVal
	}
	return val
}

func convertToStrArr(arr, defaultArr []string, mxLn int) []string {
	ln := len(arr)
	if ln == 0 || ln > mxLn {
		return defaultArr
	}
	return arr
}
