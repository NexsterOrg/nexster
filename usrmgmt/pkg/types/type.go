package types

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"reflect"

	ustr "github.com/NamalSanjaya/nexster/pkgs/utill/string"
)

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
	IndexNo  string `json:"index" validate:"required"`
	Password string `json:"passwd" validate:"required"`
}

type AccountCreationLinkBody struct {
	IndexNo string `json:"index" validate:"required"`
	// Email   string `json:"email" validate:"required"`
}

type UsrmgmtTypes interface {
	Profile | PasswordResetInfo | AccessTokenBody | AccountCreationLinkBody
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
