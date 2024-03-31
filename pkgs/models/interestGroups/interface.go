package interestgroups

const interestGroupColl string = "interestGroups"

type InterestGroup struct {
	Key  string `json:"_key"`
	Name string `json:"name"`
	Type string `json:"type"`
}
