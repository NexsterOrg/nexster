package interestgroups

import "fmt"

func MkInterestGroupDocId(key string) string {
	return fmt.Sprintf("%s/%s", interestGroupColl, key)
}
