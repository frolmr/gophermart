package formatter

import (
	"math"
	"strconv"

	"github.com/frolmr/gophermart/internal/domain"
)

func ConvertToCurrency(value int64) float64 {
	return float64(value) / domain.ToSubunitDelimeter
}

func ConvertToSubunit(value float64) int {
	return int(math.Round(value * domain.ToSubunitDelimeter))
}

func StringToInt64(stringValue string) (int64, error) {
	if value, err := strconv.ParseInt(stringValue, 10, 64); err != nil {
		return 0, err
	} else {
		return value, nil
	}
}

func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}
