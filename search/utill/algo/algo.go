package algo

import (
	edlib "github.com/hbollon/go-edlib"

	"github.com/NamalSanjaya/nexster/pkgs/utill/math"
)

func JaccardIndexBasedScore(str1, str2 string) float32 {
	// 0 - whitespace spliting, 2 - two letters at time
	return math.Max[float32](edlib.JaccardSimilarity(str1, str2, 0), edlib.JaccardSimilarity(str1, str2, 2))
}
