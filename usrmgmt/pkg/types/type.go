package types

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
)

const Student string = "student"
const BoardingOwner string = "bdOwner"
const Member string = "member"

type Profile struct {
	About      string `json:"about"`
	FirstName  string `json:"firstName"`
	SecondName string `json:"secondName"`
	Faculty    string `json:"faculty"`
	Field      string `json:"field"`
	Batch      string `json:"batch"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday"`
}

type PasswordResetInfo struct {
	CurrentPassword string `json:"cp" validate:"required"`
	NewPassword     string `json:"np" validate:"required"`
}

type AccessTokenBody struct {
	Id       string `json:"id" validate:"required"`
	Password string `json:"passwd" validate:"required"`
	Consumer string `json:"consumer" validate:"required,oneof=student bdOwner member"` // token consumer types
}

type AccountCreationLinkBody struct {
	IndexNo string `json:"index" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

/*
TODO: Validations & Transformations
1. Convert last letter to simple.
*/

type LinkCreationParams struct {
	IndexNo   string `json:"index" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	ExpiredAt string `json:"exp" validate:"required"`
	Hmac      string `json:"hmac" validate:"required"`
}

type AccCreateBody struct {
	FirstName  string `json:"firstName" validate:"required"`
	SecondName string `json:"secondName" validate:"required"`
	IndexNo    string `json:"index" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	ImageId    string `json:"imageId" validate:"required"`
	Birthday   string `json:"birthday" validate:"required"`
	Faculty    string `json:"faculty" validate:"required"`
	Field      string `json:"field"`
	Batch      string `json:"batch" validate:"required"`
	About      string `json:"about"`
	Gender     string `json:"gender" validate:"required"`
	Password   string `json:"password" validate:"required"`
	ExpiredAt  string `json:"exp" validate:"required"`
	Hmac       string `json:"hmac" validate:"required"`
}

type ForgotPasswordResetLink struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResetBody struct {
	Email     string `json:"email" validate:"required,email"`
	ExpiredAt string `json:"exp" validate:"required"`
	Hmac      string `json:"hmac" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

type ValidatePasswordResetBody struct {
	Email     string `json:"email" validate:"required,email"`
	ExpiredAt string `json:"exp" validate:"required"`
	Hmac      string `json:"hmac" validate:"required"`
}

type UsrmgmtTypes interface {
	Profile | PasswordResetInfo | AccessTokenBody | AccountCreationLinkBody | LinkCreationParams | AccCreateBody |
		ForgotPasswordResetLink | ForgotPasswordResetBody | ValidatePasswordResetBody
}

// Generic function to read http req json body
func ReadJsonBody[T UsrmgmtTypes](r *http.Request) (*T, error) {
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

// This function does: remove empty fields, setting data properly(eg: username),
// set field to "" when necessary.
func RemoveEmptyFields[T UsrmgmtTypes](data *T) map[string]interface{} {
	firstNameTemp := ""
	secondNameTemp := ""
	isFac := false
	result := make(map[string]interface{})

	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		if v.Field(i).String() != "" {
			fieldName := ustr.FirstCharToLower(field.Name)
			result[fieldName] = value
			if fieldName == "faculty" {
				isFac = true
			}
			if fieldName == "firstName" {
				firstNameTemp = value.(string)
			} else if fieldName == "secondName" {
				secondNameTemp = value.(string)
			}
		}
	}

	isFirstNameEmpty := firstNameTemp == ""
	isSecondNameEmpty := secondNameTemp == ""

	if !isFirstNameEmpty && !isSecondNameEmpty {
		result["username"] = firstNameTemp + " " + secondNameTemp
	}
	if !isFirstNameEmpty && isSecondNameEmpty {
		// not allow firstName to update
		log.Println("Edit basic user info: remove firstName")
		delete(result, "firstName")
	}
	if isFirstNameEmpty && !isSecondNameEmpty {
		// not allow secondName to update
		log.Println("Edit basic user info: remove secondName")
		delete(result, "secondName")
	}

	if isFac {
		fac := result["faculty"].(string)
		if fac != "Engineering" {
			result["field"] = ""
		}
	}

	return result
}

/* Account Creation Body
validation & tranformation
1. make sure birthday format is in 2000-01-01
2. faculty & field validations
*/

func TransformToAccCreateData(data *AccCreateBody) *AccCreateBody {
	data.Gender = strings.ToLower(data.Gender)
	data.FirstName = ustr.FirstCharToUpper(data.FirstName)
	data.SecondName = ustr.FirstCharToUpper(data.SecondName)
	data.IndexNo = strings.ToLower(data.IndexNo)
	data.Birthday = ustr.ConvertBirthdayToCorrectFormat(data.Birthday)
	return data
}
