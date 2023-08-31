package faculty

import "context"

const FacultyColl string = "faculty"
const (
	engineering  string = "engineering"
	it           string = "it"
	medicine     string = "medicine"
	business     string = "business"
	architecture string = "architecture"
)
const (
	male   string = "male"
	female string = "female"
)

type FacultyPriority struct {
	Engineering  int `json:"engineering"`
	IT           int `json:"it"`
	Medicine     int `json:"medicine"`
	Business     int `json:"business"`
	Architecture int `json:"architecture"`
}

type Interface interface {
	MkFacultyDocId(key string) string
	GetFaculty(ctx context.Context, key, gender string) (*FacultyWithGender, error)
	GetPriority(faculty, preferGender string, self *FacultyWithGender) int
}

type FacultyWithGender struct {
	Male   FacultyPriority `json:"male"`
	Female FacultyPriority `json:"female"`
}

type Faculty struct {
	Male   FacultyWithGender `json:"male"`
	Female FacultyWithGender `json:"female"`
}
