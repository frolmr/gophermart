package formatter

import (
	"math"

	"github.com/frolmr/gophermart/internal/domain"
)

func ConvertToCurrency(value int) float64 {
	return float64(value) / domain.ToSubunitDelimeter
}

func ConvertToSubunit(value float64) int {
	return int(math.Round(value * domain.ToSubunitDelimeter))
}
