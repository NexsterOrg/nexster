package faculty

import (
	"context"
	"fmt"

	argdb "github.com/NamalSanjaya/nexster/pkgs/arangodb"
)

type facultyCtrler struct {
	argClient *argdb.Client
}

func NewCtrler(argClient *argdb.Client) *facultyCtrler {
	return &facultyCtrler{argClient: argClient}
}

func (fc *facultyCtrler) MkFacultyDocId(key string) string {
	return fmt.Sprintf("%s/%s", FacultyColl, key)
}

func (fc *facultyCtrler) GetFaculty(ctx context.Context, key, gender string) (*FacultyWithGender, error) {
	var fac Faculty
	_, err := fc.argClient.Coll.ReadDocument(ctx, key, &fac)
	if err != nil {
		return nil, err
	}
	switch gender {
	case male:
		return &fac.Male, nil
	case female:
		return &fac.Female, nil
	}
	return nil, fmt.Errorf("failed to get info for gender=%s, key=%s", gender, key)
}

func (fc *facultyCtrler) GetPriority(faculty, preferGender string, self *FacultyWithGender) int {
	var priority FacultyPriority
	if preferGender == male {
		priority = self.Male
	} else if preferGender == female {
		priority = self.Female
	} else {
		return 0
	}
	switch faculty {
	case architecture:
		return priority.Architecture
	case engineering:
		return priority.Engineering
	case it:
		return priority.IT
	case medicine:
		return priority.Medicine
	case business:
		return priority.Business
	default:
		return 0
	}
}
